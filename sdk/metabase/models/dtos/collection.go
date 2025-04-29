// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dtos

type CollectionDTO struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	ParentId *int   `json:"parent_id"`
	Archived bool   `json:"archived"`
}
