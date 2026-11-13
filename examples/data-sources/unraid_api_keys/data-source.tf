data "unraid_api_keys" "all" {}

output "admin_keys" {
  value = [for k in data.unraid_api_keys.all.api_keys : k.name if contains(k.roles, "ADMIN")]
}
