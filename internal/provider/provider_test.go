package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"helpers": providerserver.NewProtocol6WithError(NewProvider("test")()),
}

func testWrap() {
	x := map[string]interface{}{
		"keyOne":   "valueOne",
		"keyTwo":   2,
		"keyThree": true,
	}
	var qwe []interface{}

	for key, value := range x {
		qwe = append(qwe, fmt.Sprintf("%v = %v", key, value))
	}
	str := fmt.Sprintf("%v", x)
	fmt.Println(str)
}
