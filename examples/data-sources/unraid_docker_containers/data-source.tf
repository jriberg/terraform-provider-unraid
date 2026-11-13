data "unraid_docker_containers" "all" {}

output "containers_with_updates" {
  value = [for c in data.unraid_docker_containers.all.containers : c.names[0] if c.is_update_available == true]
}
