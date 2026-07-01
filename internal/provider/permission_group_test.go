// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"testing"

	"github.com/csp33/terraform-provider-metabase/sdk/metabase"
	"github.com/csp33/terraform-provider-metabase/sdk/metabase/repositories"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func getGroupName() string {
	return fmt.Sprintf("Test group %d", rand.Int())
}

var groupName = getGroupName()

// testAccCheckGroupDestroyed asserts that every managed group is HARD deleted
// (Metabase groups, unlike users, are not soft-deleted): GET must return 404.
func testAccCheckGroupDestroyed(s *terraform.State) error {
	repo := repositories.NewPermissionGroupRepository(newTestMetabaseClient())
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "metabase_permission_group" {
			continue
		}
		_, err := repo.Get(context.Background(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("group %s still exists after destroy; expected a hard delete", rs.Primary.ID)
		}
		var notFound *metabase.NotFoundError
		if !errors.As(err, &notFound) {
			return fmt.Errorf("unexpected error checking destroyed group %s: %w", rs.Primary.ID, err)
		}
	}
	return nil
}

func TestAccPermissionGroupResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroyed,
		Steps: []resource.TestStep{
			// Create and Read.
			{
				Config: testAccPermissionGroupResourceConfig(groupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("metabase_permission_group.test", "name", groupName),
					resource.TestCheckResourceAttrSet("metabase_permission_group.test", "id"),
				),
			},
			// ImportState.
			{
				ResourceName:      "metabase_permission_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Rename must be in-place (Update), not a replace.
			{
				Config: testAccPermissionGroupResourceConfig("test-group-updated"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("metabase_permission_group.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("metabase_permission_group.test", "name", "test-group-updated"),
				),
			},
		},
	})
}

// TestAccPermissionGroupResource_duplicateNameErrors asserts that creating a group
// whose name already exists fails with a `terraform import` hint (no silent adopt).
func TestAccPermissionGroupResource_duplicateNameErrors(t *testing.T) {
	name := getGroupName()
	var createdID int

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Pre-create the group out-of-band.
				PreConfig: func() {
					repo := repositories.NewPermissionGroupRepository(newTestMetabaseClient())
					g, err := repo.Create(context.Background(), name)
					if err != nil {
						t.Fatalf("precondition: create failed: %s", err)
					}
					createdID = g.Id
				},
				Config:      testAccPermissionGroupResourceConfig(name),
				ExpectError: regexp.MustCompile("already exists"),
			},
		},
	})

	// Cleanup the out-of-band group.
	repo := repositories.NewPermissionGroupRepository(newTestMetabaseClient())
	_ = repo.Delete(context.Background(), fmt.Sprintf("%d", createdID))
}

// TestAccPermissionGroupResource_recreateWhenDeletedOutOfBand asserts that if a
// managed group is hard-deleted out-of-band, the next apply recreates it (Read
// removes it from state instead of erroring).
func TestAccPermissionGroupResource_recreateWhenDeletedOutOfBand(t *testing.T) {
	name := getGroupName()
	var firstID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionGroupResourceConfig(name),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						firstID = s.RootModule().Resources["metabase_permission_group.test"].Primary.ID
						return nil
					},
				),
			},
			{
				// Delete the group out-of-band, then re-apply the same config.
				PreConfig: func() {
					repo := repositories.NewPermissionGroupRepository(newTestMetabaseClient())
					if err := repo.Delete(context.Background(), firstID); err != nil {
						t.Fatalf("out-of-band delete failed: %s", err)
					}
				},
				Config: testAccPermissionGroupResourceConfig(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("metabase_permission_group.test", "id"),
					func(s *terraform.State) error {
						newID := s.RootModule().Resources["metabase_permission_group.test"].Primary.ID
						if newID == firstID {
							return fmt.Errorf("expected a recreated group with a new id, still %s", newID)
						}
						return nil
					},
				),
			},
		},
	})
}

func testAccPermissionGroupResourceConfig(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "metabase_permission_group" "test" {
  name = "%s"
}
`, name)
}
