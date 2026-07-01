// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/csp33/terraform-provider-metabase/sdk/metabase"
	"github.com/csp33/terraform-provider-metabase/sdk/metabase/models/dtos"
)

// Metabase's collection INSERT is not concurrency-safe: creating several
// collections in one apply can fail with a 5xx unique-constraint race. Retry a
// few times before giving up.
const collectionCreateMaxAttempts = 4

type CollectionRepository struct {
	client *metabase.MetabaseAPIClient
}

func NewCollectionRepository(client *metabase.MetabaseAPIClient) *CollectionRepository {
	return &CollectionRepository{client: client}
}

func (r *CollectionRepository) Create(ctx context.Context, name string, parentId *string, archived *bool) (*dtos.CollectionDTO, error) {
	body := map[string]any{"name": name}
	if parentId != nil {
		body["parent_id"] = *parentId
	}
	if archived != nil {
		body["archived"] = *archived
	} else {
		body["archived"] = false
	}

	var resp *http.Response
	var err error
	for attempt := 1; ; attempt++ {
		resp, err = r.client.Post(ctx, "/api/collection", body)
		if err == nil {
			break
		}
		var baseErr *metabase.BaseError
		if attempt < collectionCreateMaxAttempts && errors.As(err, &baseErr) && baseErr.StatusCode >= 500 {
			time.Sleep(time.Duration(attempt) * 150 * time.Millisecond)
			continue
		}
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

func (r *CollectionRepository) Update(ctx context.Context, id string, name *string, parentId *string, archived *bool) (bool, error) {
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

	if archived != nil {
		body["archived"] = *archived
	} else {
		body["archived"] = false
	}

	resp, err := r.client.Put(ctx, path, body)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return true, nil
}

// Delete permanently removes a collection. Metabase only permanently deletes
// ARCHIVED collections, so archive first (idempotent) then delete. Sending a
// collection to the Trash without deleting it is done via the archived attribute.
func (r *CollectionRepository) Delete(ctx context.Context, id string) error {
	archived := true
	if _, err := r.Update(ctx, id, nil, nil, &archived); err != nil {
		var notFound *metabase.NotFoundError
		if errors.As(err, &notFound) {
			return nil
		}
		return err
	}

	path := fmt.Sprintf("/api/collection/%s", id)
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
