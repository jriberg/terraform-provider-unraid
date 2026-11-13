data "unraid_shares" "all" {}

output "share_names" {
  value = [for s in data.unraid_shares.all.shares : s.name]
}
