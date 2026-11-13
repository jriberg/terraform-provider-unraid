data "unraid_vms" "all" {}

output "running_vms" {
  value = [for v in data.unraid_vms.all.vms : v.name if v.state == "RUNNING"]
}
