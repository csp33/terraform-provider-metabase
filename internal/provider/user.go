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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &User{}
var _ resource.ResourceWithImportState = &User{}

func NewUser() resource.Resource {
	return &User{}
}

// User defines the resource implementation.
type User struct {
	repository *repositories.UserRepository
}

func (r *User) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *User) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "User",

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
			"is_active": schema.BoolAttribute{
				MarkdownDescription: "Whether the user is active",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
		},
	}
}

func (r *User) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.repository = repositories.NewUserRepository(client)
}

func (r *User) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data terraform.UserTerraformModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if !data.IsActive.ValueBool() {
		resp.Diagnostics.AddError("Invalid value", "is_active must be true when creating a user")
		return
	}

	createResponse, err := r.repository.Create(ctx, data.Email.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("Unable to create User: %s", err))
		return
	}

	result := terraform.CreateUserTerraformModelFromDTO(createResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *User) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data terraform.UserTerraformModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	getResponse, err := r.repository.Get(ctx, data.Id.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Get Error", fmt.Sprintf("Unable to get user: %s", err))
		return
	}
	result := terraform.CreateUserTerraformModelFromDTO(getResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *User) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan terraform.UserTerraformModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.repository.Update(ctx, plan.Id.ValueString(), plan.IsActive.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("Update Error", fmt.Sprintf("Unable to update user: %s", err))
		return
	}

	// The new state is not read from the API

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *User) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddError("Can't delete", "Users can't be deleted, set is_active=false instead")
}

func (r *User) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
