// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package terraform

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/csp33/terraform-provider-metabase/sdk/metabase/models/dtos"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DatabaseTerraformModel struct {
	Id                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Engine             types.String `tfsdk:"engine"`
	Details            types.String `tfsdk:"details"`
	RedactedAttributes types.Set    `tfsdk:"redacted_attributes"`
	DeletionProtection types.Bool   `tfsdk:"deletion_protection"`
}

// details/redactedAttributes/deletionProtection are carried from plan/state, not
// the DTO (secrets are redacted server-side; the latter two are Terraform-only).
func CreateDatabaseTerraformModelFromDTO(source *dtos.DatabaseDTO, details types.String, redactedAttributes types.Set, deletionProtection types.Bool) DatabaseTerraformModel {
	return DatabaseTerraformModel{
		Id:                 types.StringValue(strconv.Itoa(source.Id)),
		Name:               types.StringValue(source.Name),
		Engine:             types.StringValue(source.Engine),
		Details:            details,
		RedactedAttributes: redactedAttributes,
		DeletionProtection: deletionProtection,
	}
}

// ReconcileDetails returns the details to store in state: a non-secret key takes
// the API value (so drift is detected); a redacted secret or a key Metabase omits
// keeps the state value; keys Metabase adds are ignored. Returns existingJSON
// unchanged when nothing differs.
func ReconcileDetails(apiDetails map[string]any, existingJSON string, redacted []string) (string, error) {
	var existing map[string]any
	if existingJSON != "" {
		if err := json.Unmarshal([]byte(existingJSON), &existing); err != nil {
			return "", fmt.Errorf("invalid details JSON in state: %w", err)
		}
	}
	if existing == nil {
		return existingJSON, nil
	}

	redactedSet := make(map[string]bool, len(redacted))
	for _, k := range redacted {
		redactedSet[k] = true
	}

	reconciled := make(map[string]any, len(existing))
	for k, existingValue := range existing {
		if redactedSet[k] {
			reconciled[k] = existingValue
			continue
		}
		if apiValue, ok := apiDetails[k]; ok {
			reconciled[k] = apiValue
		} else {
			reconciled[k] = existingValue
		}
	}

	if reflect.DeepEqual(existing, reconciled) {
		return existingJSON, nil
	}
	b, err := json.Marshal(reconciled)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
