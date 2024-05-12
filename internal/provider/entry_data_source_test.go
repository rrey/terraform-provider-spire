// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSpireEntryDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + testAccSpireEntryDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.spire_entry.test", "parent_id.trust_domain", "example.org"),
					resource.TestCheckResourceAttr("data.spire_entry.test", "parent_id.path", "/some/path"),
					resource.TestCheckResourceAttr("data.spire_entry.test", "spiffe_id.trust_domain", "example.org"),
					resource.TestCheckResourceAttr("data.spire_entry.test", "spiffe_id.path", "/some/datasource-test"),
					resource.TestCheckResourceAttr("data.spire_entry.test", "selectors.0.type", "unix"),
					resource.TestCheckResourceAttr("data.spire_entry.test", "selectors.0.value", "uid:501"),
				),
			},
		},
	})
}

const testAccSpireEntryDataSourceConfig = `
data "spire_entry" "test" {
	spiffe_id = {
	  trust_domain = "example.org"
	  path = "/some/datasource-test"
	}
   }
`
