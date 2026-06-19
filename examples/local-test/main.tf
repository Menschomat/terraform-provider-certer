terraform {
  required_providers {
    certcentral = {
      source  = "gitea.dmz.k8s.menscho.space/m0space/certcentral"
      version = "1.0.0"
    }
  }
}

provider "certcentral" {
  address = "http://localhost:8080"
  token   = "your_admin_api_token"
}

resource "certcentral_certificate" "example" {
  primary = "example.com"
  sans    = [
    "*.example.com",
    "www.example.com"
  ]
}
