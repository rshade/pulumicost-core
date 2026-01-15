# Quickstart: Plugin Info and DryRun Discovery

## Compatibility Checks

When running `finfocus`, the system now verifies plugin compatibility. If a plugin is outdated:

```bash
$ finfocus cost projected
WARN: Plugin 'aws-public' uses spec version v0.4.0, but core expects v0.4.11+. Best-effort execution will continue.
```

To bypass these warnings (e.g., in CI):

```bash
$ finfocus cost projected --skip-version-check
```

## Listing Plugin Metadata

See detailed version information for installed plugins:

```bash
$ finfocus plugin list
NAME         VERSION   SPEC      STATUS    PROVIDERS
aws-public   v0.1.0    v0.4.11   Active    aws
```

## Inspecting Plugin Capabilities

Discover which FOCUS fields a plugin supports for a specific resource type:

```bash
$ finfocus plugin inspect aws-public aws:ec2/instance:Instance
FIELD              STATUS        CONDITION
ServiceCategory    SUPPORTED     
ChargeType         SUPPORTED     
UsageType          CONDITIONAL   Depends on instance usage data
```

Or get the raw JSON output for automation:

```bash
$ finfocus plugin inspect aws-public aws:ec2/instance:Instance --json
```
