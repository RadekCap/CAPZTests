---
description: Sync local main branch with remote and optionally create a new feature branch
---

# Sync Main Branch

Synchronize your local main branch with the remote repository and optionally create a new feature branch for development.

## Workflow

1. **Fetch latest changes from remote**
   ```bash
   git fetch origin main
   ```

2. **Check current branch and status**
   ```bash
   git status
   git branch --show-current
   ```

3. **Check if local main is behind remote**
   ```bash
   git rev-list --count main..origin/main
   ```

4. **If currently on a feature branch, check for uncommitted changes**
   - If there are uncommitted changes, ask the user:
     - "You have uncommitted changes. What would you like to do?"
       - Option 1: Stash changes and continue
       - Option 2: Commit changes first
       - Option 3: Discard changes (dangerous, confirm first)
       - Option 4: Cancel sync operation

5. **Switch to main branch**
   ```bash
   git checkout main
   ```

6. **Pull latest changes**
   ```bash
   git pull origin main
   ```

7. **Show summary of updates**
   - Display number of commits pulled
   - Show brief commit log of new changes
   ```bash
   git log --oneline HEAD@{1}..HEAD
   ```

8. **Ask if user wants to create a new feature branch**
   - Use AskUserQuestion tool with options:
     - "Would you like to create a new feature branch?"
       - Option 1: Yes, create new branch (prompt for branch name)
       - Option 2: No, stay on main

9. **If creating new branch**
   - Ask for branch name using AskUserQuestion
   - Validate branch name (no spaces, lowercase, use hyphens)
   - Create and checkout the new branch:
     ```bash
     git checkout -b <branch-name>
     ```

10. **Provide summary**
    - Confirm current branch
    - Show git status
    - Remind user they're ready to start working

## Important Notes

- **Safety First**: Always check for uncommitted changes before switching branches
- **Clean Merges**: Only fast-forward pulls on main (no merge commits)
- **Branch Naming**: Follow convention: `feature-name` or `fix-issue-description`
- **Remote Tracking**: Ensure new branches track origin when first pushed

## Error Handling

If any step fails:
- Explain what went wrong
- Provide recovery commands
- Ask user how to proceed

## Examples

**Example 1: Simple sync, no new branch**
```
User: /sync-main
Claude: Fetching latest changes...
        Local main is behind by 3 commits. Updating...
        Successfully pulled 3 new commits:
        - abc1234 Fix infrastructure test
        - def5678 Add new slash command
        - ghi9012 Update documentation

        You're now on main branch, up to date with origin.
```

**Example 2: Sync with uncommitted changes**
```
User: /sync-main
Claude: You're currently on branch 'feature-xyz' with uncommitted changes.
        What would you like to do?
        [User selects: Stash changes and continue]
        Stashing changes...
        Switching to main...
        Pulling latest changes...
        Successfully updated main.
        Would you like to create a new feature branch?
        [User selects: No, stay on main]
```

**Example 3: Sync and create new branch**
```
User: /sync-main
Claude: Fetching latest changes...
        Local main is up to date with origin.
        Would you like to create a new feature branch?
        [User selects: Yes, create new branch]
        What should the new branch be called?
        [User enters: add-logging-feature]
        Creating branch 'add-logging-feature'...
        Switched to a new branch 'add-logging-feature'
        Ready to start working!
```

## Post-Sync Checklist

After sync completes, remind the user:
- Current branch name
- Git status (clean/uncommitted changes)
- Number of commits ahead/behind origin (if applicable)
- Suggestion: Run tests if significant changes were pulled
