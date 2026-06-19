resource "certcentral_api_key" "web_client" {
  name            = "web-client"
  allowed_domains = ["example.com"]
  admin           = false
}

