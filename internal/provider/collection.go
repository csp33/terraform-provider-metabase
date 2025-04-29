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

func NewCollection() resource.Resource {
	collection := &Collection{}

	baseResource := &BaseResource{
		TypeName: "collection",
		ConfigureRepository: func(client *metabase.MetabaseAPIClient) {
			collection.repository = repositories.NewCollectionRepository(client)
		},
		GetSchema: func(ctx context.Context) schema.Schema {
			return schema.Schema{
				MarkdownDescription: "In Metabase, a collection is a set of items — questions, models, dashboards, and subcollections — that are stored together for some organizational purpose. You can think of collections like folders within a file system.",
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Collection ID",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"name": schema.StringAttribute{
						MarkdownDescription: "Name of the collection",
						Required:            true,
					},
					"parent_id": schema.StringAttribute{
						MarkdownDescription: "ID of the parent collection",
						Optional:            true,
					},
					"archived": schema.BoolAttribute{
						MarkdownDescription: "Whether the collection is archived. Archived collections are not visible in the UI. Collections can be archived, but not deleted.",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
				},
			}
		},
		CreateFunc: func(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
			var plan terraform.CollectionTerraformModel

			resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

			if resp.Diagnostics.HasError() {
				return
			}
			if plan.Archived.ValueBool() {
				resp.Diagnostics.AddError("Invalid value", "A collection can't be created with archived=true")
				return
			}

			createResponse, err := collection.repository.Create(ctx, plan.Name.ValueString(), plan.ParentId.ValueStringPointer(), plan.Archived.ValueBoolPointer())
			if err != nil {
				resp.Diagnostics.AddError("Create Error", fmt.Sprintf("Unable to create collection: %s", err))
				return
			}

			result := terraform.CreateCollectionTerraformModelFromDTO(createResponse)
			resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
		},
		ReadFunc: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
			var plan terraform.CollectionTerraformModel
			resp.Diagnostics.Append(req.State.Get(ctx, &plan)...)

			if resp.Diagnostics.HasError() {
				return
			}

			getResponse, err := collection.repository.Get(ctx, plan.Id.ValueString())

			if err != nil {
				resp.Diagnostics.AddError("Get Error", fmt.Sprintf("Unable to get collection: %s", err))
				return
			}
			result := terraform.CreateCollectionTerraformModelFromDTO(getResponse)

			resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
		},
		UpdateFunc: func(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
			var plan terraform.CollectionTerraformModel
			resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

			if resp.Diagnostics.HasError() {
				return
			}

			_, err := collection.repository.Update(ctx, plan.Id.ValueString(), plan.Name.ValueStringPointer(), plan.ParentId.ValueStringPointer(), plan.Archived.ValueBoolPointer())
			if err != nil {
				resp.Diagnostics.AddError("Update Error", fmt.Sprintf("Unable to update collection: %s", err))
				return
			}

			// The new state is not read from the API

			resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		},
		DeleteFunc: func(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
			var state terraform.CollectionTerraformModel
			resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

			if resp.Diagnostics.HasError() {
				return
			}

			archived := true
			_, err := collection.repository.Update(ctx, state.Id.ValueString(), nil, nil, &archived)
			if err != nil {
				resp.Diagnostics.AddError("Archive Error", fmt.Sprintf("Unable to archive collection: %s", err))
				return
			}
		},
	}

	collection.BaseResource = baseResource

	return collection
}

// Collection defines the resource implementation.
type Collection struct {
	*BaseResource
	repository *repositories.CollectionRepository
}
