// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package terraform

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type CollectionPermissionTerraformModel struct {
	Id           types.String `tfsdk:"id"`
	GroupId      types.String `tfsdk:"group_id"`
	CollectionId types.String `tfsdk:"collection_id"`
	Permission   types.String `tfsdk:"permission"`
}
