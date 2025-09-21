package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTenantResource(t *testing.T) {
	var createdAt string
	resourceName := "permify_tenant.test"

	providerConfig := initPermify(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccTenantResourceConfig(providerConfig, "one", "uno"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", "one"),
					resource.TestCheckResourceAttr(resourceName, "name", "uno"),
					resource.TestCheckResourceAttrWith(resourceName, "created_at", func(value string) error {
						createdAt = value
						_, err := time.Parse(time.RFC3339, value)
						if err != nil {
							return fmt.Errorf("expected created_at to be a valid RFC3339 date-time, but got %s", value)
						}
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccTenantResourceConfig(providerConfig, "one", "uno"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", "one"),
					resource.TestCheckResourceAttr(resourceName, "name", "uno"),
					resource.TestCheckResourceAttrWith(resourceName, "created_at", func(value string) error {
						if createdAt != value {
							return fmt.Errorf("expected created_at to be %s, but got %s", createdAt, value)
						}
						return nil
					}),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccTenantResourceConfig(providerConfig string, id string, name string) string {
	return providerConfig + fmt.Sprintf(`
resource "permify_tenant" "test" {
  id = %[1]q
  name = %[2]q
}
`, id, name)
}
