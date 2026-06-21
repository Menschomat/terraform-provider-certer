resource "certer_api_key" "web_client" {
  description     = "web-client-token"
  allowed_domains = ["example.com"]
  admin           = false
}

