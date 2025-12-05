# Budgets & Alerts Feature Research & Design

> This document captures research, design decisions, and implementation
> details for the budgets and alerts feature across pulumicost-spec and
> pulumicost-core.

## Planned Issues

| Repository | Issue | Scope | Priority |
| ---------- | ----- | ----- | -------- |
| pulumicost-core | MVP: CLI budget alerts | Global scope, CLI output | P0 |
| pulumicost-core | Enhancement: Exit codes | CI/CD integration | P1 |
| pulumicost-core | Enhancement: Notifications | Webhooks, email | P2 |
| pulumicost-core | Enhancement: Flexible scoping | Per-provider, per-type | P2 |
| pulumicost-spec | GetBudgets RPC | Plugin-provided budgets | P3 |

## Cross-Provider Research Summary

### AWS Budgets

**API**: AWS Budgets API (CreateBudget, DescribeBudgets)

**Documentation**:

- [CreateBudget API](https://docs.aws.amazon.com/aws-cost-management/latest/APIReference/API_budgets_CreateBudget.html)
- [Budget Data Type](https://docs.aws.amazon.com/aws-cost-management/latest/APIReference/API_budgets_Budget.html)

**Budget Types**:

- COST - Spending limits
- USAGE - Usage limits
- RI_UTILIZATION - Reserved Instance utilization
- RI_COVERAGE - Reserved Instance coverage
- SAVINGS_PLANS_UTILIZATION - Savings Plans utilization
- SAVINGS_PLANS_COVERAGE - Savings Plans coverage

**Key Fields**:

```json
{
  "BudgetName": "string",
  "BudgetType": "COST | USAGE | RI_UTILIZATION | ...",
  "BudgetLimit": {
    "Amount": "100",
    "Unit": "USD"
  },
  "TimeUnit": "MONTHLY | QUARTERLY | ANNUALLY",
  "TimePeriod": {
    "Start": "timestamp",
    "End": "timestamp"
  },
  "FilterExpression": { ... },
  "Metrics": ["UNBLENDED_COST", "BLENDED_COST", ...],
  "NotificationsWithSubscribers": [
    {
      "Notification": {
        "NotificationType": "ACTUAL | FORECASTED",
        "ComparisonOperator": "GREATER_THAN | LESS_THAN | EQUAL_TO",
        "Threshold": 80,
        "ThresholdType": "PERCENTAGE | ABSOLUTE_VALUE"
      },
      "Subscribers": [
        {"SubscriptionType": "SNS", "Address": "arn:..."},
        {"SubscriptionType": "EMAIL", "Address": "user@example.com"}
      ]
    }
  ]
}
```

**Notification Channels**: SNS topics, Email

**Threshold Types**: ACTUAL, FORECASTED

### GCP Billing Budgets

**API**: Cloud Billing Budget API v1

**Documentation**:

- [Budget API Overview](https://cloud.google.com/billing/docs/how-to/budget-api)
- [REST Resource](https://cloud.google.com/billing/docs/reference/budget/rest/v1/billingAccounts.budgets)

**Key Fields**:

```json
{
  "name": "billingAccounts/xxx/budgets/yyy",
  "displayName": "Monthly Budget",
  "budgetFilter": {
    "projects": ["projects/123456789"],
    "services": ["services/A1E8-BE35-7EBC"],
    "creditTypesTreatment": "EXCLUDE_ALL_CREDITS",
    "calendarPeriod": "MONTH"
  },
  "amount": {
    "specifiedAmount": {
      "currencyCode": "USD",
      "units": "100"
    }
  },
  "thresholdRules": [
    {
      "thresholdPercent": 0.5,
      "spendBasis": "CURRENT_SPEND"
    },
    {
      "thresholdPercent": 0.9,
      "spendBasis": "FORECASTED_SPEND"
    }
  ],
  "allUpdatesRule": {
    "pubsubTopic": "projects/xxx/topics/budget-alerts",
    "schemaVersion": "1.0"
  }
}
```

**spendBasis Values**:

- `CURRENT_SPEND` - Compare actual spend against threshold
- `FORECASTED_SPEND` - Compare projected spend against threshold

**Notification Channels**: Pub/Sub, Email

### Azure Cost Management Budgets

**API**: Cost Management Budgets API (2025-03-01)

**Documentation**:

- [Create Or Update](https://learn.microsoft.com/en-us/rest/api/cost-management/budgets/create-or-update)
- [List Budgets](https://learn.microsoft.com/en-us/rest/api/cost-management/budgets/list)

**Key Fields**:

```json
{
  "properties": {
    "category": "Cost | ReservationUtilization",
    "amount": 100,
    "timeGrain": "Monthly | Quarterly | Annually",
    "timePeriod": {
      "startDate": "2024-01-01T00:00:00Z",
      "endDate": "2024-12-31T00:00:00Z"
    },
    "filter": {
      "dimensions": {
        "name": "ResourceId",
        "operator": "In",
        "values": ["..."]
      },
      "tags": {
        "name": "environment",
        "operator": "In",
        "values": ["production"]
      }
    },
    "notifications": {
      "actual_80_percent": {
        "enabled": true,
        "operator": "GreaterThan",
        "threshold": 80,
        "thresholdType": "Actual",
        "contactEmails": ["user@example.com"],
        "contactRoles": ["Contributor"],
        "contactGroups": ["/subscriptions/.../actionGroups/..."],
        "frequency": "Daily"
      }
    }
  }
}
```

**Threshold Types**: Actual, Forecasted

**Notification Channels**: Email, Contact Roles, Action Groups

### Kubecost Budgets

**API**: `/model/budget`, `/model/budgets`

**Documentation**:

- [Budget API](https://www.ibm.com/docs/en/kubecost/self-hosted/2.x?topic=apis-budget-api)
- [Budgets UI](https://www.ibm.com/docs/en/kubecost/self-hosted/2.x?topic=ui-budgets)

**Budget Types**:

- `allocations` - Kubernetes allocation costs
- `asset` - Cloud asset costs
- `cloud` - Cloud provider costs
- `collections` - Custom groupings

**Key Fields**:

```json
{
  "name": "my-budget",
  "budgetType": "allocations",
  "values": {
    "namespace": ["production", "staging"]
  },
  "interval": "monthly",
  "intervalDay": 1,
  "spendLimit": 1000,
  "actions": [
    {
      "percentage": 80,
      "emails": ["user@example.com"],
      "slackWebhooks": ["https://hooks.slack.com/..."],
      "msTeamsWebhooks": ["https://..."]
    }
  ]
}
```

**Response with Current Spend**:

```json
{
  "code": 200,
  "data": [{
    "name": "my-budget",
    "id": "abc123",
    "spendLimit": 1000,
    "currentSpend": 850,
    "window": {
      "start": "2024-01-01T00:00:00Z",
      "end": "2024-01-31T23:59:59Z"
    },
    "actions": [
      {
        "percentage": 80,
        "lastFired": "2024-01-15T10:30:00Z"
      }
    ]
  }]
}
```

**Intervals**: weekly, monthly

**Notification Channels**: Email, Slack, Microsoft Teams

**Supported Currencies**: USD, EUR, GBP, JPY, INR, AUD, CAD, BRL, CHF, CNY,
DKK, IDR, NOK, PLN, SEK

### Flexera FinOps

**API**: Cloud Cost Optimization / BudgetAlerts API

**Documentation**:

- [CCO APIs](https://docs.flexera.com/flexera/EN/Optima/CCOAPIs.htm)
- [Cost Planning](https://docs.flexera.com/flexera/EN/Optima/cloudbudgets.htm)

**Features**:

- Monthly spend budgets
- Actual or forecasted spend alerts
- Email notifications
- Cost planning with dimensions hierarchy

**Note**: Flexera is transitioning from legacy Optima APIs to new Flexera
One APIs.

## Design Decisions

### User Requirements Captured

1. **Budget Source**: Hybrid (both local config and plugin-provided)
2. **Alert Output**: Phased approach
   - MVP: CLI output only
   - Enhancement 1: Exit codes (config/env var driven)
   - Enhancement 2: Webhooks/email notifications (config driven)
3. **Budget Scope**: Start with global, add flexible scoping later
4. **Threshold Types**: Both Actual and Forecasted
5. **Spec Approach**: Create spec issue for GetBudgets RPC, implement CLI first

### MVP Configuration (Global Scope)

```yaml
# ~/.pulumicost/config.yaml
cost:
  budgets:
    amount: 100
    currency: USD  # optional, defaults to USD
    period: monthly  # optional, defaults to monthly
    alerts:
      - threshold: 80
        type: actual  # actual or forecasted
      - threshold: 100
        type: actual
```

### Enhanced Configuration (Flexible Scoping - Future)

```yaml
# ~/.pulumicost/config.yaml
cost:
  budgets:
    global:
      amount: 500
      alerts:
        - threshold: 80
          type: actual
    providers:
      aws:
        amount: 200
        alerts:
          - threshold: 90
            type: forecasted
      gcp:
        amount: 150
    resource_types:
      aws:ec2/instance:
        amount: 100
```

### Exit Codes Configuration (Enhancement 1)

```yaml
# ~/.pulumicost/config.yaml
cost:
  budgets:
    amount: 100
    alerts:
      - threshold: 80
        type: actual
    exit_on_threshold: true  # non-zero exit when threshold exceeded
    exit_code: 2  # optional, defaults to 1
```

Or via environment variable:

```bash
export PULUMICOST_BUDGET_EXIT_ON_THRESHOLD=true
export PULUMICOST_BUDGET_EXIT_CODE=2
```

### Notifications Configuration (Enhancement 2)

```yaml
# ~/.pulumicost/config.yaml
cost:
  budgets:
    amount: 100
    alerts:
      - threshold: 80
        type: actual
        notifications:
          - type: slack
            webhook: "${SLACK_WEBHOOK_URL}"
          - type: email
            to: ["alerts@example.com"]
          - type: webhook
            url: "https://api.example.com/budget-alert"
            method: POST
            headers:
              Authorization: "Bearer ${API_TOKEN}"
```

## CLI Output Design

### MVP Output (Inline with Cost Results)

```text
COST SUMMARY
============
Total Monthly Cost:  $85.00 USD
Total Resources:     12

BUDGET STATUS
=============
Budget: $100.00/month
Current Spend: $85.00 (85%)
Status: ⚠ WARNING - Exceeds 80% threshold

TOP RECOMMENDATIONS
===================
...
```

### With Forecasted Alert

```text
BUDGET STATUS
=============
Budget: $100.00/month
Current Spend: $45.00 (45%)
Forecasted Spend: $92.00 (92%)
Status: ⚠ WARNING - Forecast exceeds 80% threshold
```

### Multiple Thresholds Exceeded

```text
BUDGET STATUS
=============
Budget: $100.00/month
Current Spend: $105.00 (105%)
Status: 🚨 CRITICAL - Budget exceeded!

Thresholds:
  ✓ 50% ($50.00) - OK
  ⚠ 80% ($80.00) - Exceeded at $85.00
  🚨 100% ($100.00) - Exceeded at $105.00
```

## Spec Design (GetBudgets RPC - Future)

### Proto Schema

```protobuf
// Budget represents a spending limit with alert thresholds
message Budget {
  string id = 1;
  string name = 2;
  string source = 3;  // "aws-budgets", "gcp-billing", "kubecost", etc.

  BudgetAmount amount = 4;
  BudgetPeriod period = 5;
  BudgetFilter filter = 6;

  repeated BudgetThreshold thresholds = 7;
  BudgetStatus status = 8;

  google.protobuf.Timestamp created_at = 9;
  google.protobuf.Timestamp updated_at = 10;
  map<string, string> metadata = 11;
}

message BudgetAmount {
  double limit = 1;
  string currency = 2;  // ISO 4217
}

enum BudgetPeriod {
  BUDGET_PERIOD_UNSPECIFIED = 0;
  BUDGET_PERIOD_DAILY = 1;
  BUDGET_PERIOD_WEEKLY = 2;
  BUDGET_PERIOD_MONTHLY = 3;
  BUDGET_PERIOD_QUARTERLY = 4;
  BUDGET_PERIOD_ANNUALLY = 5;
}

message BudgetFilter {
  repeated string providers = 1;
  repeated string regions = 2;
  repeated string resource_types = 3;
  map<string, string> tags = 4;
}

message BudgetThreshold {
  double percentage = 1;  // 0-100
  ThresholdType type = 2;
  bool triggered = 3;
  google.protobuf.Timestamp triggered_at = 4;
}

enum ThresholdType {
  THRESHOLD_TYPE_UNSPECIFIED = 0;
  THRESHOLD_TYPE_ACTUAL = 1;
  THRESHOLD_TYPE_FORECASTED = 2;
}

message BudgetStatus {
  double current_spend = 1;
  double forecasted_spend = 2;
  double percentage_used = 3;
  double percentage_forecasted = 4;
  string currency = 5;
  BudgetHealthStatus health = 6;
}

enum BudgetHealthStatus {
  BUDGET_HEALTH_UNSPECIFIED = 0;
  BUDGET_HEALTH_OK = 1;
  BUDGET_HEALTH_WARNING = 2;
  BUDGET_HEALTH_CRITICAL = 3;
  BUDGET_HEALTH_EXCEEDED = 4;
}

// Request/Response
message GetBudgetsRequest {
  BudgetFilter filter = 1;
  bool include_status = 2;  // Fetch current spend status
}

message GetBudgetsResponse {
  repeated Budget budgets = 1;
  BudgetSummary summary = 2;
}

message BudgetSummary {
  int32 total_budgets = 1;
  int32 budgets_ok = 2;
  int32 budgets_warning = 3;
  int32 budgets_exceeded = 4;
}
```

### PluginSDK Interface

```go
// BudgetsProvider is an optional interface that plugins can implement
// to provide budget information from cloud cost management services.
type BudgetsProvider interface {
    GetBudgets(ctx context.Context, req *pbc.GetBudgetsRequest) (
        *pbc.GetBudgetsResponse, error)
}
```

## Implementation File Structure

### pulumicost-core MVP

```text
internal/
├── cli/
│   └── cost_budget.go       # Budget display in CLI output
├── config/
│   └── budget.go            # Budget configuration parsing
└── engine/
    └── budget.go            # Budget comparison logic
```

### pulumicost-core Enhancements

```text
internal/
├── budget/
│   ├── evaluator.go         # Threshold evaluation
│   ├── formatter.go         # Output formatting
│   └── notifier.go          # Notification sending (Enhancement 2)
└── config/
    └── budget.go            # Extended config with scoping
```

### pulumicost-spec (Future)

```text
proto/pulumicost/v1/
├── budget.proto             # Budget messages
└── costsource.proto         # Add GetBudgets RPC

sdk/go/
├── pluginsdk/
│   └── sdk.go               # Add BudgetsProvider interface
└── testing/
    └── mock_plugin.go       # Add GetBudgets mock
```

## Testing Strategy

### Unit Tests

- Config parsing for budget YAML
- Threshold evaluation logic
- Output formatting
- Exit code behavior

### Integration Tests

- End-to-end CLI with budget config
- Multiple threshold scenarios
- Forecasted vs actual spend comparison

### Edge Cases

- No budget configured (skip display)
- Zero budget amount
- Negative spend (refunds)
- Mixed currencies (warning)
- Missing current spend data

## Bubble Tea TUI Integration

### Dependencies

```go
import (
    "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/bubbles/progress"
    "github.com/charmbracelet/lipgloss"
    "golang.org/x/term"
)
```

### Output Mode Router

```go
func renderBudgetStatus(budget BudgetConfig, currentSpend float64) error {
    isTTY := term.IsTerminal(int(os.Stdout.Fd()))

    if isTTY {
        // Styled output with progress bar (Lip Gloss only - no interactivity)
        return renderStyledBudget(budget, currentSpend)
    }
    // Plain text for CI/CD
    return renderPlainBudget(budget, currentSpend)
}
```

### Styled Budget Status with Lip Gloss

```go
var (
    // Status styles
    okStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("82")).
        Bold(true)

    warningStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("208")).
        Bold(true)

    criticalStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("196")).
        Bold(true)

    // Section styles
    headerStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("99"))

    labelStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("245"))

    valueStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("255"))

    // Box style for budget section
    budgetBoxStyle = lipgloss.NewStyle().
        BorderStyle(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("238")).
        Padding(0, 1)
)
```

### Progress Bar Rendering

```go
func renderBudgetProgressBar(percent float64, width int) string {
    filled := int(percent / 100 * float64(width))
    if filled > width {
        filled = width
    }

    // Choose color based on percentage
    var barStyle lipgloss.Style
    switch {
    case percent >= 100:
        barStyle = criticalStyle
    case percent >= 80:
        barStyle = warningStyle
    default:
        barStyle = okStyle
    }

    bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
    return barStyle.Render(bar)
}

func renderStyledBudget(budget BudgetConfig, spend, forecast float64) string {
    percentUsed := (spend / budget.Amount) * 100
    percentForecast := (forecast / budget.Amount) * 100

    var status string
    switch {
    case percentUsed >= 100:
        status = criticalStyle.Render("🚨 CRITICAL - Budget exceeded!")
    case percentUsed >= 80 || percentForecast >= 80:
        if percentForecast >= 80 && percentUsed < 80 {
            status = warningStyle.Render("⚠ WARNING - Forecast exceeds threshold")
        } else {
            status = warningStyle.Render("⚠ WARNING - Exceeds 80% threshold")
        }
    default:
        status = okStyle.Render("✓ OK - Within budget")
    }

    content := fmt.Sprintf(`%s
%s

Budget: %s
Current Spend: %s (%s)
%s

%s`,
        headerStyle.Render("BUDGET STATUS"),
        strings.Repeat("─", 40),
        valueStyle.Render(formatMoney(budget.Amount)+"/"+budget.Period),
        valueStyle.Render(formatMoney(spend)),
        formatPercent(percentUsed),
        renderBudgetProgressBar(percentUsed, 30),
        status,
    )

    return budgetBoxStyle.Render(content)
}
```

### Enhanced Visual Output (TTY Mode)

```text
╭──────────────────────────────────────────╮
│ BUDGET STATUS                            │
│ ────────────────────────────────────────│
│                                          │
│ Budget: $100.00/month                    │
│ Current Spend: $85.00 (85%)              │
│                                          │
│   █████████████████████████░░░░░  85%    │
│                                          │
│ ⚠ WARNING - Exceeds 80% threshold        │
│                                          │
│ Thresholds:                              │
│   ✓ 50% ─────────────── OK               │
│   ⚠ 80% ─────────────── EXCEEDED         │
│   ○ 100% ────────────── approaching      │
╰──────────────────────────────────────────╯
```

### With Forecasted Spend

```text
╭──────────────────────────────────────────╮
│ BUDGET STATUS                            │
│ ────────────────────────────────────────│
│                                          │
│ Budget: $100.00/month                    │
│                                          │
│ Current:   $45.00 (45%)                  │
│   █████████████░░░░░░░░░░░░░░░░  45%     │
│                                          │
│ Forecast:  $92.00 (92%)                  │
│   ███████████████████████████░░  92%     │
│                                          │
│ ⚠ WARNING - Forecast exceeds threshold   │
╰──────────────────────────────────────────╯
```

### Threshold Status Rendering

```go
func renderThresholdStatus(thresholds []ThresholdConfig, spend float64,
    budget float64) string {
    var lines []string

    for _, t := range thresholds {
        thresholdValue := budget * (t.Percentage / 100)
        exceeded := spend >= thresholdValue

        var icon, status string
        var style lipgloss.Style

        switch {
        case exceeded && t.Percentage >= 100:
            icon = "🚨"
            status = "EXCEEDED"
            style = criticalStyle
        case exceeded:
            icon = "⚠"
            status = "EXCEEDED"
            style = warningStyle
        case spend >= thresholdValue*0.9:
            icon = "◉"
            status = "approaching"
            style = lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
        default:
            icon = "✓"
            status = "OK"
            style = okStyle
        }

        line := fmt.Sprintf("  %s %.0f%% %s %s",
            icon,
            t.Percentage,
            strings.Repeat("─", 15),
            style.Render(status))
        lines = append(lines, line)
    }

    return strings.Join(lines, "\n")
}
```

### Plain Text Fallback (CI/CD)

```go
func renderPlainBudget(budget BudgetConfig, spend float64) string {
    percentUsed := (spend / budget.Amount) * 100

    var status string
    switch {
    case percentUsed >= 100:
        status = "CRITICAL - Budget exceeded!"
    case percentUsed >= 80:
        status = "WARNING - Exceeds 80% threshold"
    default:
        status = "OK - Within budget"
    }

    return fmt.Sprintf(`BUDGET STATUS
=============
Budget: $%.2f/%s
Current Spend: $%.2f (%.0f%%)
Status: %s
`,
        budget.Amount,
        budget.Period,
        spend,
        percentUsed,
        status)
}
```

### Integration with Cost Summary

The budget status integrates into the main cost output flow:

```go
func renderCostSummary(results CostResults, budget *BudgetConfig) error {
    isTTY := term.IsTerminal(int(os.Stdout.Fd()))

    // Render cost summary first
    if isTTY {
        fmt.Println(renderStyledCostSummary(results))
    } else {
        fmt.Println(renderPlainCostSummary(results))
    }

    // Render budget status if configured
    if budget != nil {
        fmt.Println() // spacing
        if isTTY {
            fmt.Println(renderStyledBudget(*budget, results.TotalCost))
        } else {
            fmt.Println(renderPlainBudget(*budget, results.TotalCost))
        }
    }

    return nil
}
```

### Color Scheme Reference

| Element | Color Code | Hex | Usage |
| ------- | ---------- | --- | ----- |
| OK/Green | 82 | #5fd700 | Under budget, passed thresholds |
| Warning/Orange | 208 | #ff8700 | 80%+ threshold exceeded |
| Critical/Red | 196 | #ff0000 | 100%+ budget exceeded |
| Header | 99 | #875fff | Section headers |
| Label | 245 | #8a8a8a | Field labels |
| Value | 255 | #eeeeee | Field values |
| Border | 238 | #444444 | Box borders |

## External References

- [AWS Budgets API](https://docs.aws.amazon.com/aws-cost-management/latest/APIReference/API_Operations_AWS_Budgets.html)
- [GCP Cloud Billing Budget API](https://cloud.google.com/billing/docs/how-to/budget-api)
- [Azure Cost Management Budgets](https://learn.microsoft.com/en-us/rest/api/cost-management/budgets)
- [Kubecost Budget API](https://www.ibm.com/docs/en/kubecost/self-hosted/2.x?topic=apis-budget-api)
- [Flexera Cost Planning](https://docs.flexera.com/flexera/EN/Optima/cloudbudgets.htm)
