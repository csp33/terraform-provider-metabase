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
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// testAccCheckEdge asserts the group's grant on a collection resource: want ""
// means "no grant", anything else is the expected permission level.
func testAccCheckEdge(collectionResource string, want string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		group, ok := s.RootModule().Resources["metabase_permission_group.test"]
		if !ok {
			return fmt.Errorf("group resource not found in state")
		}
		col, ok := s.RootModule().Resources[collectionResource]
		if !ok {
			return fmt.Errorf("%s not found in state", collectionResource)
		}
		repo := repositories.NewCollectionPermissionRepository(newTestMetabaseClient())
		perm, found, err := repo.Get(context.Background(), group.Primary.ID, col.Primary.ID)
		if err != nil {
			return err
		}
		if want == "" {
			if found {
				return fmt.Errorf("%s: expected no grant, got %q", collectionResource, perm)
			}
			return nil
		}
		if !found || perm != want {
			return fmt.Errorf("%s: expected %q, got (%q, found=%t)", collectionResource, want, perm, found)
		}
		return nil
	}
}

// testAccCheckSubtreeRevoked asserts that after destroy no grant survives on
// any of the tree's collections (delete must revoke the whole subtree).
func testAccCheckSubtreeRevoked(s *terraform.State) error {
	repo := repositories.NewCollectionPermissionRepository(newTestMetabaseClient())
	group, ok := s.RootModule().Resources["metabase_permission_group.test"]
	if !ok {
		return nil // group gone from state; its edges die with it anyway
	}
	for _, res := range []string{"metabase_collection.parent", "metabase_collection.child", "metabase_collection.grandchild"} {
		col, ok := s.RootModule().Resources[res]
		if !ok {
			continue
		}
		_, found, err := repo.Get(context.Background(), group.Primary.ID, col.Primary.ID)
		if err != nil {
			return err
		}
		if found {
			return fmt.Errorf("%s still granted after destroy", res)
		}
	}
	return nil
}

func testAccCollectionPermissionPropagateConfig(suffix int, permission string, propagate bool) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "metabase_permission_group" "test" {
  name = "tf_acc_prop_group_%[1]d"
}

resource "metabase_collection" "parent" {
  name = "tf_acc_prop_parent_%[1]d"
}

resource "metabase_collection" "child" {
  name      = "tf_acc_prop_child_%[1]d"
  parent_id = metabase_collection.parent.id
}

resource "metabase_collection" "grandchild" {
  name      = "tf_acc_prop_grandchild_%[1]d"
  parent_id = metabase_collection.child.id
}

resource "metabase_collection_permission" "test" {
  group_id      = metabase_permission_group.test.id
  collection_id = metabase_collection.parent.id
  permission    = %[2]q
  propagate     = %[3]t

  # The tree must exist before propagation expands it.
  depends_on = [metabase_collection.child, metabase_collection.grandchild]
}
`, suffix, permission, propagate)
}

func TestAccCollectionPermissionPropagate(t *testing.T) {
	suffix := rand.Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSubtreeRevoked,
		Steps: []resource.TestStep{
			{
				// propagate=false: only the root edge is written.
				Config: testAccCollectionPermissionPropagateConfig(suffix, "read", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("metabase_collection_permission.test", "propagate", "false"),
					testAccCheckEdge("metabase_collection.parent", "read"),
					testAccCheckEdge("metabase_collection.child", ""),
					testAccCheckEdge("metabase_collection.grandchild", ""),
				),
			},
			{
				// flip false->true: in-place update propagates to the subtree.
				Config: testAccCollectionPermissionPropagateConfig(suffix, "read", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("metabase_collection_permission.test", "propagate", "true"),
					testAccCheckEdge("metabase_collection.parent", "read"),
					testAccCheckEdge("metabase_collection.child", "read"),
					testAccCheckEdge("metabase_collection.grandchild", "read"),
				),
			},
			{
				// level change re-propagates the whole subtree.
				Config: testAccCollectionPermissionPropagateConfig(suffix, "write", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEdge("metabase_collection.parent", "write"),
					testAccCheckEdge("metabase_collection.child", "write"),
					testAccCheckEdge("metabase_collection.grandchild", "write"),
				),
			},
		},
	})
}
