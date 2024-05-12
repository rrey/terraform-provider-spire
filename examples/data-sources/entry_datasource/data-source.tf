data "spire_entry" "test" {
  spiffe_id = {
    trust_domain = "example.org"
    path         = "/some/datasource-test"
  }
}