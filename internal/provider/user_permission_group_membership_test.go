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
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// testAccCheckMembershipDestroyed asserts the membership no longer exists after destroy.
func testAccCheckMembershipDestroyed(s *terraform.State) error {
	repo := repositories.NewUserPermissionGroupMembershipRepository(newTestMetabaseClient())
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "metabase_user_permission_group_membership" {
			continue
		}
		_, err := repo.Get(context.Background(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("membership %s still exists after destroy", rs.Primary.ID)
		}
		var notFound *metabase.NotFoundError
		if !errors.As(err, &notFound) {
			return fmt.Errorf("unexpected error checking destroyed membership %s: %w", rs.Primary.ID, err)
		}
	}
	return nil
}

func TestAccUserPermissionGroupMembershipResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckMembershipDestroyed,
		Steps: []resource.TestStep{
			// Create and Read (user + group + membership in one apply).
			{
				Config: testAccUserPermissionGroupMembershipResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("metabase_user_permission_group_membership.test", "id"),
					resource.TestCheckResourceAttrPair(
						"metabase_user_permission_group_membership.test", "user_id",
						"metabase_user.test", "id",
					),
					resource.TestCheckResourceAttrPair(
						"metabase_user_permission_group_membership.test", "permission_group_id",
						"metabase_permission_group.test", "id",
					),
				),
			},
			// ImportState.
			{
				ResourceName:      "metabase_user_permission_group_membership.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccUserPermissionGroupMembershipResource_duplicateErrors asserts that a
// second membership for the same (user, group) fails with a clean import-hint
// error instead of surfacing Metabase's HTTP 500.
func TestAccUserPermissionGroupMembershipResource_duplicateErrors(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckMembershipDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPermissionGroupMembershipResourceConfig(),
			},
			{
				// Add a second membership pointing at the same user+group.
				Config:      testAccUserPermissionGroupMembershipResourceConfig() + testAccDuplicateMembershipConfig(),
				ExpectError: regexp.MustCompile("already a member"),
			},
		},
	})
}

// TestAccUserPermissionGroupMembershipResource_recreateWhenGroupDeleted asserts
// that hard-deleting the group out-of-band (which cascade-removes the membership)
// is reconciled by recreating both, instead of erroring on refresh.
func TestAccUserPermissionGroupMembershipResource_recreateWhenGroupDeleted(t *testing.T) {
	var groupID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckMembershipDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPermissionGroupMembershipResourceConfig(),
				Check: func(s *terraform.State) error {
					groupID = s.RootModule().Resources["metabase_permission_group.test"].Primary.ID
					return nil
				},
			},
			{
				// Delete the group out-of-band (cascades the membership), then re-apply.
				PreConfig: func() {
					repo := repositories.NewPermissionGroupRepository(newTestMetabaseClient())
					if err := repo.Delete(context.Background(), groupID); err != nil {
						t.Fatalf("out-of-band group delete failed: %s", err)
					}
				},
				Config: testAccUserPermissionGroupMembershipResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("metabase_user_permission_group_membership.test", "id"),
					resource.TestCheckResourceAttrSet("metabase_permission_group.test", "id"),
				),
			},
		},
	})
}

// TestAccUserPermissionGroupMembershipResource_survivesUserDeactivation asserts
// that deactivating a user does NOT remove their memberships (Metabase keeps them),
// so a managed membership does not drift. The user is reactivated at the end so the
// subsequent destroy is clean.
func TestAccUserPermissionGroupMembershipResource_survivesUserDeactivation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckMembershipDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPermissionGroupMembershipResourceConfig(),
				Check: func(s *terraform.State) error {
					userID := s.RootModule().Resources["metabase_user.test"].Primary.ID
					membershipID := s.RootModule().Resources["metabase_user_permission_group_membership.test"].Primary.ID
					userRepo := repositories.NewUserRepository(newTestMetabaseClient())
					memRepo := repositories.NewUserPermissionGroupMembershipRepository(newTestMetabaseClient())

					// Deactivate the user out-of-band.
					if _, err := userRepo.Update(context.Background(), userID, nil, nil, nil, boolPtr(false)); err != nil {
						return fmt.Errorf("deactivate failed: %w", err)
					}
					// Membership must still exist.
					if _, err := memRepo.Get(context.Background(), membershipID); err != nil {
						return fmt.Errorf("membership %s vanished after user deactivation: %w", membershipID, err)
					}
					// Reactivate so the test's destroy phase is clean.
					if _, err := userRepo.Update(context.Background(), userID, nil, nil, nil, boolPtr(true)); err != nil {
						return fmt.Errorf("reactivate failed: %w", err)
					}
					return nil
				},
			},
		},
	})
}

func testAccUserPermissionGroupMembershipResourceConfig() string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "metabase_user" "test" {
  email      = "%s"
  first_name = "Membership"
  last_name  = "Test"
}

resource "metabase_permission_group" "test" {
  name = "membership-test-group-%d"
}

resource "metabase_user_permission_group_membership" "test" {
  user_id             = metabase_user.test.id
  permission_group_id = metabase_permission_group.test.id
}
`, getUserEmail(), rand.Int())
}

func testAccDuplicateMembershipConfig() string {
	return `
resource "metabase_user_permission_group_membership" "dup" {
  user_id             = metabase_user.test.id
  permission_group_id = metabase_permission_group.test.id
}
`
}
