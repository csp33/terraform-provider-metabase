// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package repositories

import "sync"

// Package-level concurrency control shared across all resources in one provider
// process. Serializes writes that race on an app-computed revision id (data and
// collection graphs, collection create; a 409/5xx otherwise) and bounds database
// create/update (heavy connection test and schema sync). Everything else is
// unbounded. Repository retries remain a second layer for inter-process contention.
var (
	permissionsGraphMu sync.Mutex
	collectionGraphMu  sync.Mutex
	collectionCreateMu sync.Mutex
	databaseWriteSem   = make(chan struct{}, 4)
)

func acquireDatabaseWrite() { databaseWriteSem <- struct{}{} }
func releaseDatabaseWrite() { <-databaseWriteSem }
