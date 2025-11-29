Process all GitHub Copilot code review findings for a pull request. Analyze each finding, implement fixes or provide rationale for denial, reply to each comment, and mark as resolved.

## Instructions

1. Ask me for the PR number if not provided as argument (e.g., `/copilot-review 123`)

2. Fetch all GitHub Copilot code review comments:
   ```bash
   gh api repos/{owner}/{repo}/pulls/{pr_number}/comments
   ```

3. For EACH finding, perform this workflow:

   a. **Analyze the finding**:
      - Read the code context
      - Understand the suggestion
      - Evaluate if it aligns with repo patterns (CLAUDE.md)
      - Check if it improves code quality, security, or maintainability

   b. **Make a decision**:

      **Option 1: ACCEPT**
      - Implement the suggested fix
      - Ensure fix follows repo patterns
      - Test that code still works (if applicable)
      - Reply to comment: "Implemented. [Brief description of what was changed]"
      - Mark comment as resolved

      **Option 2: DENY**
      - Provide clear rationale (e.g., "This conflicts with our sequential test pattern", "This would break idempotency", etc.)
      - Reply to comment: "Not implementing because: [detailed rationale]"
      - Mark comment as resolved

4. After processing all findings:
   - If any implementations were made, commit changes:
     ```
     git add .
     git commit -m "Address GitHub Copilot code review findings for PR #XXX

     - [List major changes]

     Generated with Claude Code
     Co-Authored-By: Claude <noreply@anthropic.com>"
     ```
   - Provide summary of actions taken

## Important Guidelines

- **Security findings**: ALWAYS implement security fixes unless there's a very strong reason not to
- **Pattern compliance**: Prioritize fixes that align with CLAUDE.md patterns
- **Test impact**: Consider if changes affect test behavior or idempotency
- **Be thorough**: Every finding gets a response and resolution
- **Be respectful**: Provide clear rationale for denials

## Response Format for Each Finding

**Finding #N: [Brief description]**
- **Location**: `file.go:line`
- **Copilot Suggestion**: [Summary of what Copilot suggested]
- **Decision**: ✅ ACCEPTED / ❌ DENIED
- **Action**: [What was implemented OR why it was denied]
- **Reply Posted**: [The actual comment posted to PR]
- **Status**: ✅ Resolved

## Using GitHub CLI

To reply to comments:
```bash
gh api -X POST repos/{owner}/{repo}/pulls/comments/{comment_id}/replies \
  -f body="Your reply here"
```

To resolve conversations (if supported by API or UI action needed)

## Summary Report

At the end, provide:

**GitHub Copilot Review Summary for PR #XXX**

- **Total Findings**: X
- **Accepted**: Y
- **Denied**: Z
- **Commits Made**: [Yes/No]

**Key Changes**:
- [List significant implementations]

**Denied Items**:
- [List denials with brief rationale]
