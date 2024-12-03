package provider

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"os"
	"testing"
)

func TestOsGetEnvFunction(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := os.Setenv("TF_ENV", "dev"); err != nil {
						t.Fatal(err)
					}
				},
				// test function without fallback
				Config: `output "tf_env" { value = provider::helpers::os_get_env("TF_ENV") }`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("tf_env", knownvalue.StringExact("dev")),
				},
			},
			{
				// test function with fallback
				Config: `output "tf_env" { value = provider::helpers::os_get_env("TF_ENV", "testing") }`,
				PreConfig: func() {
					if err := os.Unsetenv("TF_ENV"); err != nil {
						t.Fatal(err)
					}
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("tf_env", knownvalue.StringExact("testing")),
				},
			},
			{
				// test function with empty result
				Config: `output "tf_env" { value = provider::helpers::os_get_env("TF_ENV") }`,
				PreConfig: func() {
					if err := os.Unsetenv("TF_ENV"); err != nil {
						t.Fatal(err)
					}
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("tf_env", knownvalue.StringExact("")),
				},
			},
		},
	})
}
