// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package terraform

import (
	"github.com/csp33/terraform-provider-metabase/sdk/metabase/models/dtos"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PermissionGroupTerraformModel struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func CreatePermissionGroupTerraformModelFromDTO(source *dtos.PermissionGroupDTO) PermissionGroupTerraformModel {
	return PermissionGroupTerraformModel{
		Id:   types.StringValue(source.Id),
		Name: types.StringValue(source.Name),
	}

}
