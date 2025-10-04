package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccBundlesResource(t *testing.T) {
	resourceName := "permify_bundles.test"

	providerConfig := initPermify(t)

	// Create a tenant first since bundles require a tenant
	tenantConfig := providerConfig + `
resource "permify_tenant" "test" {
  id = "test-tenant"
  name = "Test Tenant"
}
`

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccBundlesResourceConfig(tenantConfig, "test-tenant", testBundlesDefinition),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", "test-tenant"),
					resource.TestCheckResourceAttr(resourceName, "tenant_id", "test-tenant"),
					resource.TestCheckResourceAttr(resourceName, "bundles.#", "2"),
				),
			},
			// ImportState testing
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: false, // Disabled because import only sets ID, not full state
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccBundlesResourceWithDifferentTenant(t *testing.T) {
	providerConfig := initPermify(t)

	// Create two different tenants
	tenantConfig := providerConfig + `
resource "permify_tenant" "test1" {
  id = "tenant-1"
  name = "Tenant 1"
}

resource "permify_tenant" "test2" {
  id = "tenant-2"
  name = "Tenant 2"
}
`

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create bundles for first tenant
			{
				Config: testAccBundlesResourceConfig(tenantConfig, "tenant-1", testBundlesDefinition),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("permify_bundles.test", "id", "tenant-1"),
					resource.TestCheckResourceAttr("permify_bundles.test", "tenant_id", "tenant-1"),
					resource.TestCheckResourceAttr("permify_bundles.test", "bundles.#", "2"),
				),
			},
		},
	})
}

func TestAccBundlesResourceSingleBundle(t *testing.T) {
	resourceName := "permify_bundles.test"

	providerConfig := initPermify(t)

	// Create a tenant first since bundles require a tenant
	tenantConfig := providerConfig + `
resource "permify_tenant" "test" {
  id = "test-tenant"
  name = "Test Tenant"
}
`

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing with single bundle
			{
				Config: testAccBundlesResourceConfig(tenantConfig, "test-tenant", singleBundleDefinition),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", "test-tenant"),
					resource.TestCheckResourceAttr(resourceName, "tenant_id", "test-tenant"),
					resource.TestCheckResourceAttr(resourceName, "bundles.#", "1"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccBundlesResourceConfig(providerConfig string, tenantID string, bundles string) string {
	return providerConfig + fmt.Sprintf(`
resource "permify_bundles" "test" {
  tenant_id = %[1]q
  %[2]s
}
`, tenantID, bundles)
}

// Test bundle definitions
const testBundlesDefinition = `
bundles = [
  {
    name = "user_bundle"
    arguments = ["user_id", "organization_id"]
    operations = [
      {
        relationships_write = [
          "organization:$${organization_id}#member@user:$${user_id}",
          "organization:$${organization_id}#admin@user:$${user_id}"
        ]
        relationships_delete = []
        attributes_write = []
        attributes_delete = []
      }
    ]
  },
  {
    name = "permission_bundle"
    arguments = ["resource_id"]
    operations = [
      {
        relationships_write = [
          "resource:$${resource_id}#owner@user:$${user_id}"
        ]
        relationships_delete = []
        attributes_write = []
        attributes_delete = []
      }
    ]
  }
]
`

const singleBundleDefinition = `
bundles = [
  {
    name = "simple_bundle"
    arguments = ["user_id"]
    operations = [
      {
        relationships_write = [
          "user:$${user_id}#active@true"
        ]
        relationships_delete = []
        attributes_write = []
        attributes_delete = []
      }
    ]
  }
]
`
