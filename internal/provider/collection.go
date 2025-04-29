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
var _ resource.Resource = &Collection{}
var _ resource.ResourceWithImportState = &Collection{}

func NewCollection() resource.Resource {
	return &Collection{}
}

// Collection defines the resource implementation.
type Collection struct {
	repository *repositories.CollectionRepository
}

// CollectionModel describes the resource data model.

func (r *Collection) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_collection"
}

func (r *Collection) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "collection",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the collection",
				Required:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Collection ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"parent_id": schema.StringAttribute{
				MarkdownDescription: "ID of the parent collection",
				Optional:            true,
			},
		},
	}
}

func (r *Collection) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.repository = repositories.NewCollectionRepository(client)
}

func (r *Collection) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data terraform.CollectionTerraformModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	createResponse, err := r.repository.Create(ctx, data.Name.ValueString(), data.ParentId.ValueStringPointer())
	if err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("Unable to create collection: %s", err))
		return
	}

	result := terraform.CreateCollectionTerraformModelFromDTO(createResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *Collection) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data terraform.CollectionTerraformModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	getResponse, err := r.repository.Get(ctx, data.Id.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Get Error", fmt.Sprintf("Unable to get collection: %s", err))
		return
	}
	result := terraform.CreateCollectionTerraformModelFromDTO(getResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *Collection) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data terraform.CollectionTerraformModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.repository.Update(ctx, data.Id.ValueString(), data.Name.ValueStringPointer(), data.ParentId.ValueStringPointer(), data.Archived.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("Update Error", fmt.Sprintf("Unable to update collection: %s", err))
		return
	}

	// The new state is not read from the API

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Collection) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddError("Can't delete", "Collections can't be updated, set archived=true instead")
}

func (r *Collection) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
