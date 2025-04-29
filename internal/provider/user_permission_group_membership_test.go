// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserPermissionGroupMembershipResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
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
			// ImportState testing
			{
				ResourceName:      "metabase_user_permission_group_membership.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Delete testing automatically occurs in TestCase
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
  name = "membership-test-group"
}

resource "metabase_user_permission_group_membership" "test" {
  user_id             = metabase_user.test.id
  permission_group_id = metabase_permission_group.test.id
}
`, getUserEmail())
}
