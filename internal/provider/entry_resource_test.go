// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSpireEntryResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + testAccSpireEntryResourceConfig("service"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("spire_entry.test", "parent_id.trust_domain", "example.org"),
					resource.TestCheckResourceAttr("spire_entry.test", "parent_id.path", "/some/path"),
					resource.TestCheckResourceAttr("spire_entry.test", "spiffe_id.trust_domain", "example.org"),
					resource.TestCheckResourceAttr("spire_entry.test", "spiffe_id.path", "/some/service"),
					resource.TestCheckResourceAttr("spire_entry.test", "selectors.0.type", "unix"),
					resource.TestCheckResourceAttr("spire_entry.test", "selectors.0.value", "uid:501"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "spire_entry.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccSpireEntryResourceConfig("service2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("spire_entry.test", "spiffe_id.path", "/some/service2"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccSpireEntryResourceConfig(serviceName string) string {
	return fmt.Sprintf(`
	resource "spire_entry" "test" {
		parent_id = {
		  trust_domain = "example.org"
		  path = "/some/path"
		}
		spiffe_id = {
		  trust_domain = "example.org"
		  path = "/some/%[1]s"
		}
		selectors = [{
			type = "unix"
			value = "uid:501"
		}]
	   }
`, serviceName)
}
