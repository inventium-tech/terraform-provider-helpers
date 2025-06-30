<p align="center">
  <img src="./assets/provider_logo.svg" width="200" alt="logo"/>
</p>

---

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/inventium-tech/terraform-provider-helpers?style=for-the-badge&logo=go)
![Terraform version](https://img.shields.io/badge/terraform-v1.11-black?style=for-the-badge&logo=terraform&color=844FBA)
![Terraform Provider Downloads](https://img.shields.io/terraform/provider/dw/5756?style=for-the-badge&logo=terraform&color=844FBA)
![GitHub License](https://img.shields.io/github/license/inventium-tech/terraform-provider-helpers?style=for-the-badge)

[![GitHub Pre-Release](https://img.shields.io/github/v/release/inventium-tech/terraform-provider-postgresql?include_prereleases&sort=semver&display_name=release&style=for-the-badge&logo=semanticrelease&label=pre-release&color=orange)](https://github.com/inventium-tech/terraform-provider-postgresql/releases)
[![GitHub Release](https://img.shields.io/github/v/release/inventium-tech/terraform-provider-postgresql?sort=semver&display_name=release&style=for-the-badge&logo=semanticrelease&color=green)](https://github.com/inventium-tech/terraform-provider-helpers/releases/latest)

[![Codecov](https://img.shields.io/codecov/c/github/inventium-tech/terraform-provider-helpers?style=for-the-badge&logo=codecov&label=CodeCov%20Coverage)](https://codecov.io/gh/inventium-tech/terraform-provider-helpers)
[![Sonar](https://img.shields.io/sonar/coverage/inventium-tech_terraform-provider-helpers?server=https%3A%2F%2Fsonarcloud.io&style=for-the-badge&logo=sonarqubecloud&label=SonarCloud%20Coverage)](https://sonarcloud.io/project/overview?id=inventium-tech_terraform-provider-helpers)
![Sonar Quality Gate](https://img.shields.io/sonar/quality_gate/inventium-tech_terraform-provider-helpers?server=https%3A%2F%2Fsonarcloud.io&style=for-the-badge&logo=sonarqubecloud)

<h2>📋 Table of Contents</h2>

<!-- TOC -->
* [🧰 Terraform Provider: Helpers Functions](#-terraform-provider-helpers-functions)
  * [Available Functions](#available-functions)
  * [Example Usage](#example-usage)
<!-- TOC -->

# 🧰 Terraform Provider: Helpers Functions

This is a Terraform Provider only for Helper functions. The main idea is to extend the built-in functionalities of
Terraform with some functions that are not available by default and can be useful in some scenarios.

For detailed examples and documentation check inside the [docs](./docs/index.md) folder or directly in the
Terraform Registry.

## Available Functions

**Collection Functions:**

* [collection_filter](./docs/functions/collection_filter.md): Filter collection of objects.

**Object Functions:**

* [object_set_value](./docs/functions/object_set_value.md): Sets a value in an Object or creates a new key with the value.

**OS Functions:**

* [os_get_env](./docs/functions/os_get_env.md): Read an environment variable.

## Example Usage

```terraformterraform {
  required_providers {
    helpers = {
      source = "inventium-tech/helpers"
    }
  }
}

locals {
  test_object = {
    key1 = "value1"
    key2 = true
    key3 = 3
    key4 = ""
    key5 = null
  }
}

output "write_all_operation" {
  value = {
    test_value_change_on_existing_key = provider::helpers::object_set_value(local.test_object, "key1", "new_value", "write_all")
    test_add_new_key_value            = provider::helpers::object_set_value(local.test_object, "new_key", "new_value", "write_all")
  }
}
```

<a href="https://www.buymeacoffee.com/refucktor" target="_blank">
  <img src="https://cdn.buymeacoffee.com/buttons/v2/default-red.png" alt="Buy Me A Coffee"
    style="height: 60px !important;width: 217px !important;">
</a>
