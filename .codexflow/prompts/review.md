Task: Produce a CODE REVIEW REPORT as a Markdown document.

Hard rules:
- You must review the code changes against the provided plan.
- Do NOT modify any files. Only write the review report.

What to review:
1) Plan adherence:
   - Are changes limited to files listed in the plan’s “Files to change”?
   - Are checklist items satisfied? Any missing steps?
2) Correctness:
   - Bugs, edge cases, error handling, security concerns
3) Maintainability:
   - Readability, structure, naming, complexity
4) Tests:
   - Are tests adequate?
   - Are the plan’s test commands present and executed?

Output requirements:
- Output MUST be valid Markdown.
- Use this exact structure:

# Review Report
## Plan
- Plan: <path>
- Summary of intended change (1-3 bullets)

## Diff Summary
- Files changed (bullet list)
- High-level change summary (3-8 bullets)

## Findings (prioritized)
### High
- [ ] item
### Medium
- [ ] item
### Low
- [ ] item

## Plan Compliance Checklist
- [ ] Only plan-listed files modified
- [ ] Checklist items completed
- [ ] Tests added/updated as needed
- [ ] Tests executed (state what was run; if unknown, say "not provided")

## Suggested Next Steps
- 3-8 bullets
