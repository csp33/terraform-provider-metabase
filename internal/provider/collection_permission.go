// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/csp33/terraform-provider-metabase/sdk/metabase"
	"github.com/csp33/terraform-provider-metabase/sdk/metabase/models/terraform"
	"github.com/csp33/terraform-provider-metabase/sdk/metabase/repositories"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func NewCollectionPermission() resource.Resource {
	collectionPermission := &CollectionPermission{}

	baseResource := &BaseResource{
		TypeName: "collection_permission",
		ConfigureRepository: func(client *metabase.MetabaseAPIClient) {
			collectionPermission.repository = repositories.NewCollectionPermissionRepository(client)
		},
		GetSchema: func(ctx context.Context) schema.Schema {
			return schema.Schema{
				MarkdownDescription: "Permission of one permission group on one collection (an edge of the Metabase collection graph). Removing the resource revokes access (sets it to \"none\").",
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
						MarkdownDescription: "ID of the collection",
						Required:            true,
						PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
					},
					"permission": schema.StringAttribute{
						MarkdownDescription: "Access level: \"read\" or \"write\".",
						Required:            true,
						Validators:          []validator.String{OneOfValidator("read", "write")},
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

			err := collectionPermission.repository.Set(ctx, plan.GroupId.ValueString(), plan.CollectionId.ValueString(), plan.Permission.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("Create Error", fmt.Sprintf("Unable to set collection permission: %s", err))
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
				resp.State.RemoveResource(ctx)
				return
			}

			result := terraform.CollectionPermissionTerraformModel{
				Id:           idOf(groupId, collectionId),
				GroupId:      stringValue(groupId),
				CollectionId: stringValue(collectionId),
				Permission:   stringValue(permission),
			}
			resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
		},
		UpdateFunc: func(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
			var plan terraform.CollectionPermissionTerraformModel
			resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
			if resp.Diagnostics.HasError() {
				return
			}

			err := collectionPermission.repository.Set(ctx, plan.GroupId.ValueString(), plan.CollectionId.ValueString(), plan.Permission.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("Update Error", fmt.Sprintf("Unable to update collection permission: %s", err))
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

			err := collectionPermission.repository.Set(ctx, state.GroupId.ValueString(), state.CollectionId.ValueString(), "none")
			if err != nil {
				resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("Unable to revoke collection permission: %s", err))
				return
			}
		},
	}

	collectionPermission.BaseResource = baseResource

	return collectionPermission
}

// CollectionPermission defines the resource implementation.
type CollectionPermission struct {
	*BaseResource
	repository *repositories.CollectionPermissionRepository
}
