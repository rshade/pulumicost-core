# Troubleshooting Guide

Common issues, solutions, and debugging techniques for PulumiCost Core.

## Table of Contents

- [Installation Issues](#installation-issues)
- [Pulumi Integration](#pulumi-integration)
- [Plugin Problems](#plugin-problems)
- [Cost Calculation Issues](#cost-calculation-issues)
- [Performance Problems](#performance-problems)
- [Configuration Issues](#configuration-issues)
- [Network and Authentication](#network-and-authentication)
- [Debug Mode](#debug-mode)
- [Getting Help](#getting-help)

## Installation Issues

### Binary Not Found

**Problem**: Command `pulumicost` not found after installation.

```bash
$ pulumicost --help
bash: pulumicost: command not found
```

**Solutions**:

1. **Check if binary is in PATH**:
   ```bash
   which pulumicost
   echo $PATH
   ```

2. **Add to PATH temporarily**:
   ```bash
   export PATH=$PATH:/usr/local/bin
   ```

3. **Add to PATH permanently**:
   ```bash
   # For bash
   echo 'export PATH=$PATH:/usr/local/bin' >> ~/.bashrc
   source ~/.bashrc
   
   # For zsh
   echo 'export PATH=$PATH:/usr/local/bin' >> ~/.zshrc
   source ~/.zshrc
   ```

4. **Use full path**:
   ```bash
   /usr/local/bin/pulumicost --help
   ```

### Permission Denied

**Problem**: Permission denied when running binary.

```bash
$ pulumicost --help
bash: ./pulumicost: Permission denied
```

**Solutions**:

1. **Make binary executable**:
   ```bash
   chmod +x /usr/local/bin/pulumicost
   ```

2. **Check file permissions**:
   ```bash
   ls -la /usr/local/bin/pulumicost
   # Should show: -rwxr-xr-x (executable permissions)
   ```

3. **For downloaded files**:
   ```bash
   chmod +x ./pulumicost
   sudo mv ./pulumicost /usr/local/bin/
   ```

### macOS Security Warning

**Problem**: macOS blocks unsigned binary.

```
"pulumicost" cannot be opened because the developer cannot be verified.
```

**Solutions**:

1. **Allow via System Preferences**:
   - Go to System Preferences > Security & Privacy
   - Click "Allow Anyway" for pulumicost

2. **Command line bypass**:
   ```bash
   sudo spctl --add /usr/local/bin/pulumicost
   sudo xattr -dr com.apple.quarantine /usr/local/bin/pulumicost
   ```

3. **Temporary bypass**:
   ```bash
   xattr -dr com.apple.quarantine ./pulumicost
   ./pulumicost --help
   ```

### Version Compatibility

**Problem**: Binary doesn't work on your system architecture.

**Solutions**:

1. **Check system architecture**:
   ```bash
   uname -m
   # x86_64 = amd64, aarch64 = arm64
   ```

2. **Download correct binary**:
   ```bash
   # For Intel Macs
   curl -L https://github.com/rshade/pulumicost-core/releases/latest/download/pulumicost-darwin-amd64 -o pulumicost
   
   # For Apple Silicon Macs
   curl -L https://github.com/rshade/pulumicost-core/releases/latest/download/pulumicost-darwin-arm64 -o pulumicost
   ```

## Pulumi Integration

### Pulumi Plan JSON Issues

**Problem**: `pulumi preview --json` fails or produces invalid JSON.

**Diagnosis**:
```bash
# Test Pulumi JSON output
pulumi preview --json > plan.json
cat plan.json | jq '.' # Should parse without errors
```

**Solutions**:

1. **Fix malformed JSON**:
   ```bash
   # Check JSON validity
   python -m json.tool plan.json
   # or
   jq '.' plan.json
   ```

2. **Pulumi authentication issues**:
   ```bash
   pulumi login
   pulumi stack select your-stack-name
   ```

3. **Resource provider issues**:
   ```bash
   # Update providers
   pulumi plugin install
   
   # List available providers
   pulumi plugin ls
   ```

4. **Use stack export as alternative**:
   ```bash
   # If preview fails, try current state
   pulumi stack export > current-state.json
   pulumicost cost projected --pulumi-json current-state.json
   ```

### Empty Pulumi Plan

**Problem**: Pulumi plan contains no resources.

```bash
$ pulumicost cost projected --pulumi-json plan.json
No resources found in Pulumi plan
```

**Solutions**:

1. **Check plan contents**:
   ```bash
   jq '.steps | length' plan.json
   jq '.steps[0]' plan.json
   ```

2. **Verify Pulumi stack**:
   ```bash
   pulumi stack ls
   pulumi stack select correct-stack
   pulumi preview
   ```

3. **Check for pending changes**:
   ```bash
   pulumi up --diff
   ```

### Resource Type Recognition

**Problem**: Resources not recognized by PulumiCost.

```
Resource type 'custom:provider/resource:Type' not supported
```

**Solutions**:

1. **Check supported resource types**:
   ```bash
   # Common supported patterns:
   # aws:ec2/instance:Instance
   # aws:s3/bucket:Bucket
   # aws:rds/instance:Instance
   ```

2. **Plugin Compatibility**:
   Some plugins may require specific resource type formats (e.g., `aws:ec2:Instance` vs `aws:ec2/instance:Instance`). 
   - Check plugin documentation for supported types.
   - Ensure your Pulumi provider versions are compatible with the plugin.

3. **Use resource filtering**:
   ```bash
   # Filter to supported resources only
   pulumicost cost projected --pulumi-json plan.json --filter "type=aws:ec2"
   ```

4. **Create custom pricing spec**:
   ```bash
   mkdir -p ~/.pulumicost/specs
   # Create YAML spec for custom resource type
   ```

## Plugin Problems

### Plugin Not Found

**Problem**: Plugins not discovered or loaded.

```bash
$ pulumicost plugin list
No plugins found
```

**Solutions**:

1. **Check plugin directory structure**:
   ```bash
   ls -la ~/.pulumicost/plugins/
   ls -la ~/.pulumicost/plugins/*/*/
   ```

2. **Verify directory structure**:
   ```
   ~/.pulumicost/plugins/
   └── kubecost/
       └── 1.0.0/
           └── pulumicost-kubecost
   ```

3. **Make plugin executable**:
   ```bash
   chmod +x ~/.pulumicost/plugins/*/*/pulumicost-*
   ```

4. **Check plugin names**:
   ```bash
   # Plugin binary must start with 'pulumicost-'
   ls ~/.pulumicost/plugins/*/*/pulumicost-*
   ```

### Plugin Validation Failures

**Problem**: Plugin validation fails.

```bash
$ pulumicost plugin validate
kubecost: FAILED - connection timeout
```

**Solutions**:

1. **Test plugin directly**:
   ```bash
   ~/.pulumicost/plugins/kubecost/1.0.0/pulumicost-kubecost
   # Should start and show port information
   ```

2. **Check plugin logs**:
   ```bash
   PLUGIN_DEBUG=1 pulumicost plugin validate --adapter kubecost
   ```

3. **Network connectivity**:
   ```bash
   # Test external API connectivity
   curl -v http://kubecost.example.com:9090/api/v1/costDataModel
   ```

4. **Authentication setup**:
   ```bash
   # Verify required environment variables
   env | grep -E "(KUBECOST|AWS|AZURE|GCP)"
   ```

### Plugin Communication Errors

**Problem**: gRPC communication failures between core and plugins.

```
Error: failed to connect to plugin: context deadline exceeded
```

**Solutions**:

1. **Increase timeout**:
   ```bash
   export PLUGIN_TIMEOUT=60s
   pulumicost cost actual --adapter kubecost --from 2025-01-01
   ```

2. **Check port conflicts**:
   ```bash
   # List processes using common gRPC ports
   lsof -i :50051
   netstat -tulpn | grep :50051
   ```

3. **Firewall issues**:
   ```bash
   # Temporarily disable firewall for testing
   sudo ufw disable  # Ubuntu
   sudo systemctl stop firewalld  # CentOS/RHEL
   ```

4. **Plugin process issues**:
   ```bash
   # Kill any stuck plugin processes
   pkill pulumicost-kubecost
   ps aux | grep pulumicost
   ```

## Cost Calculation Issues

### No Cost Data Available

**Problem**: All resources show "$0.00" or "No pricing information available".

```
aws:ec2/instance:Instance    none    $0.00    USD    No pricing information available
```

**Solutions**:

1. **Install pricing plugins**:
   ```bash
   # Install appropriate cost plugins
   # See Plugin System documentation
   ```

2. **Create local pricing specs**:
   ```bash
   mkdir -p ~/.pulumicost/specs
   cat > ~/.pulumicost/specs/aws-ec2-t3-micro.yaml << 'EOF'
   provider: aws
   service: ec2
   sku: t3.micro
   currency: USD
   pricing:
     onDemandHourly: 0.0104
     monthlyEstimate: 7.59
   EOF
   ```

3. **Use specific adapter**:
   ```bash
   pulumicost cost projected --pulumi-json plan.json --adapter aws-plugin
   ```

4. **Check resource properties**:
   ```bash
   # Verify resource has necessary properties
   jq '.steps[0].inputs' plan.json
   ```

### Missing Cost Data (Empty Inputs)

**Problem**: Logs show "resource descriptor missing required fields (sku, region)" or plugins return "not supported" because properties are missing.

**Diagnosis**:
This often happens when `pulumi preview --json` structure changes (e.g., nesting inputs under `newState`).

**Solution**:
Ensure you are using a compatible version of `pulumicost` that handles the JSON structure of your Pulumi CLI version.
- Update `pulumicost` to the latest version.
- Check `pulumi version` and ensure compatibility.

### Inaccurate Cost Estimates

**Problem**: Cost estimates seem too high or too low.

**Diagnosis**:

1. **Compare with manual calculations**:
   ```bash
   # For t3.micro: $0.0104/hour × 730 hours = $7.59/month
   ```

2. **Check pricing spec accuracy**:
   ```bash
   cat ~/.pulumicost/specs/aws-ec2-t3-micro.yaml
   ```

3. **Verify resource configuration**:
   ```bash
   jq '.steps[] | select(.type == "aws:ec2/instance:Instance") | .inputs' plan.json
   ```

**Solutions**:

1. **Update pricing specifications**:
   ```bash
   # Get latest AWS pricing
   # Update local specs accordingly
   ```

2. **Use region-specific pricing**:
   ```bash
   # Different regions have different costs
   # Ensure specs match your deployment region
   ```

3. **Account for reserved instances**:
   ```bash
   # Pricing specs should reflect your actual pricing model
   # (on-demand vs reserved vs spot)
   ```

### Historical Cost Data Issues

**Problem**: Actual cost queries return no data or errors.

```bash
$ pulumicost cost actual --pulumi-json plan.json --from 2025-01-01
Error: no cost data available for time range
```

**Solutions**:

1. **Check date range**:
   ```bash
   # Ensure date range is valid and not too recent
   pulumicost cost actual --pulumi-json plan.json --from 2025-01-07 --to 2025-01-14
   ```

2. **Billing data lag**:
   ```bash
   # Most billing APIs have 24-48 hour delays
   # Try querying older date ranges
   pulumicost cost actual --pulumi-json plan.json --from 2025-01-01 --to 2025-01-02
   ```

3. **Resource matching**:
   ```bash
   # Resources might not exist in the historical time range
   # Check when resources were actually deployed
   ```

4. **Plugin configuration**:
   ```bash
   # Verify plugin can access billing APIs
   export KUBECOST_API_URL="http://kubecost.example.com:9090"
   pulumicost plugin validate --adapter kubecost
   ```

## Performance Problems

### Slow Cost Calculations

**Problem**: Cost calculations take too long to complete.

**Solutions**:

1. **Use specific adapter**:
   ```bash
   # Avoid querying all plugins
   pulumicost cost projected --pulumi-json plan.json --adapter kubecost
   ```

2. **Filter resources**:
   ```bash
   # Process fewer resources
   pulumicost cost projected --pulumi-json plan.json --filter "type=aws:ec2"
   ```

3. **Increase timeout**:
   ```bash
   export PLUGIN_TIMEOUT=300s
   pulumicost cost actual --pulumi-json plan.json --from 2025-01-01
   ```

4. **Parallel processing**:
   ```bash
   # Split large plans into smaller chunks
   jq '.steps[:10]' plan.json > plan-chunk1.json
   jq '.steps[10:20]' plan.json > plan-chunk2.json
   ```

### Memory Usage

**Problem**: High memory usage with large Pulumi plans.

**Solutions**:

1. **Process in batches**:
   ```bash
   # Split large plans
   jq '.steps | .[0:100]' large-plan.json > batch1.json
   jq '.steps | .[100:200]' large-plan.json > batch2.json
   ```

2. **Use NDJSON output**:
   ```bash
   # More memory-efficient for large datasets
   pulumicost cost projected --pulumi-json plan.json --output ndjson
   ```

3. **Filter early**:
   ```bash
   # Reduce processing load
   pulumicost cost projected --pulumi-json plan.json --filter "provider=aws"
   ```

### Network Timeouts

**Problem**: Requests to external APIs timeout.

**Solutions**:

1. **Check network connectivity**:
   ```bash
   curl -v https://api.aws.com/
   ping kubecost.example.com
   ```

2. **Increase timeouts**:
   ```bash
   export HTTP_TIMEOUT=60s
   export PLUGIN_TIMEOUT=300s
   ```

3. **Configure proxy**:
   ```bash
   export HTTP_PROXY=http://proxy.company.com:8080
   export HTTPS_PROXY=http://proxy.company.com:8080
   ```

4. **Retry configuration**:
   ```bash
   export MAX_RETRIES=5
   export RETRY_DELAY=5s
   ```

## Configuration Issues

### Directory Permissions

**Problem**: Cannot create or access configuration directories.

**Solutions**:

1. **Create directories manually**:
   ```bash
   mkdir -p ~/.pulumicost/plugins
   mkdir -p ~/.pulumicost/specs
   chmod 755 ~/.pulumicost/
   ```

2. **Fix ownership**:
   ```bash
   chown -R $USER:$USER ~/.pulumicost/
   ```

3. **Check disk space**:
   ```bash
   df -h ~
   ```

### Environment Variables

**Problem**: Environment variables not being recognized.

**Solutions**:

1. **Check variable names**:
   ```bash
   env | grep PULUMICOST
   env | grep -E "(AWS|KUBECOST|AZURE)"
   ```

2. **Export variables properly**:
   ```bash
   export KUBECOST_API_URL="http://kubecost.example.com:9090"
   export AWS_REGION="us-west-2"
   ```

3. **Persistent environment setup**:
   ```bash
   # Add to shell profile
   echo 'export KUBECOST_API_URL="http://kubecost.example.com:9090"' >> ~/.bashrc
   source ~/.bashrc
   ```

### Configuration Files

**Problem**: Configuration files not being loaded.

**Solutions**:

1. **Check file locations**:
   ```bash
   ls -la ~/.pulumicost/
   ls -la ~/.pulumicost/plugins/*/*/config.*
   ```

2. **Validate YAML syntax**:
   ```bash
   python -c "import yaml; yaml.safe_load(open('config.yaml'))"
   # or
   yq eval '.' config.yaml
   ```

3. **Check file permissions**:
   ```bash
   ls -la ~/.pulumicost/specs/*.yaml
   chmod 644 ~/.pulumicost/specs/*.yaml
   ```

## Network and Authentication

### API Authentication Failures

**Problem**: Plugin cannot authenticate with external APIs.

```
Error: authentication failed: invalid API key
```

**Solutions**:

1. **Verify credentials**:
   ```bash
   # Test API credentials manually
   curl -H "Authorization: Bearer $API_TOKEN" https://api.provider.com/test
   ```

2. **Check credential format**:
   ```bash
   # Ensure no extra whitespace or newlines
   echo -n "$API_KEY" | hexdump -C
   ```

3. **Rotate credentials**:
   ```bash
   # Generate new API keys if current ones are invalid
   ```

4. **Check credential permissions**:
   ```bash
   # Ensure API key has required permissions for cost data access
   ```

### SSL Certificate Issues

**Problem**: SSL certificate verification failures.

```
Error: x509: certificate signed by unknown authority
```

**Solutions**:

1. **Update CA certificates**:
   ```bash
   # Ubuntu/Debian
   sudo apt-get update && sudo apt-get install ca-certificates
   
   # CentOS/RHEL
   sudo yum update ca-certificates
   ```

2. **Disable SSL verification (temporary)**:
   ```bash
   export INSECURE_SKIP_VERIFY=true
   ```

3. **Add custom CA certificate**:
   ```bash
   # Add your organization's CA certificate
   sudo cp custom-ca.crt /usr/local/share/ca-certificates/
   sudo update-ca-certificates
   ```

### Proxy Configuration

**Problem**: Network requests fail behind corporate proxy.

**Solutions**:

1. **Configure HTTP proxy**:
   ```bash
   export HTTP_PROXY=http://proxy.company.com:8080
   export HTTPS_PROXY=http://proxy.company.com:8080
   export NO_PROXY=localhost,127.0.0.1,.internal.domain
   ```

2. **Authenticated proxy**:
   ```bash
   export HTTP_PROXY=http://username:password@proxy.company.com:8080
   ```

3. **Plugin-specific proxy**:
   ```bash
   # Some plugins might need specific proxy configuration
   export KUBECOST_PROXY_URL=http://proxy.company.com:8080
   ```

## Debug Mode

### Enable Debug Logging

```bash
# Global debug mode
pulumicost --debug cost projected --pulumi-json plan.json

# Plugin-specific debugging
export PLUGIN_DEBUG=1
export PLUGIN_LOG_LEVEL=debug
pulumicost cost actual --adapter kubecost --from 2025-01-01
```

### Verbose Output

```bash
# Increase verbosity
pulumicost -v cost projected --pulumi-json plan.json

# Maximum verbosity
pulumicost -vv cost actual --adapter kubecost --from 2025-01-01
```

### Log Analysis

```bash
# Save logs to file
pulumicost --debug cost projected --pulumi-json plan.json > debug.log 2>&1

# Search for specific errors
grep -i "error\|failed\|timeout" debug.log

# Check plugin communication
grep -i "grpc\|plugin\|connect" debug.log
```

### Plugin Debugging

```bash
# Test plugin directly
~/.pulumicost/plugins/kubecost/1.0.0/pulumicost-kubecost &
PLUGIN_PID=$!

# Test gRPC connection
grpcurl -plaintext localhost:50051 pulumicost.v1.CostSourceService/Name

# Clean up
kill $PLUGIN_PID
```

## Common Error Messages

### "No such file or directory"

```bash
Error: exec: "pulumicost-kubecost": executable file not found in $PATH
```

**Solution**: Plugin binary missing or not executable.

```bash
ls -la ~/.pulumicost/plugins/kubecost/1.0.0/
chmod +x ~/.pulumicost/plugins/kubecost/1.0.0/pulumicost-kubecost
```

### "Connection refused"

```bash
Error: connection refused: dial tcp 127.0.0.1:50051: connect: connection refused
```

**Solution**: Plugin process not running or wrong port.

```bash
# Restart plugin validation
pulumicost plugin validate
# Check for port conflicts
lsof -i :50051
```

### "Context deadline exceeded"

```bash
Error: context deadline exceeded
```

**Solution**: Increase timeout or check network connectivity.

```bash
export PLUGIN_TIMEOUT=60s
# Check network connectivity to external APIs
```

### "Invalid JSON"

```bash
Error: invalid character 'N' looking for beginning of value
```

**Solution**: Pulumi plan JSON is malformed.

```bash
# Re-generate plan
pulumi preview --json > plan.json
# Validate JSON
jq '.' plan.json
```

## Getting Help

### Community Support

1. **GitHub Issues**: https://github.com/rshade/pulumicost-core/issues
2. **Discussions**: https://github.com/rshade/pulumicost-core/discussions
3. **Documentation**: Check all docs in this repository

### Bug Reports

When reporting bugs, include:

1. **Version information**:
   ```bash
   pulumicost --version
   ```

2. **System information**:
   ```bash
   uname -a
   go version  # if building from source
   ```

3. **Debug logs**:
   ```bash
   pulumicost --debug [command] > debug.log 2>&1
   ```

4. **Configuration details**:
   ```bash
   ls -la ~/.pulumicost/
   env | grep -E "(PULUMI|AWS|KUBECOST)"
   ```

5. **Minimal reproduction**:
   - Steps to reproduce the issue
   - Expected vs actual behavior
   - Sample Pulumi plan (sanitized)

### Feature Requests

1. Check existing issues/discussions first
2. Describe the use case clearly
3. Provide examples of desired behavior
4. Consider contributing implementation

### Self-Help Checklist

Before seeking help:

- [ ] Check this troubleshooting guide
- [ ] Try with `--debug` flag
- [ ] Verify installation with `pulumicost --version`
- [ ] Test with provided examples
- [ ] Check plugin status with `pulumicost plugin validate`
- [ ] Review environment variables and configuration
- [ ] Test network connectivity to external APIs
- [ ] Search existing GitHub issues

### Professional Support

For enterprise support or custom plugin development:
- Contact the maintainers through GitHub
- Consider sponsoring the project
- Explore commercial support options

## Related Documentation

- [Installation Guide](installation.md) - Setup and installation help
- [User Guide](user-guide.md) - Usage instructions and examples
- [Plugin System](plugin-system.md) - Plugin development and management
- [Cost Calculations](cost-calculations.md) - Understanding cost methodologies