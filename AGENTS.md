# Agent Instructions

## Language
- All plans, documentation, comments, commit messages, and reviews MUST be written in English.
- Do not mix languages.

## Planning workflow
- All non-trivial work MUST start with a Markdown plan in .codexflow/plans.
- Implementation MUST strictly follow the approved plan.
- If scope changes, STOP and update the plan before writing code.

## Plan structure
Each plan MUST include:
- Goal
- Non-goals
- Constraints / assumptions
- Proposed approach
- Files to change
- Step-by-step checklist
- Tests to run
- Rollback plan
- Open questions (if any)

## Implementation rules
- Only modify files listed in the plan.
- Follow the checklist in order.
- Mark checklist items as completed in the plan ([x]) as work progresses.
- Avoid speculative or unrelated refactors.

## Tests
- Every new or changed behavior MUST be covered by tests.
- All tests listed in the plan must be executed before the task is done.

## Reviews
- Reviews MUST focus on correctness, edge cases, adherence to plan, test coverage, and maintainability.
