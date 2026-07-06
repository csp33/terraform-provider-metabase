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

// testAccCheckDatabasePermissionRevoked asserts that after destroy the edge is
// either gone (its group was deleted) or reset to no query access ("no").
func testAccCheckDatabasePermissionRevoked(s *terraform.State) error {
	repo := repositories.NewDatabasePermissionRepository(newTestMetabaseClient())
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "metabase_database_permission" {
			continue
		}
		groupId, databaseId, err := splitEdgeID(rs.Primary.ID)
		if err != nil {
			return err
		}
		createQueries, found, err := repo.Get(context.Background(), groupId, databaseId)
		if err != nil {
			return err
		}
		if found && createQueries != "no" {
			return fmt.Errorf("database permission %s still grants %q after destroy", rs.Primary.ID, createQueries)
		}
	}
	return nil
}

func TestAccDatabasePermissionResource(t *testing.T) {
	name := fmt.Sprintf("Test graph group %d", rand.Int())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDatabasePermissionRevoked,
		Steps: []resource.TestStep{
			// Create: two edges in one apply exercise the revision retry (concurrency).
			{
				Config: testAccDatabasePermissionConfig(name, "query-builder"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("metabase_database_permission.test", "database_id", "metabase_database.a", "id"),
					resource.TestCheckResourceAttr("metabase_database_permission.test", "create_queries", "query-builder"),
					resource.TestCheckResourceAttrPair("metabase_database_permission.test", "group_id", "metabase_permission_group.test", "id"),
				),
			},
			// Import.
			{
				ResourceName:      "metabase_database_permission.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update create_queries in-place.
			{
				Config: testAccDatabasePermissionConfig(name, "query-builder-and-native"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("metabase_database_permission.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.TestCheckResourceAttr("metabase_database_permission.test", "create_queries", "query-builder-and-native"),
			},
		},
	})
}

func testAccDatabasePermissionConfig(groupName, createQueries string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "metabase_database" "a" {
  name                = "%[1]s db A"
  engine              = "postgres"
  deletion_protection = false
  details = jsonencode({
    host     = "sample-db"
    port     = 5432
    dbname   = "sampledb"
    user     = "sampleuser"
    password = "samplepass"
    ssl      = false
  })
}

resource "metabase_database" "b" {
  name                = "%[1]s db B"
  engine              = "postgres"
  deletion_protection = false
  details = jsonencode({
    host     = "sample-db"
    port     = 5432
    dbname   = "sampledb"
    user     = "sampleuser"
    password = "samplepass"
    ssl      = false
  })
}

resource "metabase_permission_group" "test" {
  name = "%[1]s"
}

resource "metabase_database_permission" "test" {
  group_id       = metabase_permission_group.test.id
  database_id    = metabase_database.a.id
  create_queries = "%[2]s"
}

resource "metabase_database_permission" "other" {
  group_id       = metabase_permission_group.test.id
  database_id    = metabase_database.b.id
  create_queries = "no"
}
`, groupName, createQueries)
}
