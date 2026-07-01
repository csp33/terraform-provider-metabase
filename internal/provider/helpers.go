// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

func idOf(left, right string) types.String {
	return types.StringValue(left + ":" + right)
}

func stringValue(s string) types.String {
	return types.StringValue(s)
}
