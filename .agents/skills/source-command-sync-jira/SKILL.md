---
name: "source-command-sync-jira"
description: "Synchronize GitHub issues with Jira sub-tasks. GitHub is source of truth, Jira mirrors status."
---

# source-command-sync-jira

Use this skill when the user asks to run the migrated source command `sync-jira`.

## Command Template

# Sync GitHub Issues to Jira

Synchronize GitHub issues (with `v1` label) to Jira sub-tasks under the CAPZ Test Automation epic.

## Prerequisites

- `credentials.json` must exist in repo root with valid Jira token
- GitHub CLI (`gh`) must be authenticated

## Jira Structure

```
ARO-24162 (Epic): Create E2E tests for CAPZ
├── ARO-24225 (Task): Pre-flight & Environment Checks
├── ARO-24211 (Task): Management Cluster Readiness
├── ARO-24190 (Task): Workload Cluster Deployment & Verification
└── ARO-24163 (Task): Documentation & Tooling
    └── Sub-tasks mirror GitHub issues
```

## Workflow

### Phase 1: Check for Unlabeled Issues

1. **Fetch all open GitHub issues without labels**
   ```bash
   gh issue list --state open --json number,title,labels | jq '[.[] | select(.labels | length == 0)]'
   ```

2. **If unlabeled issues exist**:
   - Display list of unlabeled issues with numbers and titles
   - STOP and inform user: "These issues need labels before syncing. Please add the `v1` label to issues you want tracked, or other labels for issues not in scope."
   - Do NOT proceed until user confirms they've fixed the labels

### Phase 2: Fetch and Compare Issues

1. **Get all GitHub issues with `v1` label**
   ```bash
   gh issue list --label v1 --state all --json number,title,state --limit 500
   ```

2. **Get existing Jira sub-tasks under category tasks**
   - Query Jira API for sub-tasks under: ARO-24225, ARO-24211, ARO-24190, ARO-24163
   - Extract: key, summary, status, description (for GitHub issue link)
   - Match Jira issues to GitHub issues by:
     - GitHub issue URL in Jira description
     - Or title similarity if no URL found

3. **Categorize findings**:
   - **New**: GitHub issues not in Jira (need to create sub-tasks)
   - **Status mismatch**: GitHub closed but Jira open (need to close Jira)
   - **Title mismatch**: GitHub title differs from Jira summary
   - **In sync**: No action needed

### Phase 3: Review GitHub Issue Titles

Before syncing to Jira, review GitHub issue titles for quality:

1. **Check each title for**:
   - Proper capitalization (sentence case or title case, be consistent)
   - Grammar and spelling
   - Clarity and conciseness
   - No trailing punctuation (unless question)
   - No redundant prefixes like "Bug:" or "Feature:" (use labels instead)

2. **If issues found, show suggestions**:
   ```
   Title Review Suggestions:

   #123: "fix bug in authentication"
      → "Fix bug in authentication" (capitalize first letter)

   #456: "Add new feature for the user authentication flow."
      → "Add new feature for user authentication flow" (remove article, trailing period)

   #789: "Bug: test fails"
      → "Test fails intermittently" (remove prefix, be specific)
   ```

3. **Ask user to approve title changes**
   - Use AskUserQuestion: "Apply these title changes to GitHub issues?"
   - Options: "Yes, apply all" / "Let me review individually" / "Skip title changes"

4. **If approved, update GitHub issue titles**
   ```bash
   gh issue edit <number> --title "<new title>"
   ```

### Phase 4: Show Sync Preview

Display a summary of all changes before applying:

```
=== GitHub → Jira Sync Preview ===

NEW ISSUES (will create Jira sub-tasks):
  #370: Add credentials.json to .gitignore
        → Category: ARO-24163 (Documentation & Tooling)

  #371: Fix cluster timeout issue
        → Category: ARO-24190 (Workload Cluster Deployment)

STATUS UPDATES (will close in Jira):
  #365: Update deployment script
        → ACM-28XXX: Currently "In Progress", will transition to "Closed"

TITLE UPDATES (will update Jira summary):
  #360: "Fix typo" → "Fix typo in configuration parser"
        → ACM-28YYY: Will update summary

ALREADY IN SYNC: 45 issues

Proceed with sync?
```

Ask user to confirm before proceeding.

### Phase 5: Apply Changes

1. **Read common metadata from existing sub-task**
   ```bash
   # Get fields from any existing sub-task (e.g., ARO-25350)
   curl -u "$EMAIL:$TOKEN" \
     "https://redhat.atlassian.net/rest/api/2/issue/ARO-25350?fields=labels,components,security,assignee"
   ```

