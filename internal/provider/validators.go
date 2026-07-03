// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// lowercaseValidator rejects non-lowercase values: Metabase lowercases emails
// server-side, so a mixed-case value would fail apply and orphan the user.
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

// oneOfValidator restricts a string to a fixed set of allowed values.
type oneOfValidator struct{ allowed []string }

// OneOfValidator returns a validator that accepts only the given values.
func OneOfValidator(allowed ...string) validator.String {
	return oneOfValidator{allowed: allowed}
}

func (v oneOfValidator) Description(_ context.Context) string {
	return fmt.Sprintf("value must be one of %s", strings.Join(v.allowed, ", "))
}

func (v oneOfValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v oneOfValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	value := req.ConfigValue.ValueString()
	for _, a := range v.allowed {
		if value == a {
			return
		}
	}
	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Invalid value",
		fmt.Sprintf("%q must be one of: %s", value, strings.Join(v.allowed, ", ")),
	)
}
