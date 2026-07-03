// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package terraform

import (
	"github.com/csp33/terraform-provider-metabase/sdk/metabase/models/dtos"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strconv"
)

type DatabaseTerraformModel struct {
	Id                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Engine             types.String `tfsdk:"engine"`
	Details            types.String `tfsdk:"details"`
	DeletionProtection types.Bool   `tfsdk:"deletion_protection"`
}

// details/deletionProtection are carried from plan/state, not the DTO (secrets are
// redacted server-side; deletion_protection is a Terraform-only guard).
func CreateDatabaseTerraformModelFromDTO(source *dtos.DatabaseDTO, details types.String, deletionProtection types.Bool) DatabaseTerraformModel {
	return DatabaseTerraformModel{
		Id:                 types.StringValue(strconv.Itoa(source.Id)),
		Name:               types.StringValue(source.Name),
		Engine:             types.StringValue(source.Engine),
		Details:            details,
		DeletionProtection: deletionProtection,
	}
}
