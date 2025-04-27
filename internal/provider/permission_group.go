// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/csp33/terraform-provider-metabase/sdk/metabase"
	"github.com/csp33/terraform-provider-metabase/sdk/metabase/models"
	"github.com/csp33/terraform-provider-metabase/sdk/metabase/repositories"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &PermissionGroup{}
var _ resource.ResourceWithImportState = &PermissionGroup{}

func NewPermissionGroup() resource.Resource {
	return &PermissionGroup{}
}

// PermissionGroup defines the resource implementation.
type PermissionGroup struct {
	repository *repositories.PermissionGroupRepository
}

// PermissionGroupModel describes the resource data model.

func (r *PermissionGroup) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_permission_group"
}

func (r *PermissionGroup) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Permission Group",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the group",
				Required:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Group ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *PermissionGroup) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.repository = repositories.NewPermissionGroupRepository(client)
}

func (r *PermissionGroup) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data models.PermissionGroup

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	createResponse, err := r.repository.Create(ctx, data.Name)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("Unable to create permission group: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &createResponse)...)
}

func (r *PermissionGroup) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data models.PermissionGroup
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	getResponse, err := r.repository.Get(ctx, data.Id)

	if err != nil {
		resp.Diagnostics.AddError("Get Error", fmt.Sprintf("Unable to get permission group: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &getResponse)...)
}

func (r *PermissionGroup) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data models.PermissionGroup
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	_, err := r.repository.Update(ctx, data.Id, data.Name)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", fmt.Sprintf("Unable to update permission group: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PermissionGroup) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data models.PermissionGroup
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	err := r.repository.Delete(ctx, data.Id)
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("Unable to delete permission group: %s", err))
		return
	}
}

func (r *PermissionGroup) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
