resource "certer_api_key" "web_client" {
  description          = "web-client-token"
  allowed_certificates = ["019035a1-7b00-7521-8280-60b6adbf47eb"]
  allowed_teams        = ["019035a1-7b00-7521-8280-60b6adbf47ea"]
  admin                = false
}

