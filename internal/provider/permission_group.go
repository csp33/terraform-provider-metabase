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
)

func NewPermissionGroup() resource.Resource {
	permissionGroup := &PermissionGroup{}

	baseResource := &BaseResource{
		TypeName: "permission_group",
		ConfigureRepository: func(client *metabase.MetabaseAPIClient) {
			permissionGroup.repository = repositories.NewPermissionGroupRepository(client)
		},
		GetSchema: func(ctx context.Context) schema.Schema {
			return schema.Schema{
				MarkdownDescription: "A group to which users can be added for simplified permission management. A user can belong to multiple groups.",
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "ID of the group",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"name": schema.StringAttribute{
						MarkdownDescription: "Name of the group",
						Required:            true,
					},
				},
			}
		},
		CreateFunc: func(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
			var plan terraform.PermissionGroupTerraformModel

			resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

			if resp.Diagnostics.HasError() {
				return
			}

			createResponse, err := permissionGroup.repository.Create(ctx, plan.Name.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("Create Error", fmt.Sprintf("Unable to create permission group: %s", err))
				return
			}

			result := terraform.CreatePermissionGroupTerraformModelFromDTO(createResponse)

			resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
		},
		ReadFunc: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
			var plan terraform.PermissionGroupTerraformModel
			resp.Diagnostics.Append(req.State.Get(ctx, &plan)...)

			if resp.Diagnostics.HasError() {
				return
			}

			getResponse, err := permissionGroup.repository.Get(ctx, plan.Id.ValueString())

			if err != nil {
				resp.Diagnostics.AddError("Get Error", fmt.Sprintf("Unable to get permission group: %s", err))
				return
			}
			result := terraform.CreatePermissionGroupTerraformModelFromDTO(getResponse)

			resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
		},
		UpdateFunc: func(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
			var plan terraform.PermissionGroupTerraformModel
			resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

			if resp.Diagnostics.HasError() {
				return
			}

			_, err := permissionGroup.repository.Update(ctx, plan.Id.ValueString(), plan.Name.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("Update Error", fmt.Sprintf("Unable to update permission group: %s", err))
				return
			}

			resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		},
		DeleteFunc: func(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
			var plan terraform.PermissionGroupTerraformModel
			resp.Diagnostics.Append(req.State.Get(ctx, &plan)...)

			if resp.Diagnostics.HasError() {
				return
			}

			err := permissionGroup.repository.Delete(ctx, plan.Id.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("Unable to delete permission group: %s", err))
				return
			}
		},
	}

	permissionGroup.BaseResource = baseResource

	return permissionGroup
}

// PermissionGroup defines the resource implementation.
type PermissionGroup struct {
	*BaseResource
	repository *repositories.PermissionGroupRepository
}
