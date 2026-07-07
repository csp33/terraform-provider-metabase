// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/csp33/terraform-provider-metabase/sdk/metabase"
	"github.com/csp33/terraform-provider-metabase/sdk/metabase/models/dtos"
)

type UserPermissionGroupMembershipRepository struct {
	client *metabase.MetabaseAPIClient
}

func NewUserPermissionGroupMembershipRepository(client *metabase.MetabaseAPIClient) *UserPermissionGroupMembershipRepository {
	return &UserPermissionGroupMembershipRepository{client: client}
}

func (r *UserPermissionGroupMembershipRepository) Create(ctx context.Context, userId string, groupId string) (*dtos.UserPermissionGroupMembershipDTO, error) {
	// A duplicate (user, group) membership 500s; pre-check and return an import hint.
	if existing, err := r.findByUserAndGroup(ctx, userId, groupId); err != nil {
		return nil, err
	} else if existing != nil {
		return nil, fmt.Errorf(
			"user %s is already a member of group %s (membership id %d); Terraform will not adopt it. Import it instead: `terraform import metabase_user_permission_group_membership.<name> %d`",
			userId, groupId, existing.MembershipId, existing.MembershipId,
		)
	}

	body := map[string]string{"group_id": groupId, "user_id": userId}
	resp, err := r.client.Post(ctx, "/api/permissions/membership", body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var res []dtos.UserPermissionGroupMembershipDTO
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode create response: %w", err)
	}

	for _, m := range res {
		if strconv.Itoa(m.UserId) == userId && strconv.Itoa(m.GroupId) == groupId {
			return &m, nil
		}
	}

	return nil, metabase.NewNotFoundError(fmt.Sprintf("Membership not found for user_id %s and group_id %s", userId, groupId))

}

func (r *UserPermissionGroupMembershipRepository) Get(ctx context.Context, id string) (*dtos.UserPermissionGroupMembershipDTO, error) {
	path := "/api/permissions/membership"
	resp, err := r.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var res map[string][]dtos.UserPermissionGroupMembershipDTO

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode get response: %w", err)
	}

	for _, memberships := range res {
		for _, m := range memberships {
			if strconv.Itoa(m.MembershipId) == id {
				return &m, nil
			}
		}
	}
	return nil, metabase.NewNotFoundError(fmt.Sprintf("Membership with ID %s not found", id))
}

// findByUserAndGroup returns the membership linking this user and group, or nil.
func (r *UserPermissionGroupMembershipRepository) findByUserAndGroup(ctx context.Context, userId string, groupId string) (*dtos.UserPermissionGroupMembershipDTO, error) {
	resp, err := r.client.Get(ctx, "/api/permissions/membership")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var res map[string][]dtos.UserPermissionGroupMembershipDTO
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode membership list: %w", err)
	}
	for _, memberships := range res {
		for i := range memberships {
			m := memberships[i]
			if strconv.Itoa(m.UserId) == userId && strconv.Itoa(m.GroupId) == groupId {
				return &m, nil
			}
		}
	}
	return nil, nil
}

func (r *UserPermissionGroupMembershipRepository) Delete(ctx context.Context, id string) error {
	path := fmt.Sprintf("/api/permissions/membership/%s", id)
	resp, err := r.client.Delete(ctx, path)
	if err != nil {
		// Idempotent: a membership already gone (e.g. group deleted) is a success.
		var notFound *metabase.NotFoundError
		if errors.As(err, &notFound) {
			return nil
		}
		return err
	}
	defer resp.Body.Close()

	return nil
}
