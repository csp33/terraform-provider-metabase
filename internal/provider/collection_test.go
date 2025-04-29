// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func getCollectionName() string {
	return fmt.Sprintf("Test collection %d", rand.Int())
}

var collectionName = getCollectionName()

func TestAccCollectionResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccCollectionResourceConfig(collectionName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("metabase_collection.test", "name", collectionName),
					resource.TestCheckResourceAttr("metabase_collection.test", "archived", "false"),
					resource.TestCheckResourceAttrSet("metabase_collection.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "metabase_collection.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccCollectionResourceConfig("test-collection-updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("metabase_collection.test", "name", "test-collection-updated"),
				),
			},
			// Archive testing
			{
				Config: testAccCollectionResourceConfigArchived("test-collection-updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("metabase_collection.test", "archived", "true"),
				),
			},
		},
	})
}

func testAccCollectionResourceConfig(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "metabase_collection" "test" {
  name = "%s"
}
`, name)
}

func testAccCollectionResourceConfigArchived(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "metabase_collection" "test" {
  name     = "%s"
  archived = true
}
`, name)
}
