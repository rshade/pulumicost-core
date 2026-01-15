# Feature Specification: Error Aggregation in Proto Adapter

**Feature Branch**: `003-proto-error-aggregation`  
**Created**: November 23, 2025  
**Status**: Draft

## Clarifications

### Session 2025-11-24

- Q: What logging approach should be used for structured error logging? → A: Use structured logging with zerolog (adds dependency)
- Q: What is acceptable PR size given zerolog dependency addition? → A: Allow ~150-200 lines (zerolog + full implementation)
- Q: How should engine surface errors to CLI layer? → A: Change engine to return `(*CostResultWithErrors, error)` (no backward compatibility concerns)
- Q: Should GetActualCost engine method also return CostResultWithErrors? → A: Yes, standardize both methods
- Q: How should CLI display errors to users? → A: Both inline in table (Notes column with ERROR prefix) AND summary after table

**Input**: User description: "title: [HIGH] Add error aggregation in proto adapter state: OPEN author: rshade labels: comments: 0 assignees: projects: milestone: 2025-Q1 - Core v0.1.0 MVP number: 100 -- # [HIGH] Add Error Aggregation in Proto Adapter ## Priority: HIGH ## Category: Error Handling ## Milestone: 2025-Q1 - Core v0.1.0 MVP ## Estimated Effort: 2-3 hours ## PR Size: Small (~100 lines) --- ## Problem Statement `GetProjectedCost` and `GetActualCost` in proto adapter continue silently on errors. Users don't know which resources failed, making debugging nearly impossible. ## Current Problematic Code **File**: `internal/proto/adapter.go:96-142` ```go // Lines 124-126 - Silent failure for _, resource := range resources { result, err := client.GetProjectedCost(ctx, req) if err != nil { // ERROR: Just continues, no tracking continue } results = append(results, convertResult(result)) } // Similar issue in GetActualCost (lines 160-164) for _, resource := range resources { result, err := client.GetActualCost(ctx, req) if err != nil { continue // ERROR: Silent failure } results = append(results, convertResult(result)) } ``` **Impact**: - Users don't know which resources failed - No visibility into plugin failures - Difficult to debug cost calculation issues - Silent data loss --- ## Implementation Guide ### Step 1: Create Error Aggregation Structure **Add to** `internal/proto/adapter.go`: ```go // ErrorDetail captures information about a failed resource cost calculation type ErrorDetail struct { ResourceType string ResourceID string PluginName string Error error Timestamp time.Time } // CostResultWithErrors wraps results and any errors encountered type CostResultWithErrors struct { Results []engine.CostResult Errors []ErrorDetail } // HasErrors returns true if any errors were encountered func (c *CostResultWithErrors) HasErrors() bool { return len(c.Errors) > 0 } // ErrorSummary returns a human-readable summary of errors func (c *CostResultWithErrors) ErrorSummary() string { if !c.HasErrors() { return "" } var summary strings.Builder summary.WriteString(fmt.Sprintf("%d resource(s) failed:\n", len(c.Errors))) for i, err := range c.Errors { if i >= 5 { summary.WriteString(fmt.Sprintf("... and %d more errors\n", len(c.Errors)-5)) break } summary.WriteString(fmt.Sprintf(" - %s (%s): %v\n", err.ResourceType, err.ResourceID, err.Error)) } return summary.String() } ``` ### Step 2: Update GetProjectedCost with Error Tracking **Replace** `GetProjectedCost` function: ```go func (a *Adapter) GetProjectedCost( ctx context.Context, client *pluginhost.Client, resources []engine.ResourceDescriptor, ) *CostResultWithErrors { result := &CostResultWithErrors{ Results: []engine.CostResult{}, Errors: []ErrorDetail{}, } pluginName := "unknown" if client != nil { pluginName = client.Name() } for _, resource := range resources { // Build request req := &pb.GetProjectedCostRequest{ ResourceType: resource.Type, ResourceId: resource.ID, Properties: convertProperties(resource.Properties), } // Call plugin resp, err := client.GetProjectedCost(ctx, req) if err != nil { // NEW: Track error instead of silent failure result.Errors = append(result.Errors, ErrorDetail{ ResourceType: resource.Type, ResourceID: resource.ID, PluginName: pluginName, Error: fmt.Errorf("plugin call failed: %w", err), Timestamp: time.Now(), }) // Add placeholder result with error note result.Results = append(result.Results, engine.CostResult{ ResourceType: resource.Type, ResourceID: resource.ID, Adapter: pluginName, Currency: "USD", Monthly: 0, Hourly: 0, Notes: fmt.Sprintf("ERROR: %v", err), }) continue } // Convert successful result costResult := engine.CostResult{ ResourceType: resource.Type, ResourceID: resource.ID, Adapter: pluginName, Currency: resp.Currency, Monthly: resp.MonthlyEstimate, Hourly: resp.HourlyRate, Notes: resp.Notes, } result.Results = append(result.Results, costResult) } return result } ``` ### Step 3: Update GetActualCost Similarly **Replace** `GetActualCost` function: ```go func (a *Adapter) GetActualCost( ctx context.Context, client *pluginhost.Client, resources []engine.ResourceDescriptor, from, to time.Time, ) *CostResultWithErrors { result := &CostResultWithErrors{ Results: []engine.CostResult{}, Errors: []ErrorDetail{}, } pluginName := "unknown" if client != nil { pluginName = client.Name() } for _, resource := range resources { req := &pb.GetActualCostRequest{ ResourceType: resource.Type, ResourceId: resource.ID, StartDate: timestamppb.New(from), EndDate: timestamppb.New(to), } resp, err := client.GetActualCost(ctx, req) if err != nil { // NEW: Track error result.Errors = append(result.Errors, ErrorDetail{ ResourceType: resource.Type, ResourceID: resource.ID, PluginName: pluginName, Error: fmt.Errorf("plugin call failed: %w", err), Timestamp: time.Now(), }) result.Results = append(result.Results, engine.CostResult{ ResourceType: resource.Type, ResourceID: resource.ID, Adapter: pluginName, Currency: "USD", TotalCost: 0, Notes: fmt.Sprintf("ERROR: %v", err), StartDate: from, EndDate: to, }) continue } costResult := engine.CostResult{ ResourceType: resource.Type, ResourceID: resource.ID, Adapter: pluginName, Currency: resp.Currency, TotalCost: resp.TotalCost, StartDate: from, EndDate: to, Notes: resp.Notes, } result.Results = append(result.Results, costResult) } return result } ``` ### Step 4: Update Engine to Handle Errors **Update** `internal/engine/engine.go` to use new return type: ```go func (e *Engine) GetProjectedCost( ctx context.Context, resources []ResourceDescriptor, ) ([]CostResult, error) { var allResults []CostResult var allErrors []proto.ErrorDetail // Try each plugin for _, client := range e.pluginClients { resultWithErrs := e.adapter.GetProjectedCost(ctx, client, resources) allResults = append(allResults, resultWithErrs.Results...) allErrors = append(allErrors, resultWithErrs.Errors...) } // Log errors if any if len(allErrors) > 0 { log.Printf("Encountered %d errors during cost calculation:\n", len(allErrors)) for _, err := range allErrors { log.Printf(" - %s/%s: %v\n", err.ResourceType, err.ResourceID, err.Error) } } // Fallback to specs for failed resources // ... existing spec fallback logic ... return allResults, nil } ``` ### Step 5: Add Structured Logging **Add** logging package integration: ```go import ( "github.com/rshade/finfocus/internal/logging" ) // In error handling: logger := logging.GetLogger() logger.WithFields(logging.Fields{ "resource_type": err.ResourceType, "resource_id": err.ResourceID, "plugin": err.PluginName, "timestamp": err.Timestamp, }).Error("Plugin cost calculation failed", err.Error) ``` --- ## Testing **Create** `internal/proto/adapter_test.go`: ```go package proto import ( "context" "errors" "testing" "github.com/rshade/finfocus/internal/engine" ) func TestErrorAggregation_GetProjectedCost(t *testing.T) { adapter := NewAdapter() // Mock client that returns errors mockClient := &mockPluginClient{ projectededCostFunc: func(ctx context.Context, req *pb.GetProjectedCostRequest) (*pb.GetProjectedCostResponse, error) { // Fail for specific resource if req.ResourceId == "fail-resource" { return nil, errors.New("simulated plugin failure") } return &pb.GetProjectedCostResponse{ MonthlyEstimate: 100.0, HourlyRate: 0.137, Currency: "USD", }, nil }, } resources := []engine.ResourceDescriptor{ {Type: "aws:ec2:Instance", ID: "success-1"}, {Type: "aws:ec2:Instance", ID: "fail-resource"}, {Type: "aws:ec2:Instance", ID: "success-2"}, } result := adapter.GetProjectedCost(context.Background(), mockClient, resources) // Should have 3 results (2 success, 1 error placeholder) if len(result.Results) != 3 { t.Errorf("Expected 3 results, got %d", len(result.Results)) } // Should have 1 error if len(result.Errors) != 1 { t.Errorf("Expected 1 error, got %d", len(result.Errors)) } // Error should be tracked if !result.HasErrors() { t.Error("HasErrors() should return true") } // Error details should be correct if result.Errors[0].ResourceID != "fail-resource" { t.Errorf("Error ResourceID = %s, want fail-resource", result.Errors[0].ResourceID) } // Error summary should be informative summary := result.ErrorSummary() if summary == "" { t.Error("ErrorSummary() should return non-empty string") } if !strings.Contains(summary, "fail-resource") { t.Error("ErrorSummary() should mention failed resource") } } func TestErrorSummary_MultipleErrors(t *testing.T) { result := &CostResultWithErrors{ Errors: []ErrorDetail{}, } // Add 10 errors for i := 0; i < 10; i++ { result.Errors = append(result.Errors, ErrorDetail{ ResourceType: "aws:ec2:Instance", ResourceID: fmt.Sprintf("r%d", i), Error: errors.New("error 1"), }) } summary := result.ErrorSummary() // Should mention count if !strings.Contains(summary, "3 resource(s)") { t.Error("Summary should mention total error count") } // Should list resources if !strings.Contains(summary, "r1") || !strings.Contains(summary, "r2") { t.Error("Summary should list failed resources") } } func TestErrorSummary_ManyErrors(t *testing.T) { result := &CostResultWithErrors{Errors: []ErrorDetail{}} // Add 10 errors for i := 0; i < 10; i++ { result.Errors = append(result.Errors, ErrorDetail{ ResourceType: "aws:ec2:Instance", ResourceID: fmt.Sprintf("r%d", i), Error: errors.New("error"), }) } summary := result.ErrorSummary() // Should truncate after 5 if !strings.Contains(summary, "and 5 more") { t.Error("Summary should indicate truncation") } } ``` --- ## Acceptance Criteria - [ ] Create ErrorDetail and CostResultWithErrors types - [ ] Update GetProjectedCost to track errors - [ ] Update GetActualCost to track errors - [ ] Add placeholder results for failed resources (with error in Notes) - [ ] Implement ErrorSummary() for user-friendly error reporting - [ ] Update engine to log aggregated errors - [ ] Add structured logging for error details - [ ] Create comprehensive tests for error aggregation - [ ] Test with 0 errors, 1 error, many errors - [ ] Verify error truncation works (>5 errors) - [ ] Maintain backward compatibility with engine - [ ] All tests pass: `go test ./internal/proto/...` --- ## Manual Testing ```bash # Build with changes make build # Test with failing plugin (manually stop a plugin) ./bin/finfocus cost projected --pulumi-json examples/plans/aws-simple-plan.json # Should see error summary in output # Example output: # # COST SUMMARY # ============ # Total Monthly Cost: 50.00 USD # Total Resources: 3 # # ⚠️ 2 resource(s) failed: # - aws:ec2:Instance (i-123): plugin call failed: connection refused # - aws:rds:Instance (db-456): plugin call failed: timeout ``` --- ## PR Strategy **Small PR focused on error handling**: - ~100 lines of new code - New types for error tracking - Updated adapter functions - Comprehensive tests - Maintains backward compatibility ## Reference See issues.md for full details" 

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Error Reporting for Projected Costs (Priority: P1)

