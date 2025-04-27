// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package terraform

import (
	"github.com/csp33/terraform-provider-metabase/sdk/metabase/models/dtos"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PermissionGroupTerraformModel struct {
	Name types.String `tfsdk:"name"`
	Id   types.Int32  `tfsdk:"id"`
}

func CreatePermissionGroupTerraformModelFromDTO(source *dtos.PermissionGroupDTO) PermissionGroupTerraformModel {
	return PermissionGroupTerraformModel{
		Name: types.StringValue(source.Name),
		Id:   types.Int32Value(int32(source.Id)),
	}

}
