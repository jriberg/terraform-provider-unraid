data "unraid_rclone_remotes" "all" {}

output "remote_names" {
  value = [for r in data.unraid_rclone_remotes.all.remotes : r.name]
}