When a plugin fails to calculate the projected cost for one or more resources, the system should still return results for successfully processed resources and clearly report the details of the failed resources to the user.

**Why this priority**: Directly addresses the core problem statement of silent failures and provides immediate value by making errors visible.

**Independent Test**: A mock plugin can be configured to fail for specific resources. The output of `GetProjectedCost` can then be checked to ensure successful resources are present, errors are aggregated, and the error summary is informative.

**Acceptance Scenarios**:

1.  **Given** a request for projected costs with multiple resources, **When** one plugin fails for a subset of resources, **Then** `CostResultWithErrors` contains results for successful resources and `ErrorDetail` entries for failed resources.
2.  **Given** a `CostResultWithErrors` object with at least one error, **When** `HasErrors()` is called, **Then** it returns `true`.
3.  **Given** a `CostResultWithErrors` object with errors, **When** `ErrorSummary()` is called, **Then** it returns a human-readable string listing the failed resources.

---

### User Story 2 - Error Reporting for Actual Costs (Priority: P1)

Similar to projected costs, when a plugin fails to calculate the actual cost for resources, the system should provide results for successful ones and detailed error information for failures.

**Why this priority**: Addresses the second main use case (`GetActualCost`) with the same critical need for error visibility.

**Independent Test**: A mock plugin can be configured to fail for specific resources during actual cost calculation. The output of `GetActualCost` can then be checked for successful results, aggregated errors, and an informative summary.

