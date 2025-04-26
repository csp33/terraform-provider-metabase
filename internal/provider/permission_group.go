// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"net/http"
	"strconv"
	"strings"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &PermissionGroup{}
var _ resource.ResourceWithImportState = &PermissionGroup{}

func NewPermissionGroup() resource.Resource {
	return &PermissionGroup{}
}

// PermissionGroup defines the resource implementation.
type PermissionGroup struct {
	client *MetabaseClient
}

// PermissionGroupModel describes the resource data model.
type PermissionGroupModel struct {
	Name types.String `tfsdk:"name"`
	Id   types.String `tfsdk:"id"`
}

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

	client, ok := req.ProviderData.(*MetabaseClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *MetabaseClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *PermissionGroup) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PermissionGroupModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("%s/api/permissions/group", r.client.Host)
	body := fmt.Sprintf(`{"name":"%s"}`, data.Name.ValueString())

	request, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		resp.Diagnostics.AddError("Request Error", fmt.Sprintf("Unable to create request: %s", err))
		return
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("x-api-key", r.client.ApiKey)

	response, err := r.client.Client.Do(request)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create permission group: %s", err))
		return
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unexpected status code: %d", response.StatusCode))
		return
	}

	tflog.Trace(ctx, "Group successfully created")

	var respBody struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(response.Body).Decode(&respBody); err != nil {
		resp.Diagnostics.AddError("Decode Error", fmt.Sprintf("Unable to decode response: %s", err))
		return
	}

	data.Id = types.StringValue(strconv.Itoa(respBody.ID))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PermissionGroup) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PermissionGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("%s/api/permissions/group/%s", r.client.Host, data.Id.ValueString())

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Request Error", fmt.Sprintf("Unable to create request: %s", err))
		return
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("x-api-key", r.client.ApiKey)

	response, err := r.client.Client.Do(request)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read permission group: %s", err))
		return
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNotFound {
		// If permission doesn't exist, remove it from the state
		resp.State.RemoveResource(ctx)
		return
	}

	if response.StatusCode != 200 {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unexpected status code: %d", response.StatusCode))
		return
	}

	var respBody struct {
		Id   int    `json:"id"`
		Name string `json:"name"`
	}

	if err := json.NewDecoder(response.Body).Decode(&respBody); err != nil {
		resp.Diagnostics.AddError("Decode Error", fmt.Sprintf("Unable to decode response: %s", err))
		return
	}

	data.Id = types.StringValue(strconv.Itoa(respBody.Id))
	data.Name = types.StringValue(respBody.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PermissionGroup) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PermissionGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("%s/api/permissions/group/%s", r.client.Host, data.Id.ValueString())

	body := fmt.Sprintf(`{"name":"%s"}`, data.Name.ValueString())
	request, err := http.NewRequest("PUT", url, strings.NewReader(body))
	if err != nil {
		resp.Diagnostics.AddError("Request Error", fmt.Sprintf("Unable to create request: %s", err))
		return
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("x-api-key", r.client.ApiKey)

	response, err := r.client.Client.Do(request)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update permission group: %s", err))
		return
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unexpected status code: %d", response.StatusCode))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PermissionGroup) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PermissionGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("%s/api/permissions/group/%s", r.client.Host, data.Id.ValueString())

	request, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Request Error", fmt.Sprintf("Unable to create request: %s", err))
		return
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("x-api-key", r.client.ApiKey)

	response, err := r.client.Client.Do(request)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete permission group: %s", err))
		return
	}
	defer response.Body.Close()

	if response.StatusCode != 204 && response.StatusCode != 200 {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unexpected status code: %d", response.StatusCode))
		return
	}
}

func (r *PermissionGroup) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
