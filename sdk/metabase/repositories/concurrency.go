// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package repositories

import "sync"

// Concurrency control shared across all resource instances in a single provider
// process (one per `terraform apply`). Terraform runs resource operations
// concurrently (up to -parallelism), so these package-level primitives bound or
// serialize the Metabase calls that are not concurrency-safe, keeping bulk applies
// (e.g. 100 resources at once) from overwhelming or racing Metabase.
//
// Chosen empirically by hammering each endpoint at 25-30x concurrency:
//   - Serialized (mutex): operations that race on an app-computed revision-log id.
//   - data permissions graph writes   -> permissions_revision (409/500)
//   - collection permissions graph     -> collection_permission_graph_revision (500)
//   - collection creation              -> collection_revision (500)
//   - Bounded (semaphore): concurrency-safe but heavy (connection test + schema sync).
//   - database create/update
//   - Unbounded: proven safe at 25-30x (user, group, membership, deletes, renames...).
//
// The retries in the repositories remain as a second layer for INTER-process
// contention (another apply or a UI edit at the same time); the mutexes eliminate
// the INTRA-apply races that retries alone did not absorb at high concurrency.
var (
	permissionsGraphMu sync.Mutex
	collectionGraphMu  sync.Mutex
	collectionCreateMu sync.Mutex

	// databaseWriteSem bounds concurrent database create/update so a bulk apply
	// doesn't launch many heavy connection tests / schema syncs at once.
	databaseWriteSem = make(chan struct{}, 4)
)

func acquireDatabaseWrite() { databaseWriteSem <- struct{}{} }
func releaseDatabaseWrite() { <-databaseWriteSem }