2. **Create new Jira sub-tasks**
   - For each new GitHub issue:
     - Fetch the full issue body:
       ```bash
       BODY=$(gh issue view <number> --repo stolostron/capi-tests --json body --jq '.body')
       ```
     - Build the Jira description by combining the body with the GitHub URL:
       ```
       <full GitHub issue body>

       GitHub Issue: https://github.com/stolostron/capi-tests/issues/<number>
       ```
     - Determine parent task based on category mapping
     - Create sub-task with:
       - `summary`: GitHub issue title
       - `description`: combined body + URL (NEVER leave empty — this is mandatory)
       - `labels`: ["CAPZ", "QE"]
       - `components`: [{"id": "37592"}]
       - `security`: {"id": "10034"}
       - `assignee`: {"accountId": "5fabb5fdecdae600685b01d6"}
     - If GitHub issue is closed, immediately transition to Closed

3. **Close Jira issues for closed GitHub issues**
   - First, check if GitHub issue has linked pull requests:
     ```bash
     gh issue view <number> --json closedByPullRequestsReferences --jq '.closedByPullRequestsReferences[].url'
     ```
   - If PR exists, add it as a comment to Jira before closing:
     ```bash
     curl -X POST -u "$EMAIL:$TOKEN" -H "Content-Type: application/json" \
       "https://redhat.atlassian.net/rest/api/2/issue/$KEY/comment" \
       -d '{"body": "Closed by PR: https://github.com/stolostron/capi-tests/pull/XXX"}'
     ```
   - Then transition to Closed:
     ```bash
     curl -X POST -u "$EMAIL:$TOKEN" -H "Content-Type: application/json" \
       "https://redhat.atlassian.net/rest/api/2/issue/$KEY/transitions" \
       -d '{"transition": {"id": "71"}, "fields": {"resolution": {"name": "Done"}}}'
     ```

4. **Update Jira summaries for title changes**
   ```bash
   curl -X PUT -u "$EMAIL:$TOKEN" -H "Content-Type: application/json" \
     "https://redhat.atlassian.net/rest/api/2/issue/$KEY" \
     -d '{"fields": {"summary": "<new title>"}}'
   ```

### Phase 6: Summary Report

Display final summary:
```
=== Sync Complete ===

Created: 2 new Jira sub-tasks
  - ACM-28633: Add credentials.json to .gitignore
  - ACM-28634: Fix cluster timeout issue

Closed: 1 Jira issue
  - ACM-28XXX: Update deployment script

Updated: 1 title
  - ACM-28YYY: Fix typo in configuration parser

Total synced: 48 issues
```

## Category Mapping Logic

When suggesting categories for new issues, analyze the title and description:

| Keywords/Topics | Jira Parent |
|----------------|-------------|
| dependency, prerequisite, auth, azure cli, tool check, environment, validation | ARO-24225 |
| kind, management cluster, controller, webhook, helm, chart, deploy-charts | ARO-24211 |
| yaml, generate, CR, deployment, verification, cluster, workload, control plane, delete | ARO-24190 |
| doc, readme, Codex, command, tooling, license, contributing, template | ARO-24163 |

If unclear, ask user to choose category.

## Creating New Categories

If an issue doesn't fit existing categories:

1. **Suggest creating new Task under epic**
   - Ask user: "Issue #XXX doesn't fit existing categories. Create new category?"
   - If yes, ask for category name
   - Create Task under ARO-24162 with same metadata as other category tasks
   - Create sub-task under new category

## Error Handling

- **Invalid credentials**: Check `credentials.json` exists and token is valid
- **Rate limiting**: Wait and retry with exponential backoff
- **Network errors**: Display error, offer to retry
- **Jira API errors**: Show error response, suggest manual fix

## Credentials

Read from `credentials.json` in repo root:
```javascript
const creds = JSON.parse(fs.readFileSync('credentials.json'));
const email = creds.jira.email;
const token = creds.jira.token;
// Use: Basic auth -u "$email:$token"
```

## Important Notes

- **GitHub is source of truth** - never modify GitHub based on Jira state
- **Only `v1` label** - ignore issues with other labels or no labels
- **Review before apply** - always show preview, get user approval
- **Title quality** - fix titles on GitHub first, then sync to Jira
- **Metadata consistency** - all sub-tasks must have same labels, components, security, assignee
- **DESCRIPTION IS MANDATORY** - never create a Jira issue with an empty description. Always fetch the full GitHub issue body first. Management reads Jira with AI tools; a missing description makes the issue useless.
