# GetRecommendations Feature Research & Design

> This document captures research, design decisions, and implementation
> details for the recommendations feature across pulumicost-spec and
> pulumicost-core.

## Related Issues

- **pulumicost-spec#122**: GetRecommendations RPC, proto definitions, pluginsdk interface
- **pulumicost-core#216**: CLI command, output rendering, filtering, aggregation

## Cross-Provider Research Summary

### AWS Cost Explorer Rightsizing

**API**: `GetRightsizingRecommendation`

**Documentation**: <https://docs.aws.amazon.com/aws-cost-management/latest/APIReference/API_GetRightsizingRecommendation.html>

**Key Fields**:

- `RightsizingType`: TERMINATE or MODIFY
- `CurrentInstance`: AccountId, ResourceId, InstanceType, ResourceUtilization
- `ModifyRecommendationDetail`: TargetInstances with projected utilization
- `TerminateRecommendationDetail`: EstimatedMonthlySavings
- `FindingReasonCodes`: CPU_OVER_PROVISIONED, MEMORY_OVER_PROVISIONED, etc.

**ResourceUtilization Metrics**:

- CPU (max percentage)
- Memory (if CloudWatch agent enabled)
- Network In/Out
- Disk I/O Read/Write
- EBS Volume performance

**Analysis Window**: Last 14 days of usage data

### Kubecost Request Right-Sizing

**API**: `/model/savings/requestSizingV2`

**Documentation**: <https://docs.kubecost.com/apis/savings-apis/api-request-right-sizing-v2>

**Key Fields**:

- `clusterID`, `namespace`, `controllerKind`, `controllerName`, `containerName`
- `currentRequest`: { cpu, memory }
- `recommendedRequest`: { cpu, memory }
- `currentEfficiency`: percentage
- `monthlySavings`: dollar amount

**Algorithm Options**:

- `max`: Use maximum observed usage
- `quantile`: Use percentile-based sizing (e.g., p95)

**Parameters**:

- `window`: Time range for analysis
- `targetUtilization`: Desired utilization percentage
- `minRecommendationCPU`: Minimum CPU recommendation

### Azure Advisor

**API**: `GET /subscriptions/{subscriptionId}/providers/Microsoft.Advisor/recommendations`

**Documentation**: <https://learn.microsoft.com/en-us/rest/api/advisor/recommendations/list>

**Categories**:

- Cost
- HighAvailability
- Security
- Performance
- OperationalExcellence

**Key Fields**:

- `properties.category`: Recommendation category
- `properties.impact`: High, Medium, Low
- `properties.impactedField`: Resource type affected
- `properties.impactedValue`: Resource name
- `properties.shortDescription.problem`: Issue description
- `properties.shortDescription.solution`: Recommended action
- `properties.remediation`: Detailed remediation steps

### GCP Recommender

**API**: Cloud Recommender API v1

**Documentation**: <https://cloud.google.com/recommender/docs/reference/rpc/google.cloud.recommender.v1>

**Categories**:

- Cost
- Security
- Performance
- Reliability
- Manageability

**Priority Levels**: P1 (highest) through P4 (lowest)

**States**: ACTIVE, CLAIMED, SUCCEEDED, FAILED, DISMISSED

**Key Message Structure**:

```protobuf
message Recommendation {
  string name = 1;
  string description = 2;
  string recommender_subtype = 3;
  google.protobuf.Timestamp last_refresh_time = 4;
  Impact primary_impact = 5;
  repeated Impact additional_impact = 6;
  Priority priority = 7;
  RecommendationContent content = 8;
  RecommendationStateInfo state_info = 9;
  string etag = 10;
  repeated InsightReference associated_insights = 11;
  string xor_group_id = 12;
}
```

**Impact Types**:

- CostProjection: Estimated savings
- SecurityProjection: Security improvements
- SustainabilityProjection: Carbon footprint reduction
- ReliabilityProjection: Availability improvements

### FOCUS Specification

