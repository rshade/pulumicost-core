---
title: CI/CD Integration
description: Integrating FinFocus into CI/CD pipelines
layout: default
---

You can run FinFocus in your CI/CD pipeline to enforce cost policies
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

      - name: Install FinFocus
        run: |
          curl -sL https://github.com/rshade/finfocus/releases/download/v0.1.0/finfocus-linux-amd64 -o finfocus
          chmod +x finfocus
          sudo mv finfocus /usr/local/bin/

      - name: Pulumi Preview
        run: pulumi preview --json > plan.json
        env:
          PULUMI_ACCESS_TOKEN: ${{ secrets.PULUMI_ACCESS_TOKEN }}

      - name: Estimate Cost
        run: finfocus cost projected --pulumi-json plan.json
```

## GitLab CI

```yaml
estimate_cost:
  stage: test
  script:
    - curl -sL https://github.com/rshade/finfocus/releases/latest/download/finfocus-linux-amd64 -o finfocus
    - chmod +x finfocus
    - pulumi preview --json > plan.json
    - ./finfocus cost projected --pulumi-json plan.json
```
