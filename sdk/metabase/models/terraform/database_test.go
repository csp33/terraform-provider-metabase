// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package terraform

import (
	"encoding/json"
	"testing"
)

func mustDecode(t *testing.T, s string) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		t.Fatalf("invalid JSON %q: %v", s, err)
	}
	return m
}

func TestReconcileDetails(t *testing.T) {
	const state = `{"dbname":"sampledb","host":"db.old","password":"secret","port":5432,"ssl":false}`

	t.Run("no change keeps state verbatim", func(t *testing.T) {
		api := map[string]any{"dbname": "sampledb", "host": "db.old", "password": "**REDACTED**", "port": 5432.0, "ssl": false}
		got, err := ReconcileDetails(api, state, []string{"password"})
		if err != nil {
			t.Fatal(err)
		}
		if got != state {
			t.Fatalf("expected state verbatim, got %q", got)
		}
	})

	t.Run("drift on a non-secret field is detected", func(t *testing.T) {
		api := map[string]any{"dbname": "sampledb", "host": "db.new", "password": "**REDACTED**", "port": 5432.0, "ssl": false}
		got, err := ReconcileDetails(api, state, []string{"password"})
		if err != nil {
			t.Fatal(err)
		}
		m := mustDecode(t, got)
		if m["host"] != "db.new" {
			t.Fatalf("expected host drift to db.new, got %v", m["host"])
		}
		// The redacted secret is preserved from state, never taken from the API.
		if m["password"] != "secret" {
			t.Fatalf("expected password kept from state, got %v", m["password"])
		}
	})

	t.Run("metabase-added keys are ignored", func(t *testing.T) {
		api := map[string]any{"dbname": "sampledb", "host": "db.old", "password": "**REDACTED**", "port": 5432.0, "ssl": false, "auto-run-queries": true}
		got, err := ReconcileDetails(api, state, []string{"password"})
		if err != nil {
			t.Fatal(err)
		}
		if got != state {
			t.Fatalf("expected added key ignored (state verbatim), got %q", got)
		}
	})

	t.Run("keys metabase omits keep the state value", func(t *testing.T) {
		api := map[string]any{"dbname": "sampledb", "host": "db.old", "password": "**REDACTED**", "port": 5432.0} // ssl omitted
		got, err := ReconcileDetails(api, state, []string{"password"})
		if err != nil {
			t.Fatal(err)
		}
		if got != state {
			t.Fatalf("expected omitted key kept (state verbatim), got %q", got)
		}
	})

	t.Run("empty state is returned unchanged", func(t *testing.T) {
		got, err := ReconcileDetails(map[string]any{"host": "db.old"}, "", []string{"password"})
		if err != nil {
			t.Fatal(err)
		}
		if got != "" {
			t.Fatalf("expected empty string, got %q", got)
		}
	})
}