**Acceptance Scenarios**:

1.  **Given** a request for actual costs with multiple resources, **When** one plugin fails for a subset of resources, **Then** `CostResultWithErrors` contains results for successful resources and `ErrorDetail` entries for failed resources.

---

### User Story 3 - Comprehensive Error Summary and Logging (Priority: P2)

The system should provide an aggregated error summary that truncates long lists of errors for readability and logs detailed error information for debugging purposes.

**Why this priority**: Enhances usability and debuggability by providing both concise user feedback and detailed internal logging.

**Independent Test**: Generate a scenario with more than 5 errors and verify that `ErrorSummary()` truncates the output while accurately indicating the total error count. Verify that structured logs contain the expected error details.

**Acceptance Scenarios**:

1.  **Given** a `CostResultWithErrors` object with more than 5 errors, **When** `ErrorSummary()` is called, **Then** it truncates the output after 5 errors and indicates the remaining count.
2.  **Given** a plugin failure, **When** the error is aggregated, **Then** the system logs the error with structured fields including `resource_type`, `resource_id`, `plugin`, `error`, and `timestamp`.

### Edge Cases

-   What happens when all plugins fail for all resources? The system should return an empty `Results` slice and a full `Errors` slice, with an informative summary.
-   How does the system handle an empty list of resources? It should return empty `Results` and `Errors`.
-   What happens if a plugin returns a valid response but with an empty cost? The system should record 0 cost and no error.

