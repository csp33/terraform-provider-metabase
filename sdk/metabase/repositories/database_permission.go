// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/csp33/terraform-provider-metabase/sdk/metabase"
)

// The data permissions graph uses optimistic locking via a revision; a stale
// revision returns 409. Mutations are read-modify-write with bounded retry.
const graphMaxAttempts = 5

type dataGraph struct {
	Revision int                                  `json:"revision"`
	Groups   map[string]map[string]map[string]any `json:"groups"`
}

type DatabasePermissionRepository struct {
	client *metabase.MetabaseAPIClient
}

func NewDatabasePermissionRepository(client *metabase.MetabaseAPIClient) *DatabasePermissionRepository {
	return &DatabasePermissionRepository{client: client}
}

func (r *DatabasePermissionRepository) get(ctx context.Context) (*dataGraph, error) {
	resp, err := r.client.Get(ctx, "/api/permissions/graph")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var g dataGraph
	if err := json.NewDecoder(resp.Body).Decode(&g); err != nil {
		return nil, fmt.Errorf("failed to decode permissions graph: %w", err)
	}
	return &g, nil
}

// Get returns the create_queries level for a group/database edge. found is false
// when the group or database has no entry (e.g. either was deleted).
func (r *DatabasePermissionRepository) Get(ctx context.Context, groupId string, databaseId string) (createQueries string, found bool, err error) {
	g, err := r.get(ctx)
	if err != nil {
		return "", false, err
	}
	entry, ok := g.Groups[groupId][databaseId]
	if !ok {
		return "", false, nil
	}
	return createQueriesToString(entry["create-queries"]), true, nil
}

// Set applies create_queries for the group/database edge, merging into the graph
// with revision-based retry. view-data is intentionally not sent: on OSS it is
// always "unrestricted" (the only accepted value) and Metabase leaves it untouched
// when omitted; sending it adds no value and null would be rejected.
func (r *DatabasePermissionRepository) Set(ctx context.Context, groupId string, databaseId string, createQueries string) error {
	entry := map[string]any{"create-queries": createQueries}
	return r.putEdge(ctx, groupId, databaseId, entry)
}

func (r *DatabasePermissionRepository) putEdge(ctx context.Context, groupId string, databaseId string, entry map[string]any) error {
	for attempt := 1; ; attempt++ {
		g, err := r.get(ctx)
		if err != nil {
			return err
		}
		body := map[string]any{
			"revision": g.Revision,
			"groups":   map[string]any{groupId: map[string]any{databaseId: entry}},
		}
		_, err = r.client.Put(ctx, "/api/permissions/graph", body)
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

// isRetryableGraphError reports whether a graph write should be retried under
// read-modify-write: a 409 (stale revision) or a 5xx (Metabase computes the
// permissions/collection revision-log id in app code, so concurrent writes race
// on its primary key and surface as a 500 unique-constraint violation).
func isRetryableGraphError(err error) bool {
	var conflict *metabase.ConflictError
	if errors.As(err, &conflict) {
		return true
	}
	var base *metabase.BaseError
	if errors.As(err, &base) && base.StatusCode >= 500 {
		return true
	}
	return false
}

// createQueriesToString normalizes the create-queries value read from the graph:
// a string enum stays as-is, an absent value means no query access ("no"), and a
// granular object (set out-of-band) is returned as its JSON encoding.
func createQueriesToString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case nil:
		return "no"
	default:
		b, _ := json.Marshal(t)
		return string(b)
	}
}
