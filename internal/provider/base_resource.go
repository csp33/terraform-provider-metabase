// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/csp33/terraform-provider-metabase/sdk/metabase"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Ensure BaseResource implements required interfaces.
var _ resource.Resource = &BaseResource{}
var _ resource.ResourceWithImportState = &BaseResource{}

// BaseResource provides common functionality for all resources.
type BaseResource struct {
	// TypeName is the name of the resource type (e.g., "metabase_collection")
	TypeName string

	// ConfigureRepository is a function that configures the repository for the resource
	ConfigureRepository func(client *metabase.MetabaseAPIClient)

	// GetSchema returns the schema for the resource
	GetSchema func(ctx context.Context) schema.Schema

	// CreateFunc creates a new resource
	CreateFunc func(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse)

	// ReadFunc reads a resource
	ReadFunc func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse)

	// UpdateFunc updates a resource
	UpdateFunc func(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse)

	// DeleteFunc deletes a resource
	DeleteFunc func(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse)
}

// Metadata implements resource.Resource.
func (r *BaseResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.TypeName
}

// Schema implements resource.Resource.
func (r *BaseResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = r.GetSchema(ctx)
}

// Configure implements resource.Resource.
func (r *BaseResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.ConfigureRepository(client)
}

// Create implements resource.Resource.
func (r *BaseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	r.CreateFunc(ctx, req, resp)
}

// Read implements resource.Resource.
func (r *BaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	r.ReadFunc(ctx, req, resp)
}

// Update implements resource.Resource.
func (r *BaseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.UpdateFunc(ctx, req, resp)
}

// Delete implements resource.Resource.
func (r *BaseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	r.DeleteFunc(ctx, req, resp)
}

// ImportState implements resource.ResourceWithImportState.
func (r *BaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
