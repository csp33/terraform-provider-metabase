// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// lowercaseValidator rejects values that are not already lowercase. Metabase
// stores emails lowercased server-side; if we let a mixed-case value through, the
// value returned after apply would differ from the planned value, producing
// Terraform's "Provider produced inconsistent result after apply" error and
// leaving an orphaned user in Metabase. Terraform does not allow a provider to
// silently rewrite a user-supplied value at plan time, so we fail fast with a
// clear message (before any API call) instead.
type lowercaseValidator struct{}

// LowercaseValidator returns a validator that requires the value to be lowercase.
func LowercaseValidator() validator.String {
	return lowercaseValidator{}
}

func (v lowercaseValidator) Description(_ context.Context) string {
	return "value must be lowercase"
}

func (v lowercaseValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v lowercaseValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	value := req.ConfigValue.ValueString()
	if value != strings.ToLower(value) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Email must be lowercase",
			fmt.Sprintf("Metabase stores emails in lowercase. Use %q instead of %q.", strings.ToLower(value), value),
		)
	}
}
