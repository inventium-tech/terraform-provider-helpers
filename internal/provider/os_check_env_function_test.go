package provider

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"os"
	"testing"
)

func TestOsCheckEnvFunction(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// test when the environment variable exists and is not empty (should return true)
				PreConfig: func() {
					if err := os.Setenv("TF_ENV_CHECK", "value"); err != nil {
						t.Fatal(err)
					}
				},
				Config: `output "tf_env_check" { value = provider::helpers::os_check_env("TF_ENV_CHECK") }`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("tf_env_check", knownvalue.Bool(true)),
				},
			},
			{
				// test when the environment variable exists but is empty, with strict=true (default) (should return false)
				PreConfig: func() {
					if err := os.Setenv("TF_ENV_CHECK", ""); err != nil {
						t.Fatal(err)
					}
				},
				Config: `output "tf_env_check" { value = provider::helpers::os_check_env("TF_ENV_CHECK") }`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("tf_env_check", knownvalue.Bool(false)),
				},
			},
			{
				// test when the environment variable exists but is empty, with strict=false (should return true)
				PreConfig: func() {
					if err := os.Setenv("TF_ENV_CHECK", ""); err != nil {
						t.Fatal(err)
					}
				},
				Config: `output "tf_env_check" { value = provider::helpers::os_check_env("TF_ENV_CHECK", false) }`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("tf_env_check", knownvalue.Bool(true)),
				},
			},
			{
				// test when the environment variable doesn't exist (should return false)
				PreConfig: func() {
					if err := os.Unsetenv("TF_ENV_CHECK"); err != nil {
						t.Fatal(err)
					}
				},
				Config: `output "tf_env_check" { value = provider::helpers::os_check_env("TF_ENV_CHECK") }`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("tf_env_check", knownvalue.Bool(false)),
				},
			},
		},
	})
}
