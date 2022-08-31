package main

import (
	"flag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"redhat.com/rhoas/rhoas-terraform-provider/m/rhoas"
)

func main() {

	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuging")
	flag.Parse()

	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: rhoas.Provider,
		Debug:        debug,
	})
}
