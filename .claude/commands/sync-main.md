---
description: Sync local main branch with remote and optionally create a new feature branch
---

# Sync Main Branch

Synchronize your local main branch with the remote repository. If on a feature branch, offers to rebase onto the updated main (this repository uses rebase, not merge).

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
   - Store the current branch name for later use

3. **Check for uncommitted changes**
   - Check `git status` output for any uncommitted changes (staged or unstaged)
   - If there are uncommitted changes, ask the user:
     - "You have uncommitted changes. What would you like to do?"
       - Option 1: Stash changes and continue
       - Option 2: Commit changes first
       - Option 3: Discard changes (dangerous, confirm first)
       - Option 4: Cancel sync operation
   - Handle user's choice before proceeding

4. **Check if local main is behind remote**
   ```bash
   git rev-list --count main..origin/main
   ```
   - If count is 0, main is up to date (inform user)
   - If count > 0, main is behind (proceed with sync)

5. **Determine workflow based on current branch**

   **If on main branch:**
   - Proceed to step 6 (update main directly)

   **If on a feature branch:**
   - Ask user what they want to do:
     - Option 1: Rebase current branch onto updated main (Recommended)
     - Option 2: Switch to main and update it
     - Option 3: Cancel
   - If user chooses rebase, skip to step 8 (rebase workflow)

6. **Update main branch (if on main or user chose to switch)**
   - If not on main, switch to it:
     ```bash
     git checkout main
     ```
   - Capture current commit for comparison:
     ```bash
     BEFORE_PULL=$(git rev-parse HEAD)
     ```
   - Pull latest changes with fast-forward only:
     ```bash
     git pull --ff-only origin main
     ```
     - If this fails (diverged history), explain error and suggest resolution
     - Only fast-forward merges allowed (no merge commits)

7. **Show summary of updates**
   - Display number of commits pulled
   - Show brief commit log of new changes
   ```bash
   git log --oneline $BEFORE_PULL..HEAD
   ```
   - If no new commits, confirm main was already up to date

8. **Rebase workflow (if user chose to rebase feature branch)**
   - Rebase current branch onto origin/main:
     ```bash
     git rebase origin/main
     ```
   - If conflicts occur:
     - Inform user about the conflict
     - Explain how to resolve: edit files, `git add`, `git rebase --continue`
     - Offer to abort: `git rebase --abort`
   - If rebase succeeds, show summary of rebased commits

9. **Ask about next steps**

   **If updated main (not rebasing):**
   - Ask if user wants to create a new feature branch:
     - Option 1: Yes, create new branch (prompt for branch name)
     - Option 2: No, stay on main

   **If rebased feature branch:**
   - Inform user that force push is needed if branch was already pushed:
     ```bash
     git push --force-with-lease
     ```
   - Ask if they want to force push now:
     - Option 1: Yes, force push with lease
     - Option 2: No, I'll push later

10. **If creating new branch**
    - Ask for branch name using AskUserQuestion
    - Validate branch name follows Git naming conventions
    - Create and checkout the new branch:
      ```bash
      git checkout -b <branch-name>
      ```
    - Remind user to use `git push -u origin <branch-name>` on first push to set up tracking

11. **Provide summary**
    - Confirm current branch
    - Show git status
    - If rebased: show commits ahead of origin
    - Remind user they're ready to start working

## Important Notes

- **Rebase, Not Merge**: This repository uses rebase to maintain clean, linear history. Never use `git merge main` to update feature branches.
- **Safety First**: Always check for uncommitted changes before switching branches or rebasing
- **Force Push with Lease**: After rebasing, use `--force-with-lease` (not `--force`) to push. This prevents overwriting others' changes.
- **Clean Merges**: Only fast-forward pulls on main using `--ff-only` flag (no merge commits)
- **Branch Naming**: Validate branch name follows Git naming conventions (e.g., no spaces or control characters, cannot start with `-`, `.`, or `/`, cannot contain `..`, `@{`, `\`, or end with `.lock`, cannot end with `/` or `.`)
- **Remote Tracking**: New branches are created locally. Set up tracking on first push with: `git push -u origin <branch-name>`

## Error Handling

If any step fails:
- Explain what went wrong
- Provide recovery commands
- Ask user how to proceed

Common errors:
- **Pull fails with --ff-only**: Local main has diverged. User needs to either reset to origin/main or merge manually
- **Rebase conflicts**: Guide user through conflict resolution or offer to abort
- **Invalid branch name**: Git will reject. Prompt for new name
- **Uncommitted changes**: Handle via step 3 options

## Examples

**Example 1: Sync main, no feature branch**
```
User: /sync-main
Claude: Fetching latest changes...
        You're on the main branch.
        No uncommitted changes detected.
        Local main is behind by 3 commits. Updating...
        Successfully pulled 3 new commits:
        - abc1234 Fix infrastructure test
        - def5678 Add new slash command
        - ghi9012 Update documentation

        Would you like to create a new feature branch?
        [User selects: No, stay on main]
        You're now on main branch, up to date with origin.
```

**Example 2: Rebase feature branch onto main**
```
User: /sync-main
Claude: Fetching latest changes...
        You're on branch 'feature-add-logging'.
        No uncommitted changes detected.
        Main is behind origin by 2 commits.

        What would you like to do?
        [User selects: Rebase current branch onto updated main]

        Rebasing feature-add-logging onto origin/main...
        Successfully rebased onto main (2 new commits from main).
        Your 3 commits have been replayed on top.

        Your branch has diverged from origin. To update the remote:
        [User selects: Yes, force push with lease]

        Pushed successfully! Branch is now up to date.
```

**Example 3: Rebase with conflicts**
```
User: /sync-main
Claude: Fetching latest changes...
        You're on branch 'fix-bug-123'.
        No uncommitted changes detected.
        Main is behind origin by 1 commit.

        What would you like to do?
        [User selects: Rebase current branch onto updated main]

        Rebasing fix-bug-123 onto origin/main...

        ⚠️ Conflict detected in: test/helpers.go

        To resolve:
        1. Edit test/helpers.go to resolve conflicts
        2. Stage the file: git add test/helpers.go
        3. Continue rebase: git rebase --continue

        Or to abort: git rebase --abort

        Would you like me to help resolve the conflict?
```

**Example 4: Sync and create new branch**
```
User: /sync-main
Claude: Fetching latest changes...
        You're on the main branch.
        No uncommitted changes detected.
        Local main is up to date with origin.

        Would you like to create a new feature branch?
        [User selects: Yes, create new branch]
        What should the new branch be called?
        [User enters: add-logging-feature]
        Creating branch 'add-logging-feature'...
        Switched to a new branch 'add-logging-feature'

        Tip: When ready to push, use: git push -u origin add-logging-feature

        Ready to start working!
```

## Post-Sync Checklist

After sync completes, remind the user:
- Current branch name
- Git status (clean/uncommitted changes)
- Number of commits ahead/behind origin (if applicable)
- If rebased: remind about force push if needed
- Suggestion: Run tests if significant changes were pulled
- If stashed changes: Remind how to restore them (`git stash list`, `git stash pop`)
