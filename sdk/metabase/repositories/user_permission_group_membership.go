// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/csp33/terraform-provider-metabase/sdk/metabase"
	"github.com/csp33/terraform-provider-metabase/sdk/metabase/models/dtos"
	"strconv"
)

type UserPermissionGroupMembershipRepository struct {
	client *metabase.MetabaseAPIClient
}

func NewUserPermissionGroupMembershipRepository(client *metabase.MetabaseAPIClient) *UserPermissionGroupMembershipRepository {
	return &UserPermissionGroupMembershipRepository{client: client}
}

func (r *UserPermissionGroupMembershipRepository) Create(ctx context.Context, userId string, groupId string) (*dtos.UserPermissionGroupMembershipDTO, error) {
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

func (r *UserPermissionGroupMembershipRepository) Delete(ctx context.Context, id string) error {
	path := fmt.Sprintf("/api/permissions/membership/%s", id)
	resp, err := r.client.Delete(ctx, path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
