# Remediation Plan: Plugin Integration Tests

**Status**: Optional (User requested suggestions)
**Base Artifacts**: `specs/001-plugin-integration-tests/tasks.md`

## Summary

This plan suggests minor additions to `tasks.md` to explicitly cover the "Permission Issues" edge case and ensure command aliases are fully tested, aligning strictly with the spec.

## Suggested Changes

### 1. Add Permission Failure Test (Category: Coverage)

**Rationale**: The spec lists "Permission Issues" as an edge case, but no specific task explicitly covers it. Adding a test ensures system robustness against filesystem errors.

**Action**: Add task T034 to Phase 4 (Install) or Phase 7 (Polish).

```markdown
- [ ] T034 [P] [US2] Implement `TestPluginInstall_PermissionError` verifying handling of unwritable directories (using chmod 000 on temp dir)
```

### 2. Clarify Alias Testing (Category: Underspecification)

**Rationale**: Task T029 mentions "Aliases", but the spec specifically calls out `uninstall` and `rm`. Explicitly listing them ensures both are verified.

**Action**: Update T029 description.

**Current**:
```markdown
- [ ] T029 [P] [US4] Implement `TestPluginRemove_Aliases` verifying `uninstall` and `rm` work
```

**Proposed** (No change needed, description is already specific enough. I will leave it as is).

### 3. Locking Robustness (Category: Constitution/Robustness)

**Rationale**: Task T006 implies potential implementation. To ensure completeness, we should add a verification step or note about cross-platform support.

**Action**: Update T006 description to be explicit about cross-platform.

**Current**:
```markdown
- [ ] T006 [P] Verify if `internal/registry` implements file locking; if not, add `flock` or `sync.Mutex` support to `internal/registry/installer.go`
```

**Proposed**:
```markdown
- [ ] T006 [P] Verify if `internal/registry` implements file locking; if not, add cross-platform file locking (e.g., using `gofrs/flock` or similar compatible approach) to `internal/registry/installer.go`
```

## Applied Changes (if approved)

If you approve, I would:
1.  Append T034 to Phase 4 in `tasks.md`.
2.  Update T006 in `tasks.md` to emphasize cross-platform compatibility.

*Since these are minor and I am in read-only analysis mode for the `speckit.analyze` command flow, I will not apply them automatically.*

**Recommendation**: You can proceed to `/speckit.implement` directly as these are not blocking. You can manually add the permission test if you feel it's necessary during implementation.
