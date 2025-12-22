# Quickstart: Running Analyzer E2E Tests

## Prerequisites

1. **Go 1.25.5+** installed
2. **Pulumi CLI** installed (`pulumi version` should work)
3. **pulumicost binary** built (`make build`)
1.  **Ensure AWS Credentials**: Ensure you have valid AWS credentials configured (via `aws configure`, environment variables, or SSO).
2.  **Build and Run**:
    ```bash
    make build
    make test-e2e
    ```

### Analyzer Tests Only

```bash
# Build binary first
make build

# Run only analyzer E2E tests
cd test/e2e
go test -v -tags e2e -run "TestAnalyzer" -timeout 30m
```

### Quick Verification

```bash
# Verify binary exists
ls -la bin/pulumicost

# Verify Pulumi CLI
pulumi version

# Dry run (skip if no Pulumi)
go test -v -tags e2e -run "TestAnalyzer" -short
```

## Test Categories

| Test Name                          | What It Validates                        |
| ---------------------------------- | ---------------------------------------- |
| `TestAnalyzer_Handshake`           | gRPC server starts, port handshake works |
| `TestAnalyzer_Diagnostics`         | Cost diagnostics appear in preview       |
| `TestAnalyzer_StackSummary`        | Total cost summary appears               |
| `TestAnalyzer_GracefulDegradation` | Errors don't block preview               |
| `TestAnalyzer_PulumiNotInstalled`  | Graceful skip when no Pulumi             |

## Environment Variables

| Variable                   | Default                | Description                |
| -------------------------- | ---------------------- | -------------------------- |
| `AWS_ACCESS_KEY_ID`        | (required)             | AWS access key             |
| `AWS_SECRET_ACCESS_KEY`    | (required)             | AWS secret key             |
| `AWS_REGION`               | `us-east-1`            | AWS region for tests       |
| `PULUMICOST_BINARY`        | `../../bin/pulumicost` | Path to pulumicost binary  |
| `PULUMI_CONFIG_PASSPHRASE` | `e2e-test`             | Passphrase for local state |
| `E2E_TIMEOUT_MINS`         | `30`                   | Test timeout in minutes    |

## Troubleshooting

### "Pulumi CLI not installed"

Install Pulumi:

```bash
curl -fsSL https://get.pulumi.com | sh
```

### "pulumicost binary not found"

Build the binary:

```bash
make build
```

### "analyzer plugin failed to start"

Check binary permissions and path:

```bash
chmod +x bin/pulumicost
./bin/pulumicost analyzer serve  # Should print port number
```

### "test timeout exceeded"

Increase timeout:

```bash
go test -v -tags e2e -timeout 60m ./...
```

## CI/CD Integration

Tests run automatically in the nightly workflow (`.github/workflows/nightly.yml`).
AWS credentials are already configured via `aws-actions/configure-aws-credentials@v4`:

```yaml
- name: Configure AWS credentials
  uses: aws-actions/configure-aws-credentials@v4
  with:
    aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
    aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
    aws-region: us-east-1

- name: Run E2E tests
  env:
    AWS_REGION: us-east-1
    E2E_REGION: us-east-1
    PULUMI_CONFIG_PASSPHRASE: ${{ secrets.PULUMI_CONFIG_PASSPHRASE || 'e2e-test' }}
  run: |
    cd test/e2e
    go test -v -tags e2e -timeout 60m ./...
```
