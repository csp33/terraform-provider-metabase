// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/csp33/terraform-provider-metabase/sdk/metabase/repositories"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func getCollectionName() string {
	return fmt.Sprintf("Test collection %d", rand.Int())
}

var collectionName = getCollectionName()

// testAccCheckCollectionArchived asserts that destroy sends collections to the
// Trash (archived=true) and never permanently deletes them.
func testAccCheckCollectionArchived(s *terraform.State) error {
	repo := repositories.NewCollectionRepository(newTestMetabaseClient())
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "metabase_collection" {
			continue
		}
		c, err := repo.Get(context.Background(), rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("collection %s get failed after destroy: %w", rs.Primary.ID, err)
		}
		if !c.Archived {
			return fmt.Errorf("collection %s is not archived after destroy", rs.Primary.ID)
		}
	}
	return nil
}

func TestAccCollectionResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCollectionArchived,
		Steps: []resource.TestStep{
			// Create and Read.
			{
				Config: testAccCollectionResourceConfig(collectionName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("metabase_collection.test", "name", collectionName),
					resource.TestCheckResourceAttr("metabase_collection.test", "archived", "false"),
					resource.TestCheckResourceAttrSet("metabase_collection.test", "id"),
				),
			},
			// ImportState.
			{
				ResourceName:      "metabase_collection.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Rename in-place (not a replace).
			{
				Config: testAccCollectionResourceConfig("test-collection-updated"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("metabase_collection.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.TestCheckResourceAttr("metabase_collection.test", "name", "test-collection-updated"),
			},
			// Archive in-place.
			{
				Config: testAccCollectionResourceConfigArchived("test-collection-updated"),
				Check:  resource.TestCheckResourceAttr("metabase_collection.test", "archived", "true"),
			},
		},
	})
}

// TestAccCollectionResource_nested verifies nesting (parent_id derived from the
// Metabase "location" path) and moving a collection to another parent in-place.
func TestAccCollectionResource_nested(t *testing.T) {
	suffix := rand.Int()
	nameA := fmt.Sprintf("Test parent A %d", suffix)
	nameB := fmt.Sprintf("Test parent B %d", suffix)
	nameChild := fmt.Sprintf("Test child %d", suffix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCollectionArchived,
		Steps: []resource.TestStep{
			{
				Config: testAccCollectionNestedConfig(nameA, nameB, nameChild, "a"),
				Check: resource.TestCheckResourceAttrPair(
					"metabase_collection.child", "parent_id",
					"metabase_collection.parent_a", "id",
				),
			},
			{
				// Move the child under parent_b: in-place update.
				Config: testAccCollectionNestedConfig(nameA, nameB, nameChild, "b"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("metabase_collection.child", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.TestCheckResourceAttrPair(
					"metabase_collection.child", "parent_id",
					"metabase_collection.parent_b", "id",
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

func testAccCollectionNestedConfig(nameA, nameB, nameChild, parent string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "metabase_collection" "parent_a" { name = "%s" }
resource "metabase_collection" "parent_b" { name = "%s" }
resource "metabase_collection" "child" {
  name      = "%s"
  parent_id = metabase_collection.parent_%s.id
}
`, nameA, nameB, nameChild, parent)
}
