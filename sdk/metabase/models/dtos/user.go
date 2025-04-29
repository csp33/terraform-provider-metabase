// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dtos

type UserDTO struct {
	Id       int    `json:"id"`
	Email    string `json:"email"`
	IsActive bool   `json:"is_active"`
}
