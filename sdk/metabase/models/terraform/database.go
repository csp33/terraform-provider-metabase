// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package terraform

import (
	"github.com/csp33/terraform-provider-metabase/sdk/metabase/models/dtos"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strconv"
)

type DatabaseTerraformModel struct {
	Id      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Engine  types.String `tfsdk:"engine"`
	Details types.String `tfsdk:"details"`
}

// details is carried from plan/state, not the DTO: Metabase redacts secrets and
// injects extra default keys, so the API response is not round-trip safe.
func CreateDatabaseTerraformModelFromDTO(source *dtos.DatabaseDTO, details types.String) DatabaseTerraformModel {
	return DatabaseTerraformModel{
		Id:      types.StringValue(strconv.Itoa(source.Id)),
		Name:    types.StringValue(source.Name),
		Engine:  types.StringValue(source.Engine),
		Details: details,
	}
}
