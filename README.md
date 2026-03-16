<p align="center">
  <img src="./assets/provider_logo.svg" width="200" alt="logo"/>
</p>

# Terraform Provider: Helpers

Utility Terraform provider that adds reusable helper functions to extend Terraform language capabilities.

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/inventium-tech/terraform-provider-helpers?style=for-the-badge&logo=go)
![Terraform version](https://img.shields.io/badge/terraform-v1.11-black?style=for-the-badge&logo=terraform&color=844FBA)
![Terraform Provider Downloads](https://img.shields.io/terraform/provider/dw/5756?style=for-the-badge&logo=terraform&color=844FBA)
![GitHub License](https://img.shields.io/github/license/inventium-tech/terraform-provider-helpers?style=for-the-badge)

[![GitHub Pre-Release](https://img.shields.io/github/v/release/inventium-tech/terraform-provider-helpers?include_prereleases&sort=semver&display_name=release&style=for-the-badge&logo=semanticrelease&label=pre-release&color=orange)](https://github.com/inventium-tech/terraform-provider-helpers/releases)
[![GitHub Release](https://img.shields.io/github/v/release/inventium-tech/terraform-provider-helpers?sort=semver&display_name=release&style=for-the-badge&logo=semanticrelease&color=green)](https://github.com/inventium-tech/terraform-provider-helpers/releases/latest)
[![Codecov](https://img.shields.io/codecov/c/github/inventium-tech/terraform-provider-helpers?style=for-the-badge&logo=codecov&label=CodeCov%20Coverage)](https://codecov.io/gh/inventium-tech/terraform-provider-helpers)
[![Sonar](https://img.shields.io/sonar/coverage/inventium-tech_terraform-provider-helpers?server=https%3A%2F%2Fsonarcloud.io&style=for-the-badge&logo=sonarqubecloud&label=SonarCloud%20Coverage)](https://sonarcloud.io/project/overview?id=inventium-tech_terraform-provider-helpers)
![Sonar Quality Gate](https://img.shields.io/sonar/quality_gate/inventium-tech_terraform-provider-helpers?server=https%3A%2F%2Fsonarcloud.io&style=for-the-badge&logo=sonarqubecloud)

## Quick Start

Terraform 1.8+ is required to use provider functions.

```terraform
terraform {
  required_providers {
    helpers = {
      source = "registry.terraform.io/inventium-tech/helpers"
    }
  }
}

locals {
  target_object = {
    key1 = "value1"
  }
}

output "updated_value" {
  value = provider::helpers::object_set_value(local.target_object, "key1", "new_value", "write_all")
}
```

Then run:

```sh
terraform init
terraform apply
```

Expected result: output `updated_value = {"key1" = "new_value"}`.

## What It Does

- Adds helper functions that complement built-in Terraform functions.
- Supports object, collection, and OS-environment function use cases.
- Focuses on function-only extensions (no resources or data sources).

## Available Functions

- Collection: [collection_filter](./docs/functions/collection_filter.md)
- Object: [object_set_value](./docs/functions/object_set_value.md), [object_filter_keys](./docs/functions/object_filter_keys.md), [object_contains_keys](./docs/functions/object_contains_keys.md)
- OS: [os_get_env](./docs/functions/os_get_env.md), [os_check_env](./docs/functions/os_check_env.md)

## Documentation

| Document | Purpose |
|----------|---------|
| [docs/index.md](./docs/index.md) | End-user provider and function reference |
| [CONTRIBUTING.md](./CONTRIBUTING.md) | Change process for contributors |
| [ARCHITECTURE.md](./ARCHITECTURE.md) | System structure and design decisions |
| [AGENTS.md](./AGENTS.md) | AI/automation execution rules |
| [LICENSE](./LICENSE) | License terms |

Supported platform note: CI validates on Linux with Go and Terraform 1.11.x/1.12.x.
