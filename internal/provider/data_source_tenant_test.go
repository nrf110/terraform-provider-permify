package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTenantDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTenantDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.permify_tenant.test", "id", "two"),
					resource.TestCheckResourceAttr("data.permify_tenant.test", "name", "dos"),
					resource.TestCheckResourceAttrWith("data.permify_tenant.test", "created_at", func(value string) error {
						_, err := time.Parse(time.RFC3339, value)
						if err != nil {
							return fmt.Errorf("expected created_at to be a valid RFC3339 date-time, but got %s", value)
						}
						return nil
					}),
				),
			},
		},
	})
}

const testAccTenantDataSourceConfig = providerConfig + `
resource "permify_tenant" "test" {
  id = "two"
  name = "dos"
}

data "permify_tenant" "test" {
 id = permify_tenant.test.id
}
`
