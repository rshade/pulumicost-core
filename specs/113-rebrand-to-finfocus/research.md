# Research: Project Rename to FinFocus

## 1. Binary Rename via GoReleaser

**Decision**: Update `.goreleaser.yaml` to output `finfocus` binary.
**Rationale**: GoReleaser controls the final binary name and archive structure.
**Changes Required**:
- `builds.binary`: Change `finfocus` to `finfocus`
- `builds.main`: Change `./cmd/finfocus` to `./cmd/finfocus` (after directory move)
- `ldflags`: Update package paths from `github.com/rshade/finfocus/...` to `github.com/rshade/finfocus/...` (after module rename)

## 2. Configuration Migration Strategy

**Decision**: Implement "Safe Copy" migration on startup.
**Rationale**: We must strictly avoid data loss. Moving (renaming) the directory is risky if the process crashes mid-operation. Copying preserves the original data as a backup.
**Mechanism**:
1. Check if `~/.finfocus` exists. If yes, proceed normally.
2. If no, check if `~/.finfocus` exists.
3. If yes, prompt user: "Migrate config from ~/.finfocus to ~/.finfocus? [y/N]"
4. If yes, perform recursive copy of `~/.finfocus` to `~/.finfocus`.
5. Log success and strict instruction: "Migration complete. The old directory ~/.finfocus has been preserved."

## 3. Go Module Rename

**Decision**: `go mod edit -module github.com/rshade/finfocus` followed by global import replace.
**Rationale**: Standard Go practice.
**Caveats**:
- Must be done in a single atomic commit to avoid breaking build.
- `go.sum` will need regeneration.
- All internal imports must be updated using `gofmt -w -r ...` or `sed`.

## 4. Environment Variables

**Decision**: Use `viper` aliasing or explicit middleware.
**Rationale**: Viper supports `SetEnvPrefix`. We will change this to `FINFOCUS`.
**Legacy Support**:
- Create a specific loader that checks `FINFOCUS_COMPAT=1`.
- If set, manually check `os.Getenv("FINFOCUS_...")` and inject into Viper configuration if the `FINFOCUS_` equivalent is missing.

## 5. Plugin Discovery

**Decision**: Update `pluginhost` package to scan `~/.finfocus/plugins`.
**Legacy Support**:
- If `FINFOCUS_LOG_LEGACY=1`, *also* scan for files matching `finfocus-plugin-*` in the same directory (or potentially the old directory, but strict spec says `~/.finfocus/plugins`).
- *Correction from Spec*: Spec FR-008 says "look for plugins in ~/.finfocus/plugins". FR-014 says support legacy binary names. So we only look in the NEW directory, but accept OLD binary names if the toggle is set. We do not scan `~/.finfocus/plugins`. Users must copy plugins during migration.

