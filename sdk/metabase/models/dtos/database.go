// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dtos

type DatabaseDTO struct {
	Id      int            `json:"id"`
	Name    string         `json:"name"`
	Engine  string         `json:"engine"`
	Details map[string]any `json:"details"`
}
