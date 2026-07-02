// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

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
		// Metabase never truly deletes users: it deactivates them, and the email
		// stays reserved. Re-creating a user whose email already exists (active or
		// deactivated) returns 400 "Email address already in use." Rather than
		// silently adopting/mutating a pre-existing account, we fail with an
		// actionable error pointing to `terraform import` (the idiomatic way to
		// bring an existing resource under management). Reactivating a deactivated
		// user is then done by importing it and setting is_active = true.
		var badRequest *metabase.BadRequestError
		if errors.As(err, &badRequest) && strings.Contains(strings.ToLower(badRequest.Message), "already in use") {
			return nil, r.emailInUseError(ctx, email)
		}
		return nil, err
	}
	defer resp.Body.Close()

	var res dtos.UserDTO
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode create response: %w", err)
	}
	return &res, nil
}

// emailInUseError builds an actionable error for the "email already in use" case,
// enriched with the existing user's id and active state so the operator can import
// it. It never mutates the existing user.
func (r *UserRepository) emailInUseError(ctx context.Context, email string) error {
	existing, err := r.FindByEmail(ctx, email)
	if err == nil && existing != nil {
		return fmt.Errorf(
			"a user with email %q already exists (id %d, is_active=%t); Terraform will not adopt it. Import it instead: `terraform import metabase_user.<name> %d` (to reactivate a deactivated user, import it and set is_active = true)",
			email, existing.Id, existing.IsActive, existing.Id,
		)
	}
	return fmt.Errorf("a user with email %q already exists in Metabase; import it with `terraform import metabase_user.<name> <id>` instead of creating it", email)
}

// FindByEmail returns the user with the given email (active or deactivated), or
// nil if none exists. Matching is case-insensitive.
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*dtos.UserDTO, error) {
	path := fmt.Sprintf("/api/user?status=all&query=%s", url.QueryEscape(email))
	resp, err := r.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var listResponse struct {
		Data []dtos.UserDTO `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&listResponse); err != nil {
		return nil, fmt.Errorf("failed to decode user search response: %w", err)
	}

	for i := range listResponse.Data {
		if strings.EqualFold(listResponse.Data[i].Email, email) {
			return &listResponse.Data[i], nil
		}
	}
	return nil, nil
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
		// Idempotent: reactivating an already-active user returns 400
		// ("Not able to reactivate an active user"); treat that as success.
		var badRequest *metabase.BadRequestError
		if errors.As(err, &badRequest) && strings.Contains(strings.ToLower(badRequest.Message), "active user") {
			return true, nil
		}
		return false, err
	}
	defer resp.Body.Close()

	return true, nil
}

func (r *UserRepository) Update(ctx context.Context, id string, email *string, firstName *string, lastName *string, isActive *bool) (bool, error) {
	// Order matters: Metabase returns 404 on PUT /api/user/:id for a deactivated
	// user, so we must reactivate BEFORE editing fields, and deactivate AFTER.
	//   - reactivate first  -> field edits target an active user
	//   - edit fields (only when active)
	//   - deactivate last   -> fields were editable while still active
	if isActive != nil && *isActive {
		if _, err := r.reactivate(ctx, id); err != nil {
			return false, err
		}
	}

	// Update mutable fields only when at least one is provided. Email is mutable
	// in Metabase via PUT (no replace needed).
	if email != nil || firstName != nil || lastName != nil {
		body := map[string]any{}
		if email != nil {
			body["email"] = *email
		}
		if firstName != nil {
			body["first_name"] = *firstName
		}
		if lastName != nil {
			body["last_name"] = *lastName
		}

		path := fmt.Sprintf("/api/user/%s", id)
		resp, err := r.client.Put(ctx, path, body)
		if err != nil {
			return false, err
		}
		resp.Body.Close()
	}

	if isActive != nil && !*isActive {
		if _, err := r.deactivate(ctx, id); err != nil {
			return false, err
		}
	}

	return true, nil
}
