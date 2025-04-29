// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"os"
	"testing"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"metabase": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	// You can add code here to run prior to any test case execution, for example assertions
	// about the appropriate environment variables being set are common to see in a pre-check
	// function.
	if os.Getenv("METABASE_HOST") == "" {
		t.Fatal("METABASE_HOST must be set for acceptance tests")
	}
	if os.Getenv("METABASE_API_KEY") == "" {
		t.Fatal("METABASE_API_KEY must be set for acceptance tests")
	}
}

// for acceptance tests.
func testAccProviderConfig() string {
	return `
provider "metabase" {
  host    = "` + os.Getenv("METABASE_HOST") + `"
  api_key = "` + os.Getenv("METABASE_API_KEY") + `"
}
`
}
