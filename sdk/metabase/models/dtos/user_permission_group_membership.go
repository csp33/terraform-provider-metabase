// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dtos

type UserPermissionGroupMembershipDTO struct {
	MembershipId int    `json:"membership_id"`
	UserId       string `json:"user_id"`
	GroupId      string `json:"group_id"`
}
