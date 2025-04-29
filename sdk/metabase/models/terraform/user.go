// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package terraform

import (
	"github.com/csp33/terraform-provider-metabase/sdk/metabase/models/dtos"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strconv"
)

type UserTerraformModel struct {
	Id       types.String `tfsdk:"id"`
	Email    types.String `tfsdk:"email"`
	IsActive types.Bool   `tfsdk:"is_active"`
}

func CreateUserTerraformModelFromDTO(source *dtos.UserDTO) UserTerraformModel {
	return UserTerraformModel{
		Id:       types.StringValue(strconv.Itoa(source.Id)),
		Email:    types.StringValue(source.Email),
		IsActive: types.BoolValue(source.IsActive),
	}

}
