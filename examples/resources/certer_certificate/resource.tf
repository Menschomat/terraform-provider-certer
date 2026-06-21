resource "certer_certificate" "example" {
  primary = "example.com"
  team_id = "some-team-uuid"
  sans    = [
    "*.example.com",
    "www.example.com"
  ]
}
