// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package terraform

import (
	"github.com/csp33/terraform-provider-metabase/sdk/metabase/models/dtos"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strconv"
)

type CollectionTerraformModel struct {
	Id       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	ParentId types.String `tfsdk:"parent_id"`
	Archived types.Bool   `tfsdk:"archived"`
}

func CreateCollectionTerraformModelFromDTO(source *dtos.CollectionDTO) CollectionTerraformModel {
	var parentId types.String
	if source.ParentId == nil {
		parentId = types.StringNull()
	} else {
		parentId = types.StringValue(strconv.Itoa(*source.ParentId))
	}
	return CollectionTerraformModel{
		Id:       types.StringValue(strconv.Itoa(source.Id)),
		Name:     types.StringValue(source.Name),
		ParentId: parentId,
		Archived: types.BoolValue(source.Archived),
	}

}
