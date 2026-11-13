data "unraid_server" "this" {}

output "server_name" {
  value = data.unraid_server.this.name
}
