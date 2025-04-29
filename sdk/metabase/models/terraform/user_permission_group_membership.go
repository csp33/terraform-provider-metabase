// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package terraform

import (
	"github.com/csp33/terraform-provider-metabase/sdk/metabase/models/dtos"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strconv"
)

type UserPermissionGroupMembershipTerraformModel struct {
	Id                types.String `tfsdk:"id"`
	UserId            types.String `tfsdk:"user_id"`
	PermissionGroupId types.String `tfsdk:"permission_group_id"`
}

func CreateUserPermissionGroupMembershipTerraformModelFromDTO(source *dtos.UserPermissionGroupMembershipDTO) UserPermissionGroupMembershipTerraformModel {
	return UserPermissionGroupMembershipTerraformModel{
		Id:                types.StringValue(strconv.Itoa(source.MembershipId)),
		UserId:            types.StringValue(strconv.Itoa(source.UserId)),
		PermissionGroupId: types.StringValue(strconv.Itoa(source.GroupId)),
	}

}
