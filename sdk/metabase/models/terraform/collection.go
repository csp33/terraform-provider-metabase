// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package terraform

import (
	"github.com/csp33/terraform-provider-metabase/sdk/metabase/models/dtos"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strconv"
	"strings"
)

type CollectionTerraformModel struct {
	Id       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	ParentId types.String `tfsdk:"parent_id"`
	Archived types.Bool   `tfsdk:"archived"`
}

func CreateCollectionTerraformModelFromDTO(source *dtos.CollectionDTO) CollectionTerraformModel {
	return CollectionTerraformModel{
		Id:       types.StringValue(strconv.Itoa(source.Id)),
		Name:     types.StringValue(source.Name),
		ParentId: parentIdFromLocation(source.Location),
		Archived: types.BoolValue(source.Archived),
	}
}

// parentIdFromLocation extracts the immediate parent id from a Metabase collection
// "location" path ("/" => root/null, "/12/" => 12, "/12/34/" => 34).
func parentIdFromLocation(location string) types.String {
	trimmed := strings.Trim(location, "/")
	if trimmed == "" {
		return types.StringNull()
	}
	parts := strings.Split(trimmed, "/")
	return types.StringValue(parts[len(parts)-1])
}
