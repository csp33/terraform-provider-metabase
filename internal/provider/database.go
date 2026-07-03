// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
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

func NewDatabase() resource.Resource {
	database := &Database{}

	baseResource := &BaseResource{
		TypeName: "database",
		ConfigureRepository: func(client *metabase.MetabaseAPIClient) {
			database.repository = repositories.NewDatabaseRepository(client)
		},
		GetSchema: func(ctx context.Context) schema.Schema {
			return schema.Schema{
				MarkdownDescription: "A database connection in Metabase.",
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Database ID",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"name": schema.StringAttribute{
						MarkdownDescription: "Name of the database",
						Required:            true,
					},
					"engine": schema.StringAttribute{
						MarkdownDescription: "Database engine (e.g. postgres, mysql, h2, bigquery-cloud-sdk)",
						Required:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					// Config is authoritative: Metabase redacts secrets, so details is never read back.
					"details": schema.StringAttribute{
						MarkdownDescription: "Connection details as a JSON object (use jsonencode). Config is authoritative; not read back from Metabase.",
						Required:            true,
						Sensitive:           true,
					},
					// Terraform-only guard: deleting a database hard-deletes all content built on it.
					"deletion_protection": schema.BoolAttribute{
						MarkdownDescription: "If true (default), refuses to delete the database. Metabase hard-deletes a database and all content built on it, so set this to false and apply before destroying.",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(true),
					},
				},
			}
		},
		CreateFunc: func(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
			var plan terraform.DatabaseTerraformModel
			resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
			if resp.Diagnostics.HasError() {
				return
			}

			createResponse, err := database.repository.Create(ctx, plan.Name.ValueString(), plan.Engine.ValueString(), plan.Details.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("Create Error", fmt.Sprintf("Unable to create database: %s", err))
				return
			}

			result := terraform.CreateDatabaseTerraformModelFromDTO(createResponse, plan.Details, plan.DeletionProtection)
			resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
		},
		ReadFunc: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
			var state terraform.DatabaseTerraformModel
			resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
			if resp.Diagnostics.HasError() {
				return
			}

			getResponse, err := database.repository.Get(ctx, state.Id.ValueString())
			if err != nil {
				var notFound *metabase.NotFoundError
				if errors.As(err, &notFound) {
					resp.State.RemoveResource(ctx)
					return
				}
				resp.Diagnostics.AddError("Get Error", fmt.Sprintf("Unable to get database: %s", err))
				return
			}

			result := terraform.CreateDatabaseTerraformModelFromDTO(getResponse, state.Details, state.DeletionProtection)
			resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
		},
		UpdateFunc: func(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
			var plan terraform.DatabaseTerraformModel
			resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
			if resp.Diagnostics.HasError() {
				return
			}

			_, err := database.repository.Update(ctx, plan.Id.ValueString(), plan.Name.ValueStringPointer(), plan.Details.ValueStringPointer())
			if err != nil {
				resp.Diagnostics.AddError("Update Error", fmt.Sprintf("Unable to update database: %s", err))
				return
			}

			resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		},
		DeleteFunc: func(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
			var state terraform.DatabaseTerraformModel
			resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
			if resp.Diagnostics.HasError() {
				return
			}

			if state.DeletionProtection.ValueBool() {
				resp.Diagnostics.AddError(
					"Database deletion protected",
					"This database has deletion_protection = true. Metabase hard-deletes a database and all content built on it. Set deletion_protection = false and apply before removing or destroying this resource.",
				)
				return
			}

			err := database.repository.Delete(ctx, state.Id.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("Unable to delete database: %s", err))
				return
			}
		},
	}

	database.BaseResource = baseResource

	return database
}

// Database defines the resource implementation.
type Database struct {
	*BaseResource
	repository *repositories.DatabaseRepository
}