**Finding**: FOCUS does not define recommendation formats. It focuses solely
on billing data standardization. PulumiCost needs its own recommendation
schema.

## Proto Schema Design

### Enums

```protobuf
enum RecommendationCategory {
  RECOMMENDATION_CATEGORY_UNSPECIFIED = 0;
  RECOMMENDATION_CATEGORY_COST = 1;
  RECOMMENDATION_CATEGORY_PERFORMANCE = 2;
  RECOMMENDATION_CATEGORY_SECURITY = 3;
  RECOMMENDATION_CATEGORY_RELIABILITY = 4;
}

enum RecommendationActionType {
  RECOMMENDATION_ACTION_TYPE_UNSPECIFIED = 0;
  RECOMMENDATION_ACTION_TYPE_RIGHTSIZE = 1;
  RECOMMENDATION_ACTION_TYPE_TERMINATE = 2;
  RECOMMENDATION_ACTION_TYPE_PURCHASE_COMMITMENT = 3;
  RECOMMENDATION_ACTION_TYPE_ADJUST_REQUESTS = 4;
  RECOMMENDATION_ACTION_TYPE_MODIFY = 5;
  RECOMMENDATION_ACTION_TYPE_DELETE_UNUSED = 6;
}

enum RecommendationPriority {
  RECOMMENDATION_PRIORITY_UNSPECIFIED = 0;
  RECOMMENDATION_PRIORITY_LOW = 1;
  RECOMMENDATION_PRIORITY_MEDIUM = 2;
  RECOMMENDATION_PRIORITY_HIGH = 3;
  RECOMMENDATION_PRIORITY_CRITICAL = 4;
}
```

### Request/Response

```protobuf
message GetRecommendationsRequest {
  RecommendationFilter filter = 1;
  string projection_period = 2;  // "monthly" (default), "annual", "daily"
  int32 page_size = 3;
  string page_token = 4;
}

message RecommendationFilter {
  string provider = 1;
  string region = 2;
  string resource_type = 3;
  RecommendationCategory category = 4;
  RecommendationActionType action_type = 5;
}

message GetRecommendationsResponse {
  repeated Recommendation recommendations = 1;
  RecommendationSummary summary = 2;
  string next_page_token = 3;
}
```

### Core Recommendation Message

```protobuf
message Recommendation {
  string id = 1;
  RecommendationCategory category = 2;
  RecommendationActionType action_type = 3;
  ResourceInfo resource = 4;

  oneof action_detail {
    RightsizeAction rightsize = 5;
    TerminateAction terminate = 6;
    CommitmentAction commitment = 7;
    KubernetesAction kubernetes = 8;
    ModifyAction modify = 9;
  }

  RecommendationImpact impact = 10;
  RecommendationPriority priority = 11;
  optional double confidence_score = 12;  // 0.0-1.0, nil if unavailable
  string description = 13;
  repeated string reasoning = 14;
  string source = 15;
  google.protobuf.Timestamp created_at = 16;
  map<string, string> metadata = 17;
}
```

### Action-Specific Messages

```protobuf
message RightsizeAction {
  string current_sku = 1;
  string recommended_sku = 2;
  string current_instance_type = 3;
  string recommended_instance_type = 4;
  ResourceUtilization projected_utilization = 5;
}

message TerminateAction {
  string termination_reason = 1;
  int32 idle_days = 2;
}

message CommitmentAction {
  string commitment_type = 1;  // "reserved_instance", "savings_plan", "cud"
  string term = 2;             // "1_year", "3_year"
  string payment_option = 3;
  double recommended_quantity = 4;
  string scope = 5;
}

message KubernetesAction {
  string cluster_id = 1;
  string namespace = 2;
  string controller_kind = 3;
  string controller_name = 4;
  string container_name = 5;
  KubernetesResources current_requests = 6;
  KubernetesResources recommended_requests = 7;
  KubernetesResources current_limits = 8;
  KubernetesResources recommended_limits = 9;
  string algorithm = 10;
}

message KubernetesResources {
  string cpu = 1;
  string memory = 2;
}

message ModifyAction {
  string modification_type = 1;
  map<string, string> current_config = 2;
  map<string, string> recommended_config = 3;
}
```

