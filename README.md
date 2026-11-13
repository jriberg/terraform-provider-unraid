# terraform-provider-unraid

Terraform/OpenTofu provider for [Unraid](https://unraid.net), built on the official Unraid GraphQL API
(`definition.graphql` in this repo is the schema the client is generated from).

User documentation: [`docs/`](docs/index.md) (published to the Terraform and OpenTofu registries as `jriberg/unraid`).

## Status

| Kind | Implemented |
|---|---|
| Data sources | `unraid_vars`, `unraid_server`, `unraid_system_time`, `unraid_ups_configuration`, `unraid_settings`, `unraid_shares`, `unraid_array`, `unraid_disks`, `unraid_docker_containers`, `unraid_vms`, `unraid_api_keys`, `unraid_rclone_remotes` |
| Resources | planned: server identity, SSH, system time, UPS, API keys, rclone remotes, unified settings |

Not possible via the API (no mutations exist): shares, Docker container/VM creation, network, users.

## Development

Requirements: Go (see `go.mod`), OpenTofu or Terraform.

```shell
go generate ./...          # regenerate GraphQL client (genqlient) and docs (tfplugindocs)
go vet ./...
TF_ACC_TERRAFORM_PATH="$(which tofu)" TF_ACC_PROVIDER_NAMESPACE=jriberg go test ./...
go tool tfplugindocs validate
```

Tests run the real provider through tofu against a fake GraphQL server; no Unraid server is needed.

Never edit `internal/client/generated.go` or `docs/` by hand: change `internal/client/operations.graphql`,
schema descriptions, `examples/` or `templates/` and regenerate.

### Trying it against a real server

```shell
go install .
cat > ~/.tofurc <<'RC'
provider_installation {
  dev_overrides {
    "jriberg/unraid" = "/Users/<you>/go/bin"
  }
  direct {}
}
RC
export UNRAID_ENDPOINT="https://tower.local/graphql" UNRAID_API_KEY="..."
cd examples/data-sources/unraid_shares && tofu plan
```

(`~/.terraformrc` with the same content for Terraform, using `registry.terraform.io/jriberg/unraid`.)

## Releasing

1. Create a GPG key and add secrets `GPG_PRIVATE_KEY` and `PASSPHRASE` to the GitHub repo.
2. Push a tag `vX.Y.Z`; `.github/workflows/release.yaml` builds and signs with goreleaser.
3. Terraform Registry: sign in with GitHub, add the public GPG key, publish `jriberg/terraform-provider-unraid`.
4. OpenTofu Registry: submit the provider and GPG key via the `opentofu/registry` repository's issue forms.
