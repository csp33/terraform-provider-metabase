// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func getUserEmail() string {
	return fmt.Sprintf("test%d@test.com", rand.Int())
}

var email = getUserEmail()

func TestAccUserResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccUserResourceConfig(email, "Test", "User"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("metabase_user.test", "email", email),
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
				Config: testAccUserResourceConfig(email, "Updated", "User"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("metabase_user.test", "first_name", "Updated"),
				),
			},
			// Deactivate testing
			{
				Config: testAccUserResourceConfigDeactivated(email, "Updated", "User"),
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
