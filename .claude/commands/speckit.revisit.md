---
description: Escalate implementation issues back to research/planning phases when constitution violations, conflicts, or design gaps are discovered.
---

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Purpose

During implementation (`/speckit.implement`), issues may arise that cannot be resolved at the task level:

- **Constitution violations** discovered in generated code or tests
- **Conflicting instructions** from different sources (e.g., mode prompts vs constitution)
- **Design gaps** where the plan doesn't address a discovered requirement
- **Technical blockers** where the chosen approach doesn't work
- **Scope creep** where implementation reveals missing requirements

This command creates a formal escalation path to revisit earlier phases without losing implementation progress.

## Outline

1. **Setup**: Run `.specify/scripts/bash/check-prerequisites.sh --json --require-tasks --include-tasks` from repo root and parse FEATURE_DIR. For single quotes in args like "I'm Groot", use escape syntax: e.g 'I'\''m Groot' (or double-quote if possible: "I'm Groot").

2. **Capture the Issue**: Document the discovered problem:
   - **Issue Type**: Select from [constitution-violation | instruction-conflict | design-gap | technical-blocker | scope-creep]
   - **Description**: Clear statement of what was discovered
   - **Discovery Context**: Where/when the issue was found (task ID, file, line)
   - **Impact**: What cannot proceed until this is resolved

3. **Classify Escalation Level**: Determine which phase needs revisiting:

   | Issue Type | Escalation Target | Action |
   |------------|-------------------|--------|
   | constitution-violation | Plan + Tasks | Update plan to comply, regenerate affected tasks |
   | instruction-conflict | Research + Plan | Research the conflict, update plan with resolution |
   | design-gap | Spec + Plan | Clarify spec, update plan with missing design |
   | technical-blocker | Research | Research alternatives, update research.md |
   | scope-creep | Spec | Re-run /speckit.clarify to capture new requirements |

4. **Create Issue Log**: Append to `FEATURE_DIR/issues.md`:

   ```markdown
   ## Issue: [ID] - [Title]

   **Type**: [issue-type]
   **Discovered**: [date] during [task-id]
   **Status**: Open | Researching | Resolved

   ### Description
   [What was discovered]

   ### Impact
   [What cannot proceed]

   ### Resolution
   [To be filled after resolution]
   ```

5. **Pause Implementation**: Mark current task as BLOCKED in tasks.md:
   - Add `[BLOCKED: issue-ID]` marker to the task
   - Update task status to show dependency on issue resolution

6. **Execute Escalation**: Based on escalation level:

   **For Research-Level Issues**:
   - Add research question to research.md under "## Open Questions"
   - Dispatch research agent to investigate
   - Update research.md with findings
   - Recommend: "Run `/speckit.plan` to incorporate findings"

   **For Plan-Level Issues**:
   - Document the gap in plan.md under "## Discovered Constraints"
   - Update relevant sections (architecture, data model, etc.)
   - Recommend: "Run `/speckit.tasks` to regenerate affected tasks"

   **For Spec-Level Issues**:
   - Recommend: "Run `/speckit.clarify` with the new requirement context"
   - Provide specific questions to address

7. **Update Constitution Compliance Log**: If constitution violation, add to `FEATURE_DIR/constitution-checks.md`:

   ```markdown
   ## Violation: [ID]

   **Principle**: [which principle was violated]
   **Discovered**: [date]
   **Root Cause**: [why it happened]
   **Prevention**: [how to prevent recurrence]
   ```

8. **Report and Recommend**: Output a summary:
   - Issue ID and classification
   - Escalation target(s)
   - Updated artifacts
   - Next command to run
   - Tasks that are blocked

9. **Next Command Recommendation**: Based on escalation level, recommend:

   | Issue Type | Next Command | Purpose |
   |------------|--------------|---------|
   | constitution-violation | `/speckit.plan` | Update plan to comply with constitution |
   | instruction-conflict | `/speckit.plan` | Research conflict, update plan |
   | design-gap | `/speckit.clarify` | Capture missing requirements |
   | technical-blocker | `/speckit.plan` | Research alternatives |
   | scope-creep | `/speckit.specify` | Update specification |

   After the recommended command completes:
   - Run `/speckit.tasks` to regenerate task list
   - Run `/speckit.implement` to resume implementation

## Example Usage

```text
/speckit.revisit constitution-violation: TODO comments were added to code despite Principle VI forbidding them. Root cause: learning mode instructions conflicted with constitution.
```

## Constitution Authority

This command exists specifically to enforce constitution compliance:

- Constitution principles are **NON-NEGOTIABLE**
- Runtime mode instructions (learning, explanatory, etc.) **MUST yield** to constitution
- If a conflict exists between mode instructions and constitution, the constitution wins
- This command documents such conflicts for future prevention

## Key Rules

- Use absolute paths for all file operations
- NEVER modify tasks.md task content (only add BLOCKED markers)
- ALWAYS create an issue entry in issues.md
- ALWAYS provide a clear next-step command
- Track issue IDs incrementally (ISS-001, ISS-002, etc.)
