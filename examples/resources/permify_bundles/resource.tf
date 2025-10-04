resource "permify_bundles" "test" {
    tenant_id = "test"
    bundles = [
        {
            name = "organization_created"
            arguments = [
                "creatorID",
                "organizationID",
            ]
            operations = [
                {
                    relationships_write = [
                        "organization:{{.organizationID}}#admin@user:{{.creatorID}}",
                        "organization:{{.organizationID}}#manager@user:{{.creatorID}}"
                    ],
                    attributes_write = [
                        "organization:{{.organizationID}}$public|boolean:false"
                    ]
                }
            ]
        }
    ]
}