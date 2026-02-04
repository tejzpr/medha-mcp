# Medha MCP Instructions

Git-backed AI memory system. **Recall before answering, store after solving.**

## Tools

| Tool | Intent |
|------|--------|
| `medha_recall` | Find stored information (use `topic`, `exact`, or `list_all`) |
| `medha_remember` | Create/update memories (requires `title`, `content`) |
| `medha_history` | View changes over time |
| `medha_connect` | Link related memories |
| `medha_forget` | Archive outdated info |
| `medha_restore` | Undelete archived memories |
| `medha_sync` | Manual git sync |

## Key Behaviors

1. **Check first** - Use `medha_recall` before answering questions
2. **Store valuable info** - Decisions, solutions, context, action items
3. **Supersede, don't duplicate** - Use `replaces` param when updating
4. **Connect while storing** - Use `connections` param to link memories

## Remember Params

- `title`, `content` (required)
- `tags`, `slug`, `path` (optional)
- `replaces`: slug of memory being superseded
- `connections`: `[{"to": "slug", "relationship": "type"}]`

## Relationship Types

`related` (default), `references`, `follows`, `supersedes`, `part_of`, `person`, `project`

## Don't Store

Credentials/secrets, or anything user doesn't want stored.
