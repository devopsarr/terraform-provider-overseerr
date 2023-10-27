package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSettingsResource(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Unauthorized Create
			{
				Config:      testAccSettingsResourceConfig("overseerr") + testUnauthorizedProvider,
				ExpectError: regexp.MustCompile("Client Error"),
			},
			// Create and Read testing
			{
				Config: testAccSettingsResourceConfig("overseerr"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("overseerr_settings.test", "application_title", "overseerr"),
					resource.TestCheckResourceAttrSet("overseerr_settings.test", "id"),
				),
			},
			// Unauthorized Read
			{
				Config:      testAccSettingsResourceConfig("overseerr") + testUnauthorizedProvider,
				ExpectError: regexp.MustCompile("Client Error"),
			},
			// Update and Read testing
			{
				Config: testAccSettingsResourceConfig("overseerrTest"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("overseerr_settings.test", "application_title", "overseerrTest"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "overseerr_settings.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccSettingsResourceConfig(name string) string {
	return fmt.Sprintf(`
	resource "overseerr_settings" "test" {
		application_title = %s
	}`, name)
}