### Supporting Messages

```protobuf
message ResourceInfo {
  string id = 1;
  string name = 2;
  string provider = 3;
  string resource_type = 4;
  string region = 5;
  string sku = 6;
  map<string, string> tags = 7;
  ResourceUtilization utilization = 8;
}

message ResourceUtilization {
  double cpu_percent = 1;
  double memory_percent = 2;
  double storage_percent = 3;
  double network_in_mbps = 4;
  double network_out_mbps = 5;
  map<string, double> custom_metrics = 6;
}

message RecommendationImpact {
  double estimated_savings = 1;
  string currency = 2;
  string projection_period = 3;
  double current_cost = 4;
  double projected_cost = 5;
  double savings_percentage = 6;
  optional double implementation_cost = 7;
  optional double migration_effort_hours = 8;
}

message RecommendationSummary {
  int32 total_recommendations = 1;
  double total_estimated_savings = 2;
  string currency = 3;
  string projection_period = 4;
  map<string, int32> count_by_category = 5;
  map<string, double> savings_by_category = 6;
  map<string, int32> count_by_action_type = 7;
  map<string, double> savings_by_action_type = 8;
}
```

## CLI UX Design

### Command Structure

```bash
# Basic usage (--pulumi-json required)
pulumicost cost recommendations --pulumi-json plan.json

# With filters
pulumicost cost recommendations --pulumi-json plan.json --filter "category=cost"
pulumicost cost recommendations --pulumi-json plan.json --filter "action=rightsize"
pulumicost cost recommendations --pulumi-json plan.json --filter "priority=high"
pulumicost cost recommendations --pulumi-json plan.json --filter "savings>100"

# Verbose mode
pulumicost cost recommendations --pulumi-json plan.json --verbose

# Output formats
pulumicost cost recommendations --pulumi-json plan.json --output json
pulumicost cost recommendations --pulumi-json plan.json --output ndjson
```

### Default Table Output (Summary Mode)

Shows summary + top 5 recommendations:

```text
RECOMMENDATIONS SUMMARY
=======================
Total Recommendations:    12
Total Potential Savings:  $1,234.56/month

BY CATEGORY                              BY ACTION TYPE
-----------                              --------------
Cost:         8  ($987.00/month)         Rightsize:            6  ($654.00/month)
Performance:  3  ($200.00/month)         Terminate:            3  ($333.00/month)
Reliability:  1  ($47.56/month)          Purchase Commitment:  2  ($200.00/month)
                                         Adjust Requests:      1  ($47.56/month)

TOP RECOMMENDATIONS (use --verbose for all)
===========================================
Priority  Resource                    Action     Savings     Source
--------  --------                    ------     -------     ------
HIGH      aws:ec2/i-abc123            Rightsize  $120.00/mo  aws-plugin
HIGH      aws:ec2/i-def456            Terminate  $89.00/mo   aws-plugin
MEDIUM    gcp:compute/vm-123          Rightsize  $150.00/mo  gcp-plugin
MEDIUM    k8s:ns/prod/deploy/api      Adjust     $47.56/mo   kubecost
LOW       aws:rds/mydb                Rightsize  $45.00/mo   aws-plugin

(7 more recommendations, run with --verbose to see all)
```

### Verbose Table Output

