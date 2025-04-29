// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func getGroupName() string {
	return fmt.Sprintf("Test group %d", rand.Int())
}

var groupName = getGroupName()

func TestAccPermissionGroupResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccPermissionGroupResourceConfig(groupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("metabase_permission_group.test", "name", groupName),
					resource.TestCheckResourceAttrSet("metabase_permission_group.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "metabase_permission_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccPermissionGroupResourceConfig("test-group-updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("metabase_permission_group.test", "name", "test-group-updated"),
				),
			},
			// Delete testing automatically occurs in TestCase
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
