---
applyTo: "**/COMMIT_EDITMSG,**/.git/COMMIT_EDITMSG,**/.git/MERGE_MSG,**/.git/SQUASH_MSG"
---

# Commit message assistant (Conventional Commits + Chris Beams)

When editing a commit message buffer:
- **Write the commit directly in the `COMMIT_EDITMSG` file**, not in chat.
- **Insert only the final commit text at the very top of the file.**
- Do **not** modify or delete any lines starting with `#`.
- Do **not** modify the cut line `------------------------ >8 ------------------------` or anything after it.
- If non-comment text already exists at the top, **replace only that initial non-comment block** (up to the first blank line or `#`). Leave the rest unchanged.

## Format
- **Subject (line 1):** `<type>(<scope>): <imperative summary>`
  Types: `feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert`
  ≤50 chars, imperative, no trailing period.
- **Blank line**
- **Body (wrap at 72):** what & why (not how), user impact, trade-offs; `Closes #123`. Use a bullet point list.
- **Footer (optional):** `BREAKING CHANGE: ...`, co-authors, etc.

**Mapping hints:** docs→`docs`; build/config→`build`; CI→`ci`; formatting-only→`style`; behavior-preserving→`refactor`; perf-only→`perf`.

**Rules:** Subject ≤50; body wrap 72; avoid passive voice; if multiple unrelated changes, focus on the dominant change and summarize the rest in the body.

## Correct placement example (how the buffer should look after your insertion)

```.git/COMMIT_EDITMSG
feat(instructions): add path-specific commit assistant for COMMIT_EDITMSG

Explain placement-only edits, enforce 50/72, and keep comments/cut line intact.
Closes #123

# Please enter the commit message for your changes. Lines starting
# with '#' will be ignored, and an empty message aborts the commit.
#
# Date:      <...>
#
# On branch <...>
# Changes to be committed:
#	<...>
#
# ------------------------ >8 ------------------------
# Do not modify or remove the line above.
# Everything below it will be ignored.
<diffs that the agent MUST read>
```
