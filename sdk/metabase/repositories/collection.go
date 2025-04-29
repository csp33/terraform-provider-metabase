// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/csp33/terraform-provider-metabase/sdk/metabase"
	"github.com/csp33/terraform-provider-metabase/sdk/metabase/models/dtos"
)

type CollectionRepository struct {
	client *metabase.MetabaseAPIClient
}

func NewCollectionRepository(client *metabase.MetabaseAPIClient) *CollectionRepository {
	return &CollectionRepository{client: client}
}

func (r *CollectionRepository) Create(ctx context.Context, name string, parentId *string, archived bool) (*dtos.CollectionDTO, error) {
	body := map[string]string{"name": name, "archived": fmt.Sprintf("%t", archived)}
	if parentId != nil {
		body["parent_id"] = *parentId
	}

	resp, err := r.client.Post(ctx, "/api/collection", body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var res dtos.CollectionDTO
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode create response: %w", err)
	}
	return &res, nil
}

func (r *CollectionRepository) Get(ctx context.Context, id string) (*dtos.CollectionDTO, error) {
	path := fmt.Sprintf("/api/collection/%s", id)
	resp, err := r.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var res dtos.CollectionDTO
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode get response: %w", err)
	}
	return &res, nil
}

func (r *CollectionRepository) Update(ctx context.Context, id string, name *string, parentId *string, archived bool) (bool, error) {
	path := fmt.Sprintf("/api/collection/%s", id)
	body := map[string]any{"archived": archived}
	if name != nil {
		body["name"] = *name
	}

	if parentId != nil {
		body["parent_id"] = *parentId
	} else {
		body["parent_id"] = nil
	}

	resp, err := r.client.Put(ctx, path, body)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return true, nil
}