## Requirements *(mandatory)*

### Functional Requirements

-   **FR-001**: The system MUST define an `ErrorDetail` struct to capture specific information about failed resource cost calculations (ResourceType, ResourceID, PluginName, Error, Timestamp).
-   **FR-002**: The system MUST define a `CostResultWithErrors` struct to wrap a slice of `engine.CostResult` and a slice of `ErrorDetail`.
-   **FR-003**: The `CostResultWithErrors` struct MUST include a `HasErrors()` method that returns `true` if any errors were encountered.
-   **FR-004**: The `CostResultWithErrors` struct MUST include an `ErrorSummary()` method that returns a human-readable summary of errors, truncating the output after 5 errors.
-   **FR-005**: The `Adapter.GetProjectedCost` function MUST be updated to return `*CostResultWithErrors`.
-   **FR-006**: When `Adapter.GetProjectedCost` encounters an error from a plugin, it MUST track the error in the `Errors` slice of `CostResultWithErrors` and add a placeholder `engine.CostResult` to the `Results` slice with the error noted in its `Notes` field.
-   **FR-007**: The `Adapter.GetActualCost` function MUST be updated to return `*CostResultWithErrors`.
-   **FR-008**: When `Adapter.GetActualCost` encounters an error from a plugin, it MUST track the error in the `Errors` slice of `CostResultWithErrors` and add a placeholder `engine.CostResult` to the `Results` slice with the error noted in its `Notes` field.
-   **FR-009**: The `Engine.GetProjectedCost` and `Engine.GetActualCost` functions MUST be updated to return `*CostResultWithErrors` (breaking change acceptable).
-   **FR-010**: Both engine functions MUST aggregate all `ErrorDetail` objects from all plugin calls.
-   **FR-011**: Both engine functions MUST log aggregated errors using structured logging with `zerolog`.
-   **FR-012**: The CLI MUST display errors both inline in the table (Notes column with "ERROR:" prefix) AND as a summary after the table.

### Key Entities *(include if feature involves data)*

-   **ErrorDetail**: Stores details of a single failed resource cost calculation, including `ResourceType`, `ResourceID`, `PluginName`, `Error` (Go `error` type), and `Timestamp`.
-   **CostResultWithErrors**: A container type that holds both successfully calculated `engine.CostResult` objects and `ErrorDetail` objects for failures.

## Success Criteria *(mandatory)*

### Measurable Outcomes

-   **SC-001**: 100% of errors occurring during `GetProjectedCost` and `GetActualCost` plugin calls are captured and reported via `CostResultWithErrors`.
-   **SC-002**: The `ErrorSummary()` function accurately summarizes error details for single and multiple failures, truncating long lists effectively (verified for >5 errors).
-   **SC-003**: Detailed error information (resource type, ID, plugin name, error message, timestamp) is consistently available in aggregated logs for every plugin failure.
-   **SC-004**: The system's cost calculation functionality remains operational, returning partial results alongside error reports, even when some plugin calls fail.