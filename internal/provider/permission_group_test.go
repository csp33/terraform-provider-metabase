// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPermissionGroupResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccPermissionGroupResourceConfig("test-group"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("metabase_permission_group.test", "name", "test-group"),
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
