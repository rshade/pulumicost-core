# Research: CLI Filter Flag Support

**Feature**: CLI Filter Flag Support
**Branch**: `023-add-cli-filter-flag`
**Date**: 2025-12-23

## Decision: Add `--filter` Flag to `cost actual` Command

I will add the `--filter` flag to the `pulumicost cost actual` command in `internal/cli/cost_actual.go` to match the behavior of `pulumicost cost projected`.

### Rationale

- **Consistency**: The `projected` command already supports `--filter`. Users expect consistent flags across cost commands.
- **Test Compliance**: Existing integration tests (`TestActualCost_FilterByTag`) explicitly use this flag and are currently failing.
- **Code Reuse**: The `internal/engine` package already exports `FilterResources`, which takes a list of resources and a filter string. We can reuse this logic directly.
- **Separation of Concerns**: Currently, `actual` supports some filtering via `--group-by tag:key=value`. While functional, this conflates grouping and filtering. Adding an explicit `--filter` flag allows filtering by type, provider, OR tags, independent of how the output is grouped.

### Implementation Details

1.  **Modify `costActualParams` struct**: Add `filter []string` (slice to support multiple flags).
2.  **Update `NewCostActualCmd`**: Register the flag as a string array:
    ```go
    cmd.Flags().StringArrayVar(&params.filter, "filter", []string{}, "Resource filter expressions (e.g., 'type=aws:ec2/instance', 'tag:env=prod')")
    ```
3.  **Update `executeCostActual`**:
    -   Iteratively apply filters (AND logic).
    -   Call `engine.FilterResources` for each filter expression provided:
        ```go
        for _, f := range params.filter {
            if f != "" {
                resources = engine.FilterResources(resources, f)
                log.Debug().Ctx(ctx).Str("filter", f).Int("filtered_count", len(resources)).
                    Msg("applied resource filter")
            }
        }
        ```
    -   Pass the filtered resources to `engine.GetActualCostWithOptionsAndErrors`.

### Alternatives Considered

-   **Modify `ActualCostRequest` to handle raw filter strings**: We could pass the filter string down to the engine's `GetActualCost...` method. However, `engine.FilterResources` works on the `[]schema.Resource` slice *before* any cost calculation. Applying the filter early in the CLI handler (controller layer) is cleaner and matches the `projected` command pattern.
-   **Stick with `--group-by` for filtering**: rejected because it doesn't support filtering by type/provider without grouping by them, and it fails the "Test Compliance" requirement.

## Open Questions Resolved

-   **Where is the missing flag?**: `internal/cli/cost_actual.go`.
-   **How to implement?**: Mirror `internal/cli/cost_projected.go` implementation.
-   **Conflict with GroupBy?**: No conflict. Filtering happens first (reducing the dataset), then grouping happens on the result. `TestActualCost_FilterByTagAndType` confirms this interaction is desired.
