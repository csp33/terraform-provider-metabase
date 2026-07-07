// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package terraform

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DatabasePermissionTerraformModel struct {
	Id            types.String `tfsdk:"id"`
	GroupId       types.String `tfsdk:"group_id"`
	DatabaseId    types.String `tfsdk:"database_id"`
	CreateQueries types.String `tfsdk:"create_queries"`
}
