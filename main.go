package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/opetch/terraform-provider-aws/aws"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: aws.Provider})
}
