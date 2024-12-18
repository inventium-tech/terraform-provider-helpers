<p align="center">
  <img src="./assets/provider_logo.svg" width="200" alt="logo"/>
</p>

---

![Golang][shield-golang]
![Terraform][shield-terraform]

[![🛠️ Build Workflow][badge-gh-action-build]][link-gh-action-build]
[![🔎 MegaLinter][badge-gh-action-megalinter]][link-gh-action-megalinter]
[![❇️ CodeQL][badge-gh-action-codeql]][link-gh-action-codeql]

![GitHub language count][shield-lang-count]
![GitHub Actions Workflow Status][shield-gh-action-status]
![GitHub License][shield-license]

<h2>📋 Table of Contents</h2>

<!-- TOC -->
* [🧰 Terraform Provider: Helpers Functions](#-terraform-provider-helpers-functions)
  * [Available Functions](#available-functions)
  * [Example Usage](#example-usage)
<!-- TOC -->

# 🧰 Terraform Provider: Helpers Functions

This is a Terraform Provider only for Helper functions. The main idea is to extend the built-in functionalities of
Terraform with some functions that are not available by default and can be useful in some scenarios.

For a detailed examples and documentation check inside the [docs](./docs/index.md) folder or directly in the
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

<!-- MARKDOWN LINKS & IMAGES -->
[shield-golang]: <https://img.shields.io/badge/-Golang-black?style=for-the-badge&logoColor=white&logo=go&color=00ADD8>
[shield-terraform]: <https://img.shields.io/badge/-Terraform-black?style=for-the-badge&logoColor=white&logo=terraform&color=844FBA>
[shield-lang-count]: <https://img.shields.io/github/languages/count/inventium-tech/terraform-provider-helpers>
[shield-gh-action-status]: <https://img.shields.io/github/actions/workflow/status/inventium-tech/terraform-provider-helpers/go.yml?branch=main&logo=githubactions&logoColor=white&logoSize=5>
[shield-license]: <https://img.shields.io/github/license/inventium-tech/terraform-provider-helpers>

[badge-gh-action-build]: <https://github.com/inventium-tech/terraform-provider-helpers/actions/workflows/build.yml/badge.svg>
[badge-gh-action-megalinter]: <https://github.com/inventium-tech/terraform-provider-helpers/actions/workflows/mega-linter.yml/badge.svg>
[badge-gh-action-codeql]: <https://github.com/inventium-tech/terraform-provider-helpers/actions/workflows/codeql.yml/badge.svg>

[link-gh-action-build]: <https://github.com/inventium-tech/terraform-provider-helpers/actions/workflows/build.yml>
[link-gh-action-megalinter]: <https://github.com/inventium-tech/terraform-provider-helpers/actions/workflows/mega-linter.yml>
[link-gh-action-codeql]: <https://github.com/inventium-tech/terraform-provider-helpers/actions/workflows/codeql.yml>
