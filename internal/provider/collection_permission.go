// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/csp33/terraform-provider-metabase/sdk/metabase"
	"github.com/csp33/terraform-provider-metabase/sdk/metabase/models/terraform"
	"github.com/csp33/terraform-provider-metabase/sdk/metabase/repositories"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Propagation & edge-case semantics (mirrors the UI "Also change sub-collections"):
//   - propagate=true: create/update write the SAME permission on the collection
//     and every current descendant in ONE graph PUT; delete revokes the whole
//     subtree. Only the ROOT edge is tracked in state — descendants are a
//     side effect, so sub-collections stay UI-manageable without drift, and a
//     descendant created later inherits the parent's perms natively anyway.
//   - propagate true->false never revokes previously propagated edges (the
//     provider doesn't own them); delete the resource to revoke.
//   - personal collections: never touched (filtered out of the expansion).
//   - archived (trashed) collections: SKIPPED on create/update (restore must
//     recover a collection's own permissions untouched; Metabase hides them
//     from the graph but silently accepts writes) and INCLUDED on delete (no
//     access may survive in the Trash after a revoke). An archived ROOT
//     freezes the resource (clean plan) instead of looping create/apply, and
//     creating/updating onto an archived root fails loudly.
//   - hard-deleted (Trash emptied) ROOT: Read drops it from state and Create
//     fails with an explicit error. This resource only manages the edge and
//     never creates collections — the operator either recreates the collection
//     (metabase_collection can manage them) or removes the grant.
//   - collection_id "root" (the virtual "Our Analytics" root): propagation
//     expands to EVERY non-personal collection, and Metabase copies root's
//     permissions to any collection created at any level afterwards — the
//     natural shape for groups that see/curate everything. Root cannot be
//     archived or deleted, so the existence/archived checks are skipped.
func NewCollectionPermission() resource.Resource {
	collectionPermission := &CollectionPermission{}

	baseResource := &BaseResource{
		TypeName: "collection_permission",
		ConfigureRepository: func(client *metabase.MetabaseAPIClient) {
			collectionPermission.repository = repositories.NewCollectionPermissionRepository(client)
			collectionPermission.collections = repositories.NewCollectionRepository(client)
		},
		GetSchema: func(ctx context.Context) schema.Schema {
			return schema.Schema{
				MarkdownDescription: "Permission of one permission group on one collection (an edge of the Metabase collection graph). Removing the resource revokes access (sets it to \"none\").\n\n" +
					"With `propagate = true` the permission is also applied to every descendant sub-collection in a single graph update, like the UI's \"Also change sub-collections\" toggle (Metabase does not compute inheritance — a sub-collection only copies its parent's permissions when it is created). Only this edge is tracked in state: descendants are a side effect, so they never drift, and collections moved in or out of the subtree later are not re-reconciled (re-trigger with `terraform apply -replace=...`).\n\n" +
					"Edge cases: personal collections are never touched. Archived (trashed) collections are skipped on create/update — restoring one recovers its own permissions untouched — and included on delete, so no access survives in the Trash; if the target collection itself is archived the resource freezes (clean plan) and creating onto it fails explicitly. If the collection was permanently deleted (Trash emptied), recreate it (e.g. with `metabase_collection`) or remove the grant.\n\n" +
					"A grant on `collection_id = \"root\"` (the virtual \"Our Analytics\" collection) with `propagate = true` expands to EVERY non-personal collection, and Metabase copies root's permissions to collections created at any level afterwards — the natural shape for groups that must see or curate everything. Note that write access on root is also what allows creating new top-level collections.",
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Composite id \"<group_id>:<collection_id>\"",
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"group_id": schema.StringAttribute{
						MarkdownDescription: "ID of the permission group",
						Required:            true,
						PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
					},
					"collection_id": schema.StringAttribute{
						MarkdownDescription: "ID of the collection, or `\"root\"` for the virtual root collection (\"Our Analytics\").",
						Required:            true,
						PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
					},
					"permission": schema.StringAttribute{
						MarkdownDescription: "Access level: \"read\" or \"write\" (to revoke, delete the resource; there is no \"none\" value).",
						Required:            true,
						Validators:          []validator.String{OneOfValidator("read", "write")},
					},
					"propagate": schema.BoolAttribute{
						MarkdownDescription: "Also apply the permission to every descendant sub-collection (and revoke the whole subtree on delete). Only this edge is tracked in state: descendants never drift, and collections moved in/out of the subtree later are not re-reconciled (re-trigger with `-replace` if needed). Archived descendants are skipped except on delete; personal collections are never touched. Setting it back to false does NOT revoke already-propagated edges.",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
				},
			}
		},
		CreateFunc: func(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
			var plan terraform.CollectionPermissionTerraformModel
			resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
			if resp.Diagnostics.HasError() {
				return
			}

			if err := collectionPermission.apply(ctx, plan); err != nil {
				resp.Diagnostics.AddError("Create Error", err.Error())
				return
			}

			plan.Id = idOf(plan.GroupId.ValueString(), plan.CollectionId.ValueString())
			resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		},
		ReadFunc: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
			var state terraform.CollectionPermissionTerraformModel
			resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
			if resp.Diagnostics.HasError() {
				return
			}

			groupId, collectionId, err := splitEdgeID(state.Id.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("Read Error", err.Error())
				return
			}

			permission, found, err := collectionPermission.repository.Get(ctx, groupId, collectionId)
			if err != nil {
				resp.Diagnostics.AddError("Get Error", fmt.Sprintf("Unable to get collection permission: %s", err))
				return
			}
			if !found {
				if collectionId == "root" { // root always exists and never archives
					resp.State.RemoveResource(ctx)
					return
				}
				// Absent from the graph: revoked out-of-band, archived or gone.
				col, cerr := collectionPermission.collections.Get(ctx, collectionId)
				var notFound *metabase.NotFoundError
				switch {
				case errors.As(cerr, &notFound):
					resp.State.RemoveResource(ctx) // hard-deleted: dead grant
				case cerr != nil:
					resp.Diagnostics.AddError("Get Error", fmt.Sprintf("Unable to check collection %s: %s", collectionId, cerr))
				case col.Archived:
					// Trashed: perms persist but are hidden — freeze (keep state).
				default:
					resp.State.RemoveResource(ctx) // edge truly revoked
				}
				return
			}

			result := terraform.CollectionPermissionTerraformModel{
				Id:           idOf(groupId, collectionId),
				GroupId:      stringValue(groupId),
				CollectionId: stringValue(collectionId),
				Permission:   stringValue(permission),
				Propagate:    types.BoolValue(state.Propagate.ValueBool()),
			}
			resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
		},
		UpdateFunc: func(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
			var plan terraform.CollectionPermissionTerraformModel
			resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
			if resp.Diagnostics.HasError() {
				return
			}

			if err := collectionPermission.apply(ctx, plan); err != nil {
				resp.Diagnostics.AddError("Update Error", err.Error())
				return
			}

			resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		},
		DeleteFunc: func(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
			var state terraform.CollectionPermissionTerraformModel
			resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
			if resp.Diagnostics.HasError() {
				return
			}
			collectionId := state.CollectionId.ValueString()

			if collectionId != "root" { // root always exists
				_, err := collectionPermission.collections.Get(ctx, collectionId)
				var notFound *metabase.NotFoundError
				if errors.As(err, &notFound) {
					return // hard-deleted: its permission rows are gone with it
				}
			}

			ids := []string{collectionId}
			if state.Propagate.ValueBool() {
				// Include archived descendants: no access may survive in the Trash.
				descendants, err := collectionPermission.repository.ListDescendants(ctx, collectionId, true)
				if err != nil {
					resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("Unable to list sub-collections of %s: %s", collectionId, err))
					return
				}
				ids = append(ids, descendants...)
			}

			if err := collectionPermission.repository.SetMany(ctx, state.GroupId.ValueString(), ids, "none"); err != nil {
				resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("Unable to revoke collection permission: %s", err))
			}
		},
	}

	collectionPermission.BaseResource = baseResource

	return collectionPermission
}