```text
ALL RECOMMENDATIONS
===================
[1] HIGH - Rightsize aws:ec2/i-abc123
    Current: t3.xlarge → Recommended: t3.medium
    Savings: $120.00/month (40% reduction)
    Reason: CPU utilization avg 12%, Memory avg 8% over 14 days
    Source: aws-plugin

[2] HIGH - Terminate aws:ec2/i-def456
    Reason: Instance idle for 30+ days (0% CPU, no network traffic)
    Savings: $89.00/month (100% reduction)
    Source: aws-plugin

[3] MEDIUM - Adjust Requests k8s:ns/prod/deploy/api
    Current: 2000m CPU, 4Gi RAM → Recommended: 500m CPU, 1Gi RAM
    Savings: $47.56/month (75% reduction)
    Reason: Container efficiency at 15% CPU, 22% memory
    Source: kubecost
```

### Filter Expression Syntax

```bash
# Single filter
--filter "category=cost"
--filter "priority=high"
--filter "savings>100"

# Multiple conditions (AND logic)
--filter "category=cost,priority=high"
--filter "provider=aws,action=rightsize"
--filter "savings>50,category=performance"
```

**Supported Filter Fields**:

| Field      | Values                                    | Example          |
| ---------- | ----------------------------------------- | ---------------- |
| `category` | cost, performance, security, reliability  | `category=cost`  |
| `action`   | rightsize, terminate, purchase_commitment | `action=right..` |
| `priority` | low, medium, high, critical               | `priority=high`  |
| `provider` | aws, gcp, azure, kubernetes               | `provider=aws`   |
| `savings`  | numeric with operators                    | `savings>100`    |

**Note**: `action` also supports: `adjust_requests`, `modify`, `delete_unused`

## Design Decisions

### User Requirements Captured

1. **Categories**: All FinOps categories (Cost, Performance, Security,
   Reliability)
2. **Actions**: All action types (rightsizing, terminate, commitment, K8s,
   modify, delete)
3. **State**: Stateless - just return current recommendations, no tracking
4. **CLI Output**: Both table (default) and JSON formats
5. **Filtering**: Both modes - no filter = all, with filter = filtered
6. **Savings**: Flexible projection period, default to monthly
7. **Priority/Confidence**: Both priority enum AND numeric confidence score (optional)
8. **Input**: Require `--pulumi-json` (consistent with other commands)
9. **Table Detail**: Summary only by default, `--verbose` for full details
10. **Truncation**: Top 5 recommendations in summary mode

### Technical Decisions

- **PluginSDK**: `RecommendationsProvider` as optional interface
- **Aggregation**: pulumicost-core aggregates from all plugins
- **Sorting**: By savings descending for display
- **Currency**: Use existing `sdk/go/currency` package for validation
- **Mixed currencies**: Warning in output, no aggregation of totals

## Implementation File Structure

### pulumicost-spec

```text
proto/pulumicost/v1/
├── costsource.proto          # Add GetRecommendations RPC
└── recommendations.proto     # New file for recommendation messages (optional)

sdk/go/
├── pluginsdk/
│   └── sdk.go               # Add RecommendationsProvider interface
└── testing/
    ├── harness.go           # Add ValidateRecommendationsResponse
    └── mock_plugin.go       # Add GetRecommendations mock implementation
```

### pulumicost-core

```text
internal/cli/
└── cost_recommendations.go   # New command implementation

internal/engine/
├── recommendations.go        # Types and aggregation logic
└── recommendations_render.go # Output rendering functions
```

## Testing Strategy

### Unit Tests

- Command flag parsing
- Filter expression parsing
- Output rendering (table, JSON, NDJSON)
- Summary computation
- Truncation logic

### Integration Tests

- Mock plugin responses
- Multi-plugin aggregation
- Filter application
- Error handling (plugin failures, mixed currencies)

### Edge Cases

- No recommendations available
- Single plugin with no support
- Mixed currencies across plugins
- Very large recommendation lists
- Invalid filter expressions

## Bubble Tea TUI Integration

### Dependencies

```go
import (
    "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/bubbles/table"
    "github.com/charmbracelet/bubbles/spinner"
    "github.com/charmbracelet/lipgloss"
    "golang.org/x/term"
)
```

### Output Mode Router

