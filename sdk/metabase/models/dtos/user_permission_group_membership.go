// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dtos

type UserPermissionGroupMembershipDTO struct {
	MembershipId int `json:"membership_id"`
	UserId       int `json:"user_id"`
	GroupId      int `json:"group_id"`
}
