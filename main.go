package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/jriberg/terraform-provider-unraid/internal/provider"
)

//go:generate go tool tfplugindocs generate --provider-name unraid

// version is set by goreleaser at build time.
var version = "dev"

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "run the provider with support for debuggers like delve")
	flag.Parse()

	err := providerserver.Serve(context.Background(), provider.New(version), providerserver.ServeOpts{
		Address: "registry.terraform.io/jriberg/unraid",
		Debug:   debug,
	})
	if err != nil {
		log.Fatal(err)
	}
}
