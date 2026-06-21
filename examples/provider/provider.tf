terraform {
  required_providers {
    certer = {
      source  = "Menschomat/certer"
      version = "~> 1.0"
    }
  }
}

provider "certer" {
  address = "http://localhost:8080"
  token   = "your_admin_api_token"
}
