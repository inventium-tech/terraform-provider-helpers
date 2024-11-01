package main

import (
	"context"
	"flag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"terraform-provider-helpers/internal/provider"
)

// Run "go generate" to format example terraform files and generate the docs for the registry/website

// If you do not have Terraform installed, you can remove the formatting command, but its suggested to
// ensure the documentation is formatted properly.
//go:generate terraform fmt -recursive ./examples/

// Run the docs generation tool, check its repository for more information on how it works and how docs
// can be customized.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate -provider-name helpers

var (
	// these will be set by the goreleaser configuration
	// to appropriate values for the compiled binary.
	version string = "dev"

	// goreleaser can pass other information to the main package, such as the specific commit
	// https://goreleaser.com/cookbooks/using-main.version/
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		// NOTE: for local development, make sure to add "-local" to "inventium-tech"
		// this will make sure that the provider is built and ran locally
		// DO NOT FORGET TO REMOVE IT BEFORE RELEASING
		Address: "registry.terraform.io/inventium-tech/helpers",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), provider.NewProvider(version), opts)

	if err != nil {
		tflog.Trace(context.Background(), err.Error())
		log.Fatal(err.Error())
	}
}
