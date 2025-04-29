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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func NewUser() resource.Resource {
	user := &User{}

	baseResource := &BaseResource{
		TypeName: "user",
		ConfigureRepository: func(client *metabase.MetabaseAPIClient) {
			user.repository = repositories.NewUserRepository(client)
		},
		GetSchema: func(ctx context.Context) schema.Schema {
			return schema.Schema{
				MarkdownDescription: "A user represents an individual with access to the Metabase instance. Each user has their own account and can be assigned to one or more permission groups, which determine their level of access to data, features, and administrative functions within Metabase.",
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "User ID",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"email": schema.StringAttribute{
						MarkdownDescription: "Email of the user",
						Required:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"first_name": schema.StringAttribute{
						MarkdownDescription: "First name of the user",
						Required:            true,
					},
					"last_name": schema.StringAttribute{
						MarkdownDescription: "Last name of the user",
						Required:            true,
					},
					"is_active": schema.BoolAttribute{
						MarkdownDescription: "Whether the user is active. Users can be deactivated, but not deleted.",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(true),
					},
				},
			}
		},
		CreateFunc: func(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
			var data terraform.UserTerraformModel

			resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

			if resp.Diagnostics.HasError() {
				return
			}

			if !data.IsActive.ValueBool() {
				resp.Diagnostics.AddError("Invalid value", "is_active must be true when creating a user")
				return
			}

			createResponse, err := user.repository.Create(ctx, data.Email.ValueString(), data.FirstName.ValueString(), data.LastName.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("Create Error", fmt.Sprintf("Unable to create User: %s", err))
				return
			}

			result := terraform.CreateUserTerraformModelFromDTO(createResponse)

			resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
		},
		ReadFunc: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
			var data terraform.UserTerraformModel
			resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

			if resp.Diagnostics.HasError() {
				return
			}

			getResponse, err := user.repository.Get(ctx, data.Id.ValueString())

			if err != nil {
				resp.Diagnostics.AddError("Get Error", fmt.Sprintf("Unable to get user: %s", err))
				return
			}
			result := terraform.CreateUserTerraformModelFromDTO(getResponse)

			resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
		},
		UpdateFunc: func(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
			var plan terraform.UserTerraformModel
			var state terraform.UserTerraformModel
			resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
			resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

			var isActive *bool
			if plan.IsActive.ValueBool() == state.IsActive.ValueBool() {
				isActive = nil
			} else {
				isActive = plan.IsActive.ValueBoolPointer()
			}

			if resp.Diagnostics.HasError() {
				return
			}

			_, err := user.repository.Update(ctx, plan.Id.ValueString(), plan.FirstName.ValueStringPointer(), plan.LastName.ValueStringPointer(), isActive)
			if err != nil {
				resp.Diagnostics.AddError("Update Error", fmt.Sprintf("Unable to update user: %s", err))
				return
			}

			// The new state is not read from the API

			resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		},
		DeleteFunc: func(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
			resp.Diagnostics.AddError("Can't delete", "Users can't be deleted, set is_active=false instead")
		},
	}

	user.BaseResource = baseResource

	return user
}

// User defines the resource implementation.
type User struct {
	*BaseResource
	repository *repositories.UserRepository
}
