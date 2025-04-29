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

func NewUserPermissionGroupMembership() resource.Resource {
	membership := &UserPermissionGroupMembership{}

	baseResource := &BaseResource{
		TypeName: "user_permission_group_membership",
		ConfigureRepository: func(client *metabase.MetabaseAPIClient) {
			membership.repository = repositories.NewUserPermissionGroupMembershipRepository(client)
		},
		GetSchema: func(ctx context.Context) schema.Schema {
			return schema.Schema{
				// This description is used by the documentation generator and the language server.
				MarkdownDescription: "Represents the link between a specific user and a permission group. Users can be members of multiple groups, and their effective permissions are the union of all permissions granted to the groups they belong to.",

				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Membership ID",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"user_id": schema.StringAttribute{
						MarkdownDescription: "ID of the user",
						Required:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"permission_group_id": schema.StringAttribute{
						MarkdownDescription: "ID of the permission group",
						Required:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
				},
			}
		},
		CreateFunc: func(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
			var plan terraform.UserPermissionGroupMembershipTerraformModel

			resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

			if resp.Diagnostics.HasError() {
				return
			}

			createResponse, err := membership.repository.Create(ctx, plan.UserId.ValueString(), plan.PermissionGroupId.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("Create Error", fmt.Sprintf("Unable to create user permission group membership membership: %s", err))
				return
			}

			result := terraform.CreateUserPermissionGroupMembershipTerraformModelFromDTO(createResponse)

			resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
		},
		ReadFunc: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
			var plan terraform.UserPermissionGroupMembershipTerraformModel
			resp.Diagnostics.Append(req.State.Get(ctx, &plan)...)

			if resp.Diagnostics.HasError() {
				return
			}

			getResponse, err := membership.repository.Get(ctx, plan.Id.ValueString())

			if err != nil {
				resp.Diagnostics.AddError("Get Error", fmt.Sprintf("Unable to get user permission group membership: %s", err))
				return
			}
			result := terraform.CreateUserPermissionGroupMembershipTerraformModelFromDTO(getResponse)

			resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
		},
		UpdateFunc: func(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
			resp.Diagnostics.AddError("Can't update", "User permission group membership can't be updated")
		},
		DeleteFunc: func(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
			var plan terraform.UserPermissionGroupMembershipTerraformModel
			resp.Diagnostics.Append(req.State.Get(ctx, &plan)...)

			if resp.Diagnostics.HasError() {
				return
			}

			err := membership.repository.Delete(ctx, plan.Id.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("Unable to delete user permission group membership: %s", err))
				return
			}
		},
	}

	membership.BaseResource = baseResource

	return membership
}

// UserPermissionGroupMembership defines the resource implementation.
type UserPermissionGroupMembership struct {
	*BaseResource
	repository *repositories.UserPermissionGroupMembershipRepository
}
