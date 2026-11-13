data "unraid_system_time" "this" {}

output "time_zone" {
  value = data.unraid_system_time.this.time_zone
}
