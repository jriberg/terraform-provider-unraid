data "unraid_vars" "this" {}

output "unraid_version" {
  value = data.unraid_vars.this.version
}
