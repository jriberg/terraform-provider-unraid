data "unraid_settings" "this" {}

locals {
  settings = jsondecode(data.unraid_settings.this.values_json)
}

output "api_sandbox_enabled" {
  value = nonsensitive(try(local.settings.api.sandbox, null))
}
