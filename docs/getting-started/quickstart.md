---
layout: default
title: 5-Minute Quickstart
description: Get started with PulumiCost in 5 minutes
---

# 5-Minute Quickstart

Get PulumiCost running and see your first cost estimate in just 5 minutes.

## Prerequisites

- A Pulumi project (local or existing)
- Terminal/command line access
- ~5 minutes of time

## Step 1: Install (1 minute)

### Option A: From source (recommended)

```bash
git clone https://github.com/rshade/pulumicost-core
cd pulumicost-core
make build
export PATH="$PWD/bin:$PATH"
```

### Option B: Download binary (coming soon)

```bash
# Download latest release
curl -L https://github.com/rshade/pulumicost-core/releases/latest/download/pulumicost-linux-amd64 -o pulumicost
chmod +x pulumicost
```

**Verify installation:**
```bash
pulumicost --version
```

## Step 2: Export Pulumi Plan (1 minute)

```bash
# Go to your Pulumi project
cd your-pulumi-project

# Or use the example project
cd /path/to/pulumicost-core/examples

# Export plan to JSON
pulumi preview --json > plan.json
```

## Step 3: Run PulumiCost (1 minute)

```bash
pulumicost cost projected --pulumi-json plan.json
```

**Output:**
```
RESOURCE                          TYPE              MONTHLY   CURRENCY
aws:ec2/instance:Instance         aws:ec2:Instance  $7.50     USD
aws:s3/bucket:Bucket              aws:s3:Bucket     $0.00     USD
aws:rds/instance:Instance         aws:rds:Instance  $0.00     USD

Total: $7.50 USD
```

## Step 4: Try JSON Output (1 minute)

```bash
pulumicost cost projected --pulumi-json plan.json --output json | jq .
```

**Output:**
```json
{
  "summary": {
    "totalMonthly": 7.50,
    "currency": "USD"
  },
  "resources": [
    {
      "type": "aws:ec2:Instance",
      "estimatedCost": 7.50,
      "currency": "USD"
    }
  ]
}
```

## Step 5: Try Filtering (1 minute)

```bash
# Show only EC2 resources
pulumicost cost projected --pulumi-json plan.json --filter "type=aws:ec2*"

# Show only database resources
pulumicost cost projected --pulumi-json plan.json --filter "type=aws:rds*"
```

---

## What's Next?

- **Learn more:** [User Guide](../guides/user-guide.md)
- **Installation details:** [Installation Guide](installation.md)
- **Setup with Vantage:** [Vantage Plugin Setup](../plugins/vantage/setup.md)
- **CLI reference:** [CLI Commands](../reference/cli-commands.md)
- **Examples:** [More Examples](examples/)

---

**Congratulations!** You've just run PulumiCost! 🎉
