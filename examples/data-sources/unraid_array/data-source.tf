data "unraid_array" "this" {}

output "array_state" {
  value = data.unraid_array.this.state
}

output "failing_disks" {
  value = [for d in data.unraid_array.this.disks : d.name if d.status != "DISK_OK"]
}
