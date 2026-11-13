terraform {
  required_providers {
    unraid = {
      source = "jriberg/unraid"
    }
  }
}

# Prefer UNRAID_ENDPOINT and UNRAID_API_KEY environment variables over hardcoding the key.
provider "unraid" {
  endpoint = "https://tower.local/graphql"
}
