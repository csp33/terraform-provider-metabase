// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/csp33/terraform-provider-metabase/sdk/metabase"
	"github.com/csp33/terraform-provider-metabase/sdk/metabase/models"
	"strconv"
)

type PermissionGroupRepository struct {
	client *metabase.MetabaseAPIClient
}

func NewPermissionGroupRepository(client *metabase.MetabaseAPIClient) *PermissionGroupRepository {
	return &PermissionGroupRepository{client: client}
}

func (r *PermissionGroupRepository) Create(ctx context.Context, name string) (*models.PermissionGroup, error) {
	body := map[string]string{"name": name}
	resp, err := r.client.Post(ctx, "/api/permissions/group", body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var res models.PermissionGroup
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode create response: %w", err)
	}
	return &res, nil
}

func (r *PermissionGroupRepository) Get(ctx context.Context, id int) (*models.PermissionGroup, error) {
	path := fmt.Sprintf("/api/permissions/group/%d", id)
	resp, err := r.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var res models.PermissionGroup
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode get response: %w", err)
	}
	return &res, nil
}

func (r *PermissionGroupRepository) Update(ctx context.Context, id int, name string) (bool, error) {
	path := fmt.Sprintf("/api/permissions/group/%d", id)
	body := map[string]string{"name": name}

	resp, err := r.client.Put(ctx, path, body)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return true, nil
}

func (r *PermissionGroupRepository) Delete(ctx context.Context, id int) error {
	path := fmt.Sprintf("/api/permissions/group/%s", strconv.Itoa(id))
	resp, err := r.client.Delete(ctx, path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
