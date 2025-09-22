package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSchemaResource(t *testing.T) {
	var schemaVersion string
	resourceName := "permify_schema.test"

	providerConfig := initPermify(t)

	// Create a tenant first since schema requires a tenant
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
				Config: testAccSchemaResourceConfig(tenantConfig, "test-tenant", testSchemaDefinition),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", "test-tenant"),
					resource.TestCheckResourceAttr(resourceName, "tenant_id", "test-tenant"),
					resource.TestCheckResourceAttr(resourceName, "schema", testSchemaDefinition),
					resource.TestCheckResourceAttrWith(resourceName, "schema_version", func(value string) error {
						schemaVersion = value
						if value == "" {
							return fmt.Errorf("expected schema_version to be set, but got empty string")
						}
						return nil
					}),
				),
			},
			// Update and Read testing
			{
				Config: testAccSchemaResourceConfig(tenantConfig, "test-tenant", updatedSchemaDefinition),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", "test-tenant"),
					resource.TestCheckResourceAttr(resourceName, "tenant_id", "test-tenant"),
					resource.TestCheckResourceAttr(resourceName, "schema", updatedSchemaDefinition),
					resource.TestCheckResourceAttrWith(resourceName, "schema_version", func(value string) error {
						if value == "" {
							return fmt.Errorf("expected schema_version to be set, but got empty string")
						}
						// Schema version should change when schema is updated
						if schemaVersion == value {
							return fmt.Errorf("expected schema_version to change after update, but got same value: %s", value)
						}
						return nil
					}),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccSchemaResourceWithDifferentTenant(t *testing.T) {
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
			// Create schema for first tenant
			{
				Config: testAccSchemaResourceConfig(tenantConfig, "tenant-1", testSchemaDefinition),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("permify_schema.test", "id", "tenant-1"),
					resource.TestCheckResourceAttr("permify_schema.test", "tenant_id", "tenant-1"),
					resource.TestCheckResourceAttr("permify_schema.test", "schema", testSchemaDefinition),
				),
			},
			// Create schema for second tenant
			{
				Config: testAccSchemaResourceConfig(tenantConfig, "tenant-2", testSchemaDefinition),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("permify_schema.test", "id", "tenant-2"),
					resource.TestCheckResourceAttr("permify_schema.test", "tenant_id", "tenant-2"),
					resource.TestCheckResourceAttr("permify_schema.test", "schema", testSchemaDefinition),
				),
			},
		},
	})
}

func testAccSchemaResourceConfig(providerConfig string, tenantID string, schema string) string {
	return providerConfig + fmt.Sprintf(`
resource "permify_schema" "test" {
  tenant_id = %[1]q
  schema = %[2]q
}
`, tenantID, schema)
}

// Test schema definitions
const testSchemaDefinition = `
entity user {}

entity organization {
    relation admin @user
    relation member @user
    action create_repository = admin
    action delete = admin
    action leave = member
}

entity repository {
    relation parent @organization
    relation owner @user
    relation maintainer @user @organization#member
    action push = owner
    action read = owner and maintainer
    action delete = parent.admin
}
`

const updatedSchemaDefinition = `
entity user {}

entity organization {
    relation admin @user
    relation member @user
    relation viewer @user
    action create_repository = admin
    action delete = admin
    action leave = member
    action view = member or viewer
}

entity repository {
    relation parent @organization
    relation owner @user
    relation maintainer @user @organization#member
    relation reader @user @organization#viewer
    action push = owner
    action read = owner and (maintainer or reader)
    action delete = parent.admin
}
`