```go
func runRecommendations(cmd *cobra.Command, args []string) error {
    outputFormat, _ := cmd.Flags().GetString("output")

    switch outputFormat {
    case "json":
        // Direct JSON encoding - bypass TUI
        return json.NewEncoder(os.Stdout).Encode(results)
    case "ndjson":
        for _, r := range results {
            json.NewEncoder(os.Stdout).Encode(r)
        }
        return nil
    default:
        // Table output - use TUI if TTY, plain text otherwise
        return renderTableOutput(results)
    }
}

func renderTableOutput(results []Recommendation) error {
    isTTY := term.IsTerminal(int(os.Stdout.Fd()))

    if isTTY {
        // Full interactive TUI with Bubble Tea
        return runInteractiveTUI(results)
    }
    // Plain text for CI/CD (Lip Gloss styling only)
    return renderPlainTable(results)
}
```

### Interactive TUI Model

```go
type recommendationsModel struct {
    table       table.Model
    summary     RecommendationSummary
    selected    *Recommendation
    showDetail  bool
    filter      string
    width       int
    height      int
}

func (m recommendationsModel) Init() tea.Cmd {
    return nil
}

func (m recommendationsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "q", "ctrl+c":
            return m, tea.Quit
        case "enter":
            // Show detail view for selected recommendation
            m.showDetail = true
            m.selected = getSelectedRecommendation(m.table.SelectedRow())
        case "esc":
            m.showDetail = false
        case "/":
            // Enter filter mode
            // ... filter input handling
        }
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        m.table.SetWidth(msg.Width)
        m.table.SetHeight(msg.Height - 10) // Reserve space for summary
    }

    var cmd tea.Cmd
    m.table, cmd = m.table.Update(msg)
    return m, cmd
}

func (m recommendationsModel) View() string {
    if m.showDetail {
        return renderDetailView(m.selected)
    }
    return renderSummaryAndTable(m.summary, m.table)
}
```

### Styled Components with Lip Gloss

```go
var (
    // Priority styles
    criticalStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("196")).
        Bold(true)
    highStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("208"))
    mediumStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("226"))
    lowStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("82"))

    // Section styles
    headerStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("99")).
        BorderStyle(lipgloss.NormalBorder()).
        BorderBottom(true)

    savingsStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("82")).
        Bold(true)

    // Table styles
    selectedRowStyle = lipgloss.NewStyle().
        Background(lipgloss.Color("237")).
        Foreground(lipgloss.Color("229"))
)

func renderPriority(priority string) string {
    switch priority {
    case "CRITICAL":
        return criticalStyle.Render("🚨 CRITICAL")
    case "HIGH":
        return highStyle.Render("⚠ HIGH")
    case "MEDIUM":
        return mediumStyle.Render("◉ MEDIUM")
    default:
        return lowStyle.Render("○ LOW")
    }
}
```

### Interactive Table Configuration

```go
func createRecommendationsTable(recs []Recommendation) table.Model {
    columns := []table.Column{
        {Title: "Priority", Width: 12},
        {Title: "Resource", Width: 30},
        {Title: "Action", Width: 15},
        {Title: "Savings", Width: 12},
        {Title: "Source", Width: 15},
    }

    rows := make([]table.Row, len(recs))
    for i, r := range recs {
        rows[i] = table.Row{
            r.Priority.String(),
            truncate(r.Resource.Name, 28),
            r.ActionType.String(),
            formatMoney(r.Impact.EstimatedSavings),
            r.Source,
        }
    }

    t := table.New(
        table.WithColumns(columns),
        table.WithRows(rows),
        table.WithFocused(true),
        table.WithHeight(15),
    )

    // Apply custom styles
    s := table.DefaultStyles()
    s.Header = headerStyle
    s.Selected = selectedRowStyle
    t.SetStyles(s)

    return t
}
```

### Detail View (Enter on Row)

