# Research Template: External Repo Analysis

Use this template when analyzing external repositories for patterns and decisions.

## How to Use

1. Copy the WIP memory template for the specific repo
2. Fill in findings as you examine each area
3. Update the summary section with key takeaways
4. Archive to a final memory when complete

## Research Task Checklist

### Before Starting
- [ ] Clone/download the repo to `bosun/` sibling folder
- [ ] Read README for project overview
- [ ] Identify relevant directories (skip frontend, unrelated features)
- [ ] Create WIP memory from template

### Primary Research (answer the specific question)
- [ ] Find the code that directly addresses your research question
- [ ] Document the approach with code snippets
- [ ] Note pros/cons observed in practice
- [ ] Check issues/PRs for known problems with the approach

### Secondary Research (useful patterns)
- [ ] Docker client abstraction
- [ ] Error handling patterns
- [ ] Configuration patterns (labels, env, files)
- [ ] Testing patterns
- [ ] Project structure / architecture

### After Research
- [ ] Update summary section in WIP memory
- [ ] Create final decision memory if applicable
- [ ] Clean up / archive WIP memory
- [ ] Delete cloned repo if no longer needed

## Expected Output Format

### For Decision Questions (#109, #110, #117)

```markdown
## Decision: [Option A/B/C]

### Evidence
- **Watchtower**: Uses X approach because Y
- **Portainer**: Uses Z approach because W
- **Consensus**: Both use X / They differ because...

### Rationale
1. Reason 1
2. Reason 2
3. Reason 3

### Implications for Bosun
- Port interface should: ...
- Adapter should: ...
- New issues needed: ...
```

### For Pattern Discovery

```markdown
## Pattern: [Pattern Name]

### Source
- Repo: [watchtower/portainer]
- Files: [list of files]

### Description
[What the pattern does and why]

### Code Example
```go
// Simplified example from source
```

### Applicability to Bosun
- **Where**: [Which package/file in bosun]
- **Priority**: [High/Medium/Low]
- **Effort**: [Small/Medium/Large]
- **Issue**: [Create issue # if needed]
```

## Research Question Templates

### #109 - Compose Control
Focus areas:
1. How does it interact with Docker? (client creation, API calls)
2. How does it manage multi-container apps? (ordering, dependencies)
3. Does it shell out to CLI or use API directly?
4. How does it handle health checks?

### #110 - Worker Architecture
Focus areas:
1. How are containers started with specific commands?
2. Environment variable injection patterns
3. Signal handling (SIGTERM/SIGKILL)
4. Container cleanup on success/failure

### #117 - Failure Handling
Focus areas:
1. Timeout configuration and defaults
2. Retry logic
3. Rollback/recovery patterns
4. Error types and propagation

---

*This template is for reference. Create repo-specific WIP memories for actual research.*
