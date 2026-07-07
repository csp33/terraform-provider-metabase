// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"testing"

	"github.com/csp33/terraform-provider-metabase/sdk/metabase"
	"github.com/csp33/terraform-provider-metabase/sdk/metabase/repositories"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func getUserEmail() string {
	return fmt.Sprintf("test%d@test.com", rand.Int())
}

var email = getUserEmail()
var emailChanged = getUserEmail()

func boolPtr(b bool) *bool { return &b }

func newTestMetabaseClient() *metabase.MetabaseAPIClient {
	return metabase.NewMetabaseAPIClient(os.Getenv("METABASE_HOST"), os.Getenv("METABASE_API_KEY"))
}

// testAccCheckUserDeactivated verifies that every managed user still EXISTS in
// Metabase after destroy but is DEACTIVATED (Metabase never hard-deletes users).
func testAccCheckUserDeactivated(s *terraform.State) error {
	repo := repositories.NewUserRepository(newTestMetabaseClient())
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "metabase_user" {
			continue
		}
		u, err := repo.Get(context.Background(), rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("user %s should still exist after destroy (soft-delete) but Get failed: %w", rs.Primary.ID, err)
		}
		if u.IsActive {
			return fmt.Errorf("user %s is still active after destroy; expected it to be deactivated", rs.Primary.ID)
		}
	}
	return nil
}

// TestAccUserResource covers the full lifecycle: create, import, rename,
// email change (must be in-place, not a replace), deactivate and reactivate.
// CheckDestroy asserts that destroy deactivates the user rather than leaving it active.
func TestAccUserResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckUserDeactivated,
		Steps: []resource.TestStep{
			// Create and Read.
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
			// ImportState.
			{
				ResourceName:      "metabase_user.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Rename (in-place).
			{
				Config: testAccUserResourceConfig(email, "Updated", "User"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("metabase_user.test", "first_name", "Updated"),
				),
			},
			// Email change must be applied IN-PLACE (Update), not a replace.
			{
				Config: testAccUserResourceConfig(emailChanged, "Updated", "User"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("metabase_user.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("metabase_user.test", "email", emailChanged),
				),
			},
			// Deactivate (in-place).
			{
				Config: testAccUserResourceConfigDeactivated(emailChanged, "Updated", "User"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("metabase_user.test", "is_active", "false"),
				),
			},
			// Reactivate (in-place).
			{
				Config: testAccUserResourceConfig(emailChanged, "Updated", "User"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("metabase_user.test", "is_active", "true"),
				),
			},
		},
	})
}

// TestAccUserResource_lowercaseEmailValidation asserts a mixed-case email is
// rejected at plan time (before any API call), so no orphan user is created.
func TestAccUserResource_lowercaseEmailValidation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccUserResourceConfig("MixedCase@Example.COM", "Mixed", "Case"),
				ExpectError: regexp.MustCompile("(?i)must be lowercase"),
			},
		},
	})
}

// TestAccUserResource_existingEmailErrorsWithImportHint asserts that creating a
// resource whose email already belongs to an existing (here, deactivated) user
// fails with an actionable error pointing to `terraform import`, instead of
// silently adopting or mutating that account.
func TestAccUserResource_existingEmailErrorsWithImportHint(t *testing.T) {
	reuseEmail := getUserEmail()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Pre-create the user out-of-band and deactivate it.
				PreConfig: func() {
					repo := repositories.NewUserRepository(newTestMetabaseClient())
					u, err := repo.Create(context.Background(), reuseEmail, "Pre", "Existing")
					if err != nil {
						t.Fatalf("precondition: create failed: %s", err)
					}
					if _, err := repo.Update(context.Background(), strconv.Itoa(u.Id), nil, nil, nil, boolPtr(false)); err != nil {
						t.Fatalf("precondition: deactivate failed: %s", err)
					}
				},
				Config:      testAccUserResourceConfigNamed("dup", reuseEmail, "Dup", "Licate"),
				ExpectError: regexp.MustCompile("already exists"),
			},
		},
	})
}

// TestAccUserResource_reactivateViaImport asserts the idiomatic reactivation flow:
// import an existing (deactivated) user, then apply is_active = true, which
// reactivates it in-place (same id) rather than creating a new user.
func TestAccUserResource_reactivateViaImport(t *testing.T) {
	reuseEmail := getUserEmail()
	var preID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckUserDeactivated,
		Steps: []resource.TestStep{
			// Import the pre-created (deactivated) user into state.
			{
				PreConfig: func() {
					repo := repositories.NewUserRepository(newTestMetabaseClient())
					u, err := repo.Create(context.Background(), reuseEmail, "Imp", "Ported")
					if err != nil {
						t.Fatalf("precondition: create failed: %s", err)
					}
					preID = strconv.Itoa(u.Id)
					if _, err := repo.Update(context.Background(), preID, nil, nil, nil, boolPtr(false)); err != nil {
						t.Fatalf("precondition: deactivate failed: %s", err)
					}
				},
				Config:             testAccUserResourceConfigNamed("imp", reuseEmail, "Imp", "Ported"),
				ResourceName:       "metabase_user.imp",
				ImportState:        true,
				ImportStatePersist: true,
				ImportStateIdFunc:  func(*terraform.State) (string, error) { return preID, nil },
			},
			// Applying the config (is_active defaults to true) reactivates in-place.
			{
				Config: testAccUserResourceConfigNamed("imp", reuseEmail, "Imp", "Ported"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("metabase_user.imp", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("metabase_user.imp", "is_active", "true"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["metabase_user.imp"]
						if !ok {
							return fmt.Errorf("metabase_user.imp not found in state")
						}
						if rs.Primary.ID != preID {
							return fmt.Errorf("expected reactivated user id %s, but managed id is %s", preID, rs.Primary.ID)
						}
						return nil
					},
				),
			},
		},
	})
}

// TestUserReactivateIdempotent asserts that reactivating an already-active user is
// a no-op success (Metabase returns 400 "Not able to reactivate an active user").
func TestUserReactivateIdempotent(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("acceptance test; set TF_ACC=1")
	}
	testAccPreCheck(t)

	repo := repositories.NewUserRepository(newTestMetabaseClient())
	ctx := context.Background()

	created, err := repo.Create(ctx, getUserEmail(), "Reactivate", "Idempotent")
	if err != nil {
		t.Fatalf("create failed: %s", err)
	}
	id := strconv.Itoa(created.Id)

	// The user is active; reactivating it (Update with is_active=true) must succeed.
	if _, err := repo.Update(ctx, id, nil, nil, nil, boolPtr(true)); err != nil {
		t.Fatalf("reactivating an active user should be idempotent, got: %s", err)
	}

	// Cleanup: deactivate.
	if _, err := repo.Update(ctx, id, nil, nil, nil, boolPtr(false)); err != nil {
		t.Fatalf("cleanup deactivate failed: %s", err)
	}
}

func testAccUserResourceConfig(email, firstName, lastName string) string {
	return testAccUserResourceConfigNamed("test", email, firstName, lastName)
}

func testAccUserResourceConfigNamed(name, email, firstName, lastName string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "metabase_user" "%s" {
  email      = "%s"
  first_name = "%s"
  last_name  = "%s"
}
`, name, email, firstName, lastName)
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
