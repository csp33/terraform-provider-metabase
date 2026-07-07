// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"testing"

	"github.com/csp33/terraform-provider-metabase/sdk/metabase/repositories"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// testAccCheckCollectionPermissionRevoked asserts that after destroy the group has
// no grant on the collection.
func testAccCheckCollectionPermissionRevoked(s *terraform.State) error {
	repo := repositories.NewCollectionPermissionRepository(newTestMetabaseClient())
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "metabase_collection_permission" {
			continue
		}
		groupId, collectionId, err := splitEdgeID(rs.Primary.ID)
		if err != nil {
			return err
		}
		_, found, err := repo.Get(context.Background(), groupId, collectionId)
		if err != nil {
			return err
		}
		if found {
			return fmt.Errorf("collection permission %s still granted after destroy", rs.Primary.ID)
		}
	}
	return nil
}

func TestAccCollectionPermissionResource(t *testing.T) {
	suffix := rand.Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCollectionPermissionRevoked,
		Steps: []resource.TestStep{
			{
				Config: testAccCollectionPermissionConfig(suffix, "read"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("metabase_collection_permission.test", "permission", "read"),
					resource.TestCheckResourceAttrPair("metabase_collection_permission.test", "group_id", "metabase_permission_group.test", "id"),
					resource.TestCheckResourceAttrPair("metabase_collection_permission.test", "collection_id", "metabase_collection.test", "id"),
				),
			},
			{
				ResourceName:      "metabase_collection_permission.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCollectionPermissionConfig(suffix, "write"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("metabase_collection_permission.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.TestCheckResourceAttr("metabase_collection_permission.test", "permission", "write"),
			},
		},
	})
}

// TestAccCollectionPermissionResource_invalidPermission asserts the validator rejects
// values other than read/write.
func TestAccCollectionPermissionResource_invalidPermission(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccCollectionPermissionConfig(rand.Int(), "none"),
				ExpectError: regexp.MustCompile("must be one of"),
			},
		},
	})
}

func testAccCollectionPermissionConfig(suffix int, permission string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "metabase_permission_group" "test" {
  name = "Test cperm group %d"
}

resource "metabase_collection" "test" {
  name = "Test cperm collection %d"
}

resource "metabase_collection_permission" "test" {
  group_id      = metabase_permission_group.test.id
  collection_id = metabase_collection.test.id
  permission    = "%s"
}
`, suffix, suffix, permission)
}
