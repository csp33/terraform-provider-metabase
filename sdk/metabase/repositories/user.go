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

type UserRepository struct {
	client *metabase.MetabaseAPIClient
}

func NewUserRepository(client *metabase.MetabaseAPIClient) *UserRepository {
	return &UserRepository{client: client}
}

func (r *UserRepository) Create(ctx context.Context, email string, firstName string, lastName string) (*dtos.UserDTO, error) {
	body := map[string]string{"email": email, "first_name": firstName, "last_name": lastName}

	resp, err := r.client.Post(ctx, "/api/user", body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var res dtos.UserDTO
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode create response: %w", err)
	}
	return &res, nil
}

func (r *UserRepository) Get(ctx context.Context, id string) (*dtos.UserDTO, error) {
	path := fmt.Sprintf("/api/user/%s", id)
	resp, err := r.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var res dtos.UserDTO
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode get response: %w", err)
	}
	return &res, nil
}

func (r *UserRepository) deactivate(ctx context.Context, id string) (bool, error) {
	path := fmt.Sprintf("/api/user/%s", id)

	resp, err := r.client.Delete(ctx, path)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return true, nil
}

func (r *UserRepository) reactivate(ctx context.Context, id string) (bool, error) {
	path := fmt.Sprintf("/api/user/%s/reactivate", id)
	body := map[string]any{}

	resp, err := r.client.Put(ctx, path, body)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return true, nil
}

func (r *UserRepository) Update(ctx context.Context, id string, firstName *string, lastName *string, isActive *bool) (bool, error) {
	if firstName == nil && lastName == nil {
		return true, nil
	}

	path := fmt.Sprintf("/api/user/%s", id)
	body := map[string]any{"first_name": firstName, "last_name": lastName}

	resp, err := r.client.Put(ctx, path, body)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if isActive != nil {
		var err error
		if !*isActive {
			_, err = r.deactivate(ctx, id)
		} else {
			_, err = r.reactivate(ctx, id)
		}
		if err != nil {
			return false, err
		}
	}

	return true, nil
}
