# Data Model: CLI Filter Flag Support

**Feature**: CLI Filter Flag Support
**Branch**: `023-add-cli-filter-flag`

## Overview

This feature does not introduce new persistent data models or API entities. It operates on the existing `schema.Resource` and `engine.CostResult` structures.

## Entities

### Filter Expression (Input)

Passed as a string argument to the CLI.

-   **Format**: `key=value`
-   **Supported Keys**: `type`, `provider`, `tag:<key>`
-   **Example**: `type=aws:ec2/instance`, `tag:Environment=prod`
-   **Validation**: Handled by `internal/engine/filter.go` (existing logic).

## API Contracts

N/A - CLI Feature.
