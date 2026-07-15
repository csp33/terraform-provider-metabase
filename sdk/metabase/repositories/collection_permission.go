// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
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

// Get returns a group's permission on a collection; found is false when there is
// no grant ("none" or absent).
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

// Set applies a permission ("read"/"write"/"none") for the group/collection edge.
func (r *CollectionPermissionRepository) Set(ctx context.Context, groupId string, collectionId string, permission string) error {
	return r.SetMany(ctx, groupId, []string{collectionId}, permission)
}

// SetMany applies the same permission to every collection in ONE graph PUT
// (single revision bump, no half-applied subtree). Used by `propagate`.
func (r *CollectionPermissionRepository) SetMany(ctx context.Context, groupId string, collectionIds []string, permission string) error {
	// Serialize in-process: the read-modify-write races on the shared revision id.
	collectionGraphMu.Lock()
	defer collectionGraphMu.Unlock()

	edges := map[string]any{}
	for _, id := range collectionIds {
		edges[id] = permission
	}

	for attempt := 1; ; attempt++ {
		g, err := r.get(ctx)
		if err != nil {
			return err
		}
		body := map[string]any{
			"revision": g.Revision,
			"groups":   map[string]any{groupId: edges},
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

// listedCollection is the subset of /api/collection items we need. Id is `any`
// because the listing includes the virtual root collection with id "root".
type listedCollection struct {
	Id              any    `json:"id"`
	Location        string `json:"location"`
	PersonalOwnerId *int   `json:"personal_owner_id"`
	Type            string `json:"type"`
}

func (r *CollectionPermissionRepository) list(ctx context.Context, archived bool) ([]listedCollection, error) {
	path := "/api/collection"
	if archived {
		path += "?archived=true"
	}
	resp, err := r.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var cols []listedCollection
	if err := json.NewDecoder(resp.Body).Decode(&cols); err != nil {
		return nil, fmt.Errorf("failed to decode collection list: %w", err)
	}
	return cols, nil
}

// ListDescendants returns the ids of every collection under collectionId
// (any depth), excluding personal collections and the virtual root. For
// collectionId "root" that is every non-personal collection. Archived
// descendants are included only when includeArchived is true: grants must not
// touch the Trash (restore recovers a collection's own permissions), but a
// revoke (Delete with propagate) must reach it so no access survives there.
func (r *CollectionPermissionRepository) ListDescendants(ctx context.Context, collectionId string, includeArchived bool) ([]string, error) {
	cols, err := r.list(ctx, false)
	if err != nil {
		return nil, err
	}
	if includeArchived {
		archived, err := r.list(ctx, true)
		if err != nil {
			return nil, err
		}
		cols = append(cols, archived...)
	}

	var out []string
	for _, c := range cols {
		id, ok := c.Id.(float64) // skips the virtual root ("root")
		if !ok || c.PersonalOwnerId != nil {
			continue
		}
		// The archived listing includes the Trash collection itself, whose
		// permissions Metabase refuses to edit (500).
		if c.Type == "trash" {
			continue
		}
		if collectionId == "root" { // everything hangs from the root
			out = append(out, strconv.Itoa(int(id)))
			continue
		}
		// location is the ancestor chain, e.g. "/10/24/": descendant iff one
		// of its segments is collectionId.
		for _, seg := range strings.Split(strings.Trim(c.Location, "/"), "/") {
			if seg == collectionId {
				out = append(out, strconv.Itoa(int(id)))
				break
			}
		}
	}
	return out, nil
}
