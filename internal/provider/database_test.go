// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"testing"

	"github.com/csp33/terraform-provider-metabase/sdk/metabase"
	"github.com/csp33/terraform-provider-metabase/sdk/metabase/repositories"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// These tests target the local docker "sample-db" postgres (see docker-compose).

// testAccCheckDatabaseDestroyed asserts databases are HARD deleted: GET returns 404.
func testAccCheckDatabaseDestroyed(s *terraform.State) error {
	repo := repositories.NewDatabaseRepository(newTestMetabaseClient())
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "metabase_database" {
			continue
		}
		_, err := repo.Get(context.Background(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("database %s still exists after destroy", rs.Primary.ID)
		}
		var notFound *metabase.NotFoundError
		if !errors.As(err, &notFound) {
			return fmt.Errorf("unexpected error checking destroyed database %s: %w", rs.Primary.ID, err)
		}
	}
	return nil
}

func TestAccDatabaseResource(t *testing.T) {
	name := fmt.Sprintf("Test database %d", rand.Int())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroyed,
		Steps: []resource.TestStep{
			// Create and Read (Metabase tests the connection). Idempotency despite the
			// redacted password is checked implicitly by the framework.
			{
				Config: testAccDatabaseResourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("metabase_database.test", "name", name),
					resource.TestCheckResourceAttr("metabase_database.test", "engine", "postgres"),
					resource.TestCheckResourceAttrSet("metabase_database.test", "id"),
				),
			},
			// ImportState. details is not importable (Metabase redacts secrets), so it
			// is ignored on verify.
			{
				ResourceName:            "metabase_database.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"details", "deletion_protection"},
			},
			// Rename in-place (not a replace).
			{
				Config: testAccDatabaseResourceConfig(name + "-renamed"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("metabase_database.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.TestCheckResourceAttr("metabase_database.test", "name", name+"-renamed"),
			},
		},
	})
}

func testAccDatabaseResourceConfig(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "metabase_database" "test" {
  name                = "%s"
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
`, name)
}
