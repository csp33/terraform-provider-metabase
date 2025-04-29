// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/csp33/terraform-provider-metabase/sdk/metabase"
	"github.com/csp33/terraform-provider-metabase/sdk/metabase/models/terraform"
	"github.com/csp33/terraform-provider-metabase/sdk/metabase/repositories"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &UserPermissionGroupMembership{}
var _ resource.ResourceWithImportState = &UserPermissionGroupMembership{}

func NewUserPermissionGroupMembership() resource.Resource {
	return &UserPermissionGroupMembership{}
}

// UserPermissionGroupMembership defines the resource implementation.
type UserPermissionGroupMembership struct {
	repository *repositories.UserPermissionGroupMembershipRepository
}

func (r *UserPermissionGroupMembership) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_permission_group_membership"
}

func (r *UserPermissionGroupMembership) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "User Permission group membership",

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
				MarkdownDescription: "ID of the user permission group membership",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *UserPermissionGroupMembership) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*metabase.MetabaseAPIClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *metabase.MetabaseAPIClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.repository = repositories.NewUserPermissionGroupMembershipRepository(client)
}

func (r *UserPermissionGroupMembership) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data terraform.UserPermissionGroupMembershipTerraformModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	createResponse, err := r.repository.Create(ctx, data.UserId.ValueString(), data.PermissionGroupId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("Unable to create user permission group membership membership: %s", err))
		return
	}

	result := terraform.CreateUserPermissionGroupMembershipTerraformModelFromDTO(createResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *UserPermissionGroupMembership) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data terraform.UserPermissionGroupMembershipTerraformModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	getResponse, err := r.repository.Get(ctx, data.Id.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Get Error", fmt.Sprintf("Unable to get user permission group membership: %s", err))
		return
	}
	result := terraform.CreateUserPermissionGroupMembershipTerraformModelFromDTO(getResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *UserPermissionGroupMembership) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Can't update", "User permission group membership can't be updated")
}

func (r *UserPermissionGroupMembership) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data terraform.UserPermissionGroupMembershipTerraformModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.repository.Delete(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("Unable to delete user permission group membership: %s", err))
		return
	}
}

func (r *UserPermissionGroupMembership) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
