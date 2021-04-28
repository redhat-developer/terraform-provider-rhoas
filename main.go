package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"redhat.com/rhoas/rhoas-terraform-provider/m/rhoas"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: rhoas.Provider,
	})
}
