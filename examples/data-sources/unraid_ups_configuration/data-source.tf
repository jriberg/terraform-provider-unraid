data "unraid_ups_configuration" "this" {}

output "ups_shutdown_battery_level" {
  value = data.unraid_ups_configuration.this.battery_level
}
