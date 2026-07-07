// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dtos

// Metabase returns the parent as a path in "location" (e.g. "/", "/12/", "/12/34/"),
// not as a parent_id field.
type CollectionDTO struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Location string `json:"location"`
	Archived bool   `json:"archived"`
}
