terraform {
  required_providers {
    spire = {
      source = "terraform.local/rrey/spire"
    }
  }
}

provider "spire" {
  # Configuration options
}

resource "spire_entry" "test" {
  parent_id = {
    trust_domain = "example.org"
    path         = "/some/path"
  }
  spiffe_id = {
    trust_domain = "example.org"
    path         = "/some/service2"
  }
  selectors = [{
    type  = "unix"
    value = "uid:501"
  }]
}
