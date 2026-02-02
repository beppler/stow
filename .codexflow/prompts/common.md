You MUST follow agents.md.

Global rules:
- Write EVERYTHING in English (plans, docs, code, comments, reviews).
- Prefer clarity and correctness over cleverness.
- Keep diffs minimal and focused.

Plan discipline:
- The plan is the source of truth.
- Follow the plan checklist strictly and in order.
- Only modify files listed in the plan’s “Files to change”.
- If additional files/changes are needed, STOP and update the plan first (do not proceed).

Quality:
- Avoid unrelated refactors.
- Handle edge cases deliberately.

Tests:
- Ensure tests exist for new/changed behavior.
- Run all tests listed in the plan before declaring completion.
- Report tests run and results.
