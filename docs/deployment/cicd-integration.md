---
title: CI/CD Integration
description: Integrating Pulumicost into CI/CD pipelines
layout: default
---

You can run Pulumicost in your CI/CD pipeline to enforce cost policies
and visibility.

## GitHub Actions

```yaml
name: Cost Estimate
on: [pull_request]

jobs:
  estimate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Install Pulumicost
        run: |
          curl -sL https://github.com/rshade/pulumicost-core/releases/download/v0.1.0/pulumicost-linux-amd64 -o pulumicost
          chmod +x pulumicost
          sudo mv pulumicost /usr/local/bin/

      - name: Pulumi Preview
        run: pulumi preview --json > plan.json
        env:
          PULUMI_ACCESS_TOKEN: ${{ secrets.PULUMI_ACCESS_TOKEN }}

      - name: Estimate Cost
        run: pulumicost cost projected --pulumi-json plan.json
```

## GitLab CI

```yaml
estimate_cost:
  stage: test
  script:
    - curl -sL https://... -o pulumicost
    - chmod +x pulumicost
    - pulumi preview --json > plan.json
    - ./pulumicost cost projected --pulumi-json plan.json
```