```text
┌─────────────────────────────────────────────────────────────────┐
│ ⚠ HIGH - Rightsize aws:ec2/i-abc123                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│ CURRENT STATE                    RECOMMENDED                    │
│ ─────────────                    ───────────                    │
│ Instance: t3.xlarge              Instance: t3.medium            │
│ vCPU: 4                          vCPU: 2                        │
│ Memory: 16 GiB                   Memory: 8 GiB                  │
│ Cost: $300.00/month              Cost: $180.00/month            │
│                                                                 │
│ UTILIZATION (14-day avg)                                        │
│ ─────────────────────────                                       │
│ CPU:    ████░░░░░░░░░░░░░░░░ 12%                                │
│ Memory: ██░░░░░░░░░░░░░░░░░░  8%                                │
│                                                                 │
│ SAVINGS                                                         │
│ ───────                                                         │
│ Monthly: $120.00 (40% reduction)                                │
│ Annual:  $1,440.00                                              │
│                                                                 │
│ REASONING                                                       │
│ ─────────                                                       │
│ • CPU utilization consistently below 15% over 14 days          │
│ • Memory utilization consistently below 10% over 14 days       │
│ • No performance degradation expected with smaller instance    │
│                                                                 │
│ Source: aws-plugin | Confidence: 92%                            │
└─────────────────────────────────────────────────────────────────┘
                        [ESC] Back  [q] Quit
```

### Loading State with Spinner

```go
type loadingModel struct {
    spinner spinner.Model
    message string
}

func initialLoadingModel() loadingModel {
    s := spinner.New()
    s.Spinner = spinner.Dot
    s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
    return loadingModel{
        spinner: s,
        message: "Fetching recommendations from plugins...",
    }
}

func (m loadingModel) View() string {
    return fmt.Sprintf("\n  %s %s\n", m.spinner.View(), m.message)
}
```

### Plain Text Fallback (CI/CD)

When not in a TTY, use Lip Gloss styling without interactivity:

```go
func renderPlainTable(results []Recommendation) error {
    // Summary section with styling
    fmt.Println(headerStyle.Render("RECOMMENDATIONS SUMMARY"))
    fmt.Println(strings.Repeat("═", 40))
    fmt.Printf("Total: %d | Savings: %s\n\n",
        len(results),
        savingsStyle.Render(formatMoney(totalSavings)))

    // Static table (no interactivity)
    for i, r := range results {
        if i >= 5 {
            fmt.Printf("\n(%d more recommendations)\n", len(results)-5)
            break
        }
        fmt.Printf("%s  %-30s  %s  %s\n",
            renderPriority(r.Priority.String()),
            r.Resource.Name,
            r.ActionType.String(),
            savingsStyle.Render(formatMoney(r.Impact.EstimatedSavings)))
    }
    return nil
}
```

### Keyboard Shortcuts

| Key | Action |
| --- | ------ |
| `↑/↓` | Navigate recommendations |
| `Enter` | View recommendation details |
| `Esc` | Back to list view |
| `/` | Filter recommendations |
| `f` | Cycle category filter |
| `p` | Cycle priority filter |
| `s` | Toggle sort (savings/priority) |
| `q` | Quit |

## External References

- [AWS Cost Explorer Rightsizing](https://docs.aws.amazon.com/cost-management/latest/userguide/ce-rightsizing.html)
- [AWS GetRightsizingRecommendation API](https://docs.aws.amazon.com/aws-cost-management/latest/APIReference/API_GetRightsizingRecommendation.html)
- [Kubecost Request Right-Sizing API V2](https://docs.kubecost.com/apis/savings-apis/api-request-right-sizing-v2)
- [Azure Advisor Recommendations API](https://learn.microsoft.com/en-us/rest/api/advisor/recommendations/list)
- [GCP Recommender API](https://cloud.google.com/recommender/docs/reference/rpc/google.cloud.recommender.v1)
- [GCP Recommender Types](https://cloud.google.com/recommender/docs/recommenders)
- [kubectl-cost CLI](https://github.com/kubecost/kubectl-cost)
