resource "permify_schema" "heredoc" {
    tenant_id = "test"
    schema = <<EOF
    entity User {
        field name: string
    }
    EOF
}

resource "permify_schema" "file" {
    tenant_id = "test"
    schema = file("schema.txt")
}