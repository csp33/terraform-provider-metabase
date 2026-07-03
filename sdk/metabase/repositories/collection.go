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

	// Serialize collection creation in this process: it races on the
	// collection_revision id under concurrency (retries alone don't absorb it at
	// high parallelism). The retry below still guards inter-process contention.
	collectionCreateMu.Lock()
	defer collectionCreateMu.Unlock()

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
	body := map[string]any{}
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

// Archive sends the collection to the Trash. This is recoverable and does NOT
// permanently delete the collection or its contents (questions/dashboards) —
// "better safe than sorry" since collections hold data. Sends only "archived" so
// the collection's location/parent is preserved. Idempotent on 404.
func (r *CollectionRepository) Archive(ctx context.Context, id string) error {
	path := fmt.Sprintf("/api/collection/%s", id)
	resp, err := r.client.Put(ctx, path, map[string]any{"archived": true})
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