// apply validates the target collection and sets the permission, expanding to
// active descendants when propagate is on (shared by Create and Update).
func (c *CollectionPermission) apply(ctx context.Context, plan terraform.CollectionPermissionTerraformModel) error {
	collectionId := plan.CollectionId.ValueString()

	if collectionId != "root" { // root always exists and never archives
		col, err := c.collections.Get(ctx, collectionId)
		var notFound *metabase.NotFoundError
		if errors.As(err, &notFound) {
			return fmt.Errorf("collection %s no longer exists (permanently deleted?) — recreate it (e.g. with metabase_collection) or remove this grant", collectionId)
		}
		if err != nil {
			return fmt.Errorf("unable to check collection %s: %s", collectionId, err)
		}
		if col.Archived {
			return fmt.Errorf("collection %s is archived — restore it or remove this grant (Metabase would accept the write silently without showing it)", collectionId)
		}
	}

	ids := []string{collectionId}
	if plan.Propagate.ValueBool() {
		descendants, err := c.repository.ListDescendants(ctx, collectionId, false)
		if err != nil {
			return fmt.Errorf("unable to list sub-collections of %s: %s", collectionId, err)
		}
		ids = append(ids, descendants...)
	}

	if err := c.repository.SetMany(ctx, plan.GroupId.ValueString(), ids, plan.Permission.ValueString()); err != nil {
		return fmt.Errorf("unable to set collection permission: %s", err)
	}
	return nil
}

// CollectionPermission defines the resource implementation.
type CollectionPermission struct {
	*BaseResource
	repository  *repositories.CollectionPermissionRepository
	collections *repositories.CollectionRepository
}
