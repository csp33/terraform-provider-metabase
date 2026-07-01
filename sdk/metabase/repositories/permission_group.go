// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/csp33/terraform-provider-metabase/sdk/metabase"
	"github.com/csp33/terraform-provider-metabase/sdk/metabase/models/dtos"
)

type PermissionGroupRepository struct {
	client *metabase.MetabaseAPIClient
}

func NewPermissionGroupRepository(client *metabase.MetabaseAPIClient) *PermissionGroupRepository {
	return &PermissionGroupRepository{client: client}
}

func (r *PermissionGroupRepository) Create(ctx context.Context, name string) (*dtos.PermissionGroupDTO, error) {
	body := map[string]string{"name": name}
	resp, err := r.client.Post(ctx, "/api/permissions/group", body)
	if err != nil {
		// Group names are unique; Metabase returns 400 "A group with that name
		// already exists." Rather than adopt an existing group, fail with a
		// `terraform import` hint (groups are hard-deleted, so there is no
		// deactivated-vs-active nuance as with users).
		var badRequest *metabase.BadRequestError
		if errors.As(err, &badRequest) && strings.Contains(strings.ToLower(badRequest.Message), "already exists") {
			return nil, r.nameInUseError(ctx, name)
		}
		return nil, err
	}
	defer resp.Body.Close()

	var res dtos.PermissionGroupDTO
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode create response: %w", err)
	}
	return &res, nil
}

// FindByName returns the group with the given exact name, or nil if none exists.
func (r *PermissionGroupRepository) FindByName(ctx context.Context, name string) (*dtos.PermissionGroupDTO, error) {
	resp, err := r.client.Get(ctx, "/api/permissions/group")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var groups []dtos.PermissionGroupDTO
	if err := json.NewDecoder(resp.Body).Decode(&groups); err != nil {
		return nil, fmt.Errorf("failed to decode group list: %w", err)
	}
	for i := range groups {
		if groups[i].Name == name {
			return &groups[i], nil
		}
	}
	return nil, nil
}

// nameInUseError builds an actionable "already exists" error enriched with the id.
func (r *PermissionGroupRepository) nameInUseError(ctx context.Context, name string) error {
	existing, err := r.FindByName(ctx, name)
	if err == nil && existing != nil {
		return fmt.Errorf(
			"a permission group named %q already exists (id %d); Terraform will not adopt it. Import it instead: `terraform import metabase_permission_group.<name> %d`",
			name, existing.Id, existing.Id,
		)
	}
	return fmt.Errorf("a permission group named %q already exists; import it with `terraform import metabase_permission_group.<name> <id>` instead of creating it", name)
}

func (r *PermissionGroupRepository) Get(ctx context.Context, id string) (*dtos.PermissionGroupDTO, error) {
	path := fmt.Sprintf("/api/permissions/group/%s", id)
	// TODO: mark as deleted if 404
	resp, err := r.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var res dtos.PermissionGroupDTO
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode get response: %w", err)
	}
	return &res, nil
}

func (r *PermissionGroupRepository) Update(ctx context.Context, id string, name string) (bool, error) {
	path := fmt.Sprintf("/api/permissions/group/%s", id)
	body := map[string]string{"name": name}

	resp, err := r.client.Put(ctx, path, body)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return true, nil
}

func (r *PermissionGroupRepository) Delete(ctx context.Context, id string) error {
	path := fmt.Sprintf("/api/permissions/group/%s", id)
	resp, err := r.client.Delete(ctx, path)
	if err != nil {
		// Idempotent delete: if the group is already gone, treat as success.
		var notFound *metabase.NotFoundError
		if errors.As(err, &notFound) {
			return nil
		}
		return err
	}
	defer resp.Body.Close()

	return nil
}
