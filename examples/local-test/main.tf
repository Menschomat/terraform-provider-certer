terraform {
  required_providers {
    certer = {
      source  = "gitea.dmz.k8s.menscho.space/m0space/certer"
      version = "1.0.0"
    }
  }
}

provider "certer" {
  address = "http://localhost:8080"
  token   = "your_admin_api_token"
}

resource "certer_certificate" "example" {
  primary = "example.com"
  sans    = [
    "*.example.com",
    "www.example.com"
  ]
}
