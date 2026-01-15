# Quickstart: Migrating to FinFocus

## Installation

```bash
# Install the new binary
go install github.com/rshade/finfocus/cmd/finfocus@latest

# Verify installation
finfocus --version
```

## First Run & Migration

If you are an existing `finfocus` user, `finfocus` will detect your old configuration automatically.

1. Run `finfocus cost`:
   ```text
   > finfocus cost
   
   Detected legacy configuration at ~/.finfocus.
   Would you like to migrate to ~/.finfocus? [y/N] y
   
   Migration complete. Your old config has been preserved at ~/.finfocus.
   ```

2. **Action Required**: Rename your plugins!
   Automatic migration copies the files, but does NOT rename binary executables to avoid breaking signature/integrity if they were checked.
   
   ```bash
   cd ~/.finfocus/plugins
   mv finfocus-plugin-aws-public finfocus-plugin-aws-public
   # Repeat for other plugins
   ```
   
   *Alternatively*, run with compatibility mode to use old names:
   ```bash
   export FINFOCUS_LOG_LEGACY=1
   finfocus cost
   ```

## Shell Setup

Add the recommended alias to your shell (`.bashrc` / `.zshrc`):

```bash
alias fin=finfocus
```

## Using Legacy CI/CD

If your CI pipeline relies on `FINFOCUS_` environment variables, enable compatibility mode:

```bash
export FINFOCUS_COMPAT=1
export FINFOCUS_LOG_LEVEL=debug # Now recognized
finfocus cost
```
