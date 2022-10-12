package main

import (
	"flag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/redhat-developer/terraform-provider-rhoas/rhoas"
)

func main() {

	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuging")
	flag.Parse()

	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: rhoas.Provider,
		ProviderAddr: "registry.terraform.io/redhat-developer/rhoas",
		Debug:        debug,
	})
}
