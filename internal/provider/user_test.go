// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccUserResourceConfig("test-user@example.com", "Test", "User"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("metabase_user.test", "email", "test-user@example.com"),
					resource.TestCheckResourceAttr("metabase_user.test", "first_name", "Test"),
					resource.TestCheckResourceAttr("metabase_user.test", "last_name", "User"),
					resource.TestCheckResourceAttr("metabase_user.test", "is_active", "true"),
					resource.TestCheckResourceAttrSet("metabase_user.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "metabase_user.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccUserResourceConfig("test-user@example.com", "Updated", "User"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("metabase_user.test", "first_name", "Updated"),
				),
			},
			// Deactivate testing
			{
				Config: testAccUserResourceConfigDeactivated("test-user@example.com", "Updated", "User"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("metabase_user.test", "is_active", "false"),
				),
			},
		},
	})
}

func testAccUserResourceConfig(email, firstName, lastName string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "metabase_user" "test" {
  email      = "%s"
  first_name = "%s"
  last_name  = "%s"
}
`, email, firstName, lastName)
}

func testAccUserResourceConfigDeactivated(email, firstName, lastName string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "metabase_user" "test" {
  email      = "%s"
  first_name = "%s"
  last_name  = "%s"
  is_active  = false
}
`, email, firstName, lastName)
}
