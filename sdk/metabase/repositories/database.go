// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/csp33/terraform-provider-metabase/sdk/metabase"
	"github.com/csp33/terraform-provider-metabase/sdk/metabase/models/dtos"
)

type DatabaseRepository struct {
	client *metabase.MetabaseAPIClient
}

func NewDatabaseRepository(client *metabase.MetabaseAPIClient) *DatabaseRepository {
	return &DatabaseRepository{client: client}
}

func decodeDetails(detailsJSON string) (map[string]any, error) {
	var details map[string]any
	if err := json.Unmarshal([]byte(detailsJSON), &details); err != nil {
		return nil, fmt.Errorf("invalid details JSON: %w", err)
	}
	return details, nil
}

func (r *DatabaseRepository) Create(ctx context.Context, name string, engine string, detailsJSON string) (*dtos.DatabaseDTO, error) {
	details, err := decodeDetails(detailsJSON)
	if err != nil {
		return nil, err
	}
	body := map[string]any{"name": name, "engine": engine, "details": details}

	resp, err := r.client.Post(ctx, "/api/database", body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var res dtos.DatabaseDTO
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode create response: %w", err)
	}
	return &res, nil
}

func (r *DatabaseRepository) Get(ctx context.Context, id string) (*dtos.DatabaseDTO, error) {
	path := fmt.Sprintf("/api/database/%s", id)
	resp, err := r.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var res dtos.DatabaseDTO
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode get response: %w", err)
	}
	return &res, nil
}

func (r *DatabaseRepository) Update(ctx context.Context, id string, name *string, detailsJSON *string) (bool, error) {
	body := map[string]any{}
	if name != nil {
		body["name"] = *name
	}
	if detailsJSON != nil {
		details, err := decodeDetails(*detailsJSON)
		if err != nil {
			return false, err
		}
		body["details"] = details
	}

	path := fmt.Sprintf("/api/database/%s", id)
	resp, err := r.client.Put(ctx, path, body)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return true, nil
}

func (r *DatabaseRepository) Delete(ctx context.Context, id string) error {
	path := fmt.Sprintf("/api/database/%s", id)
	resp, err := r.client.Delete(ctx, path)
	if err != nil {
		var notFound *metabase.NotFoundError
		if errors.As(err, &notFound) {
			return nil
		}
		return err
	}
	defer resp.Body.Close()

	return nil
}
