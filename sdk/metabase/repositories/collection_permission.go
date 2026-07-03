// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/csp33/terraform-provider-metabase/sdk/metabase"
)

type collectionGraph struct {
	Revision int                          `json:"revision"`
	Groups   map[string]map[string]string `json:"groups"`
}

type CollectionPermissionRepository struct {
	client *metabase.MetabaseAPIClient
}

func NewCollectionPermissionRepository(client *metabase.MetabaseAPIClient) *CollectionPermissionRepository {
	return &CollectionPermissionRepository{client: client}
}

func (r *CollectionPermissionRepository) get(ctx context.Context) (*collectionGraph, error) {
	resp, err := r.client.Get(ctx, "/api/collection/graph")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var g collectionGraph
	if err := json.NewDecoder(resp.Body).Decode(&g); err != nil {
		return nil, fmt.Errorf("failed to decode collection graph: %w", err)
	}
	return &g, nil
}

// Get returns the permission of a group on a collection. found is false when there
// is no grant ("none" or absent, e.g. the group or collection was removed).
func (r *CollectionPermissionRepository) Get(ctx context.Context, groupId string, collectionId string) (permission string, found bool, err error) {
	g, err := r.get(ctx)
	if err != nil {
		return "", false, err
	}
	perm, ok := g.Groups[groupId][collectionId]
	if !ok || perm == "none" {
		return "", false, nil
	}
	return perm, true, nil
}

// Set applies a permission ("read"/"write"/"none") for the group/collection edge,
// merging into the graph with revision-based retry. Revoking is "none".
func (r *CollectionPermissionRepository) Set(ctx context.Context, groupId string, collectionId string, permission string) error {
	// Serialize all collection-graph writes in this process (same revision race as
	// the data graph). The retry below still guards inter-process contention.
	collectionGraphMu.Lock()
	defer collectionGraphMu.Unlock()

	for attempt := 1; ; attempt++ {
		g, err := r.get(ctx)
		if err != nil {
			return err
		}
		body := map[string]any{
			"revision": g.Revision,
			"groups":   map[string]any{groupId: map[string]any{collectionId: permission}},
		}
		_, err = r.client.Put(ctx, "/api/collection/graph", body)
		if err == nil {
			return nil
		}
		if attempt < graphMaxAttempts && isRetryableGraphError(err) {
			time.Sleep(time.Duration(attempt) * 100 * time.Millisecond)
			continue
		}
		return err
	}
}
