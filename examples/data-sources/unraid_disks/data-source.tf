data "unraid_disks" "all" {}

output "disk_serials" {
  value = { for d in data.unraid_disks.all.disks : d.device => d.serial_num }
}
