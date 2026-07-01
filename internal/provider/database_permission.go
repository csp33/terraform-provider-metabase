// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/csp33/terraform-provider-metabase/sdk/metabase"
	"github.com/csp33/terraform-provider-metabase/sdk/metabase/models/terraform"
	"github.com/csp33/terraform-provider-metabase/sdk/metabase/repositories"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// splitEdgeID splits a "left:right" composite id.
func splitEdgeID(id string) (string, string, error) {
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid id %q, expected \"<left>:<right>\"", id)
	}
	return parts[0], parts[1], nil
}

func NewDatabasePermission() resource.Resource {
	databasePermission := &DatabasePermission{}

	baseResource := &BaseResource{
		TypeName: "database_permission",
		ConfigureRepository: func(client *metabase.MetabaseAPIClient) {
			databasePermission.repository = repositories.NewDatabasePermissionRepository(client)
		},
		GetSchema: func(ctx context.Context) schema.Schema {
			return schema.Schema{
				MarkdownDescription: "Query-building access of one permission group on one database (an edge of the Metabase permissions graph). On OSS this is the real data-access control (view-data is always unrestricted and is not managed).",
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Composite id \"<group_id>:<database_id>\"",
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"group_id": schema.StringAttribute{
						MarkdownDescription: "ID of the permission group",
						Required:            true,
						PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
					},
					"database_id": schema.StringAttribute{
						MarkdownDescription: "ID of the database",
						Required:            true,
						PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
					},
					"create_queries": schema.StringAttribute{
						MarkdownDescription: "Query-building access: \"no\", \"query-builder\", or \"query-builder-and-native\".",
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString("no"),
						Validators:          []validator.String{OneOfValidator("no", "query-builder", "query-builder-and-native")},
					},
				},
			}
		},
		CreateFunc: func(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
			var plan terraform.DatabasePermissionTerraformModel
			resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
			if resp.Diagnostics.HasError() {
				return
			}

			err := databasePermission.repository.Set(ctx, plan.GroupId.ValueString(), plan.DatabaseId.ValueString(), plan.CreateQueries.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("Create Error", fmt.Sprintf("Unable to set database permission: %s", err))
				return
			}

			plan.Id = idOf(plan.GroupId.ValueString(), plan.DatabaseId.ValueString())
			resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		},
		ReadFunc: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
			var state terraform.DatabasePermissionTerraformModel
			resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
			if resp.Diagnostics.HasError() {
				return
			}

			groupId, databaseId, err := splitEdgeID(state.Id.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("Read Error", err.Error())
				return
			}

			createQueries, found, err := databasePermission.repository.Get(ctx, groupId, databaseId)
			if err != nil {
				resp.Diagnostics.AddError("Get Error", fmt.Sprintf("Unable to get database permission: %s", err))
				return
			}
			if !found {
				resp.State.RemoveResource(ctx)
				return
			}

			result := terraform.DatabasePermissionTerraformModel{
				Id:            idOf(groupId, databaseId),
				GroupId:       stringValue(groupId),
				DatabaseId:    stringValue(databaseId),
				CreateQueries: stringValue(createQueries),
			}
			resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
		},
		UpdateFunc: func(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
			var plan terraform.DatabasePermissionTerraformModel
			resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
			if resp.Diagnostics.HasError() {
				return
			}

			err := databasePermission.repository.Set(ctx, plan.GroupId.ValueString(), plan.DatabaseId.ValueString(), plan.CreateQueries.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("Update Error", fmt.Sprintf("Unable to update database permission: %s", err))
				return
			}

			resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		},
		DeleteFunc: func(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
			var state terraform.DatabasePermissionTerraformModel
			resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
			if resp.Diagnostics.HasError() {
				return
			}

			// Revoke = no query access. The edge itself can't be removed (Metabase
			// keeps a view-data entry per group/db) and null would be rejected.
			err := databasePermission.repository.Set(ctx, state.GroupId.ValueString(), state.DatabaseId.ValueString(), "no")
			if err != nil {
				resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("Unable to revoke database permission: %s", err))
				return
			}
		},
	}

	databasePermission.BaseResource = baseResource

	return databasePermission
}

// DatabasePermission defines the resource implementation.
type DatabasePermission struct {
	*BaseResource
	repository *repositories.DatabasePermissionRepository
}
