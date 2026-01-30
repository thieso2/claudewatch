# Claude Session Files Format

## Overview

Claude Code (the CLI) stores conversation sessions in JSONL format (JSON Lines - one JSON object per line) at:
```
~/.claude/projects/<encoded-project-path>/<session-id>.jsonl
```

Each line represents a single event in the conversation timeline, including:
- User messages
- Claude's responses (with full thinking and token usage)
- System events (timestamps, turn duration)
- File history snapshots
- Metadata about the session

---

## Directory Structure

```
~/.claude/projects/
├── -Users-thies-Projects-MyProject/
│   ├── sessions-index.json          # Project metadata and session index
│   ├── abc123-session-id.jsonl      # Session file (one per conversation)
│   ├── def456-session-id.jsonl
│   └── ...
├── -Users-thies-Projects-OtherProj/
│   ├── sessions-index.json
│   └── xyz789-session-id.jsonl
└── ...
```

### Project Name Encoding

Project paths are encoded by replacing `/` with `-` for filesystem compatibility:
- Original: `/Users/thies/Projects/SaaS-Bonn/cloud`
- Encoded: `-Users-thies-Projects-SaaS-Bonn-cloud`

---

## Session Index File: `sessions-index.json`

This file contains metadata about all sessions in a project.

### Structure
```json
{
  "version": 1,
  "entries": [
    {
      "sessionId": "9302f35a-ca98-46bf-946e-28557eff7682",
      "fullPath": "/Users/thies/.claude/projects/...",
      "fileMtime": 1768804583385,
      "firstPrompt": "create a comprehensive architecture diagram...",
      "messageCount": 6,
      "created": "2026-01-19T06:23:09.969Z",
      "modified": "2026-01-19T06:36:23.317Z",
      "gitBranch": "arch",
      "projectPath": "/Users/thies/Projects/SaaS-Bonn/mmc-rails-arch",
      "isSidechain": false
    }
  ],
  "originalPath": "/Users/thies/Projects/SaaS-Bonn/cloud"
}
```

### Fields

| Field | Type | Description |
|-------|------|-------------|
| `version` | number | Format version (currently 1) |
| `entries` | array | List of session entries |
| `originalPath` | string | Original project directory path |

### Entry Fields

| Field | Type | Description |
|-------|------|-------------|
| `sessionId` | UUID | Unique identifier for this conversation session |
| `fullPath` | string | Absolute path to the JSONL session file |
| `fileMtime` | number | File modification timestamp (Unix milliseconds) |
| `firstPrompt` | string | The initial user message that started the session |
| `messageCount` | number | Total number of user/assistant messages (conversations) |
| `created` | ISO8601 | When the session was created |
| `modified` | ISO8601 | When the session was last modified |
| `gitBranch` | string | Git branch active when session was created |
| `projectPath` | string | Original unencoded project path |
| `isSidechain` | boolean | Whether this is a side/branching conversation |

---

## Session JSONL File Format

Each line is a complete JSON object representing one event. Events are ordered chronologically.

### Common Fields (All Events)

```json
{
  "type": "user|assistant|system|file-history-snapshot",
  "timestamp": "2026-01-09T14:02:46.229Z",
  "uuid": "unique-message-id",
  "sessionId": "session-uuid",
  "version": "2.1.1",
  "userType": "external",
  "cwd": "/path/to/working/directory",
  "parentUuid": null,
  "isSidechain": false,
  "gitBranch": "main"
}
```

| Field | Description |
|-------|-------------|
| `type` | Event type: `user`, `assistant`, `system`, `file-history-snapshot` |
| `timestamp` | ISO8601 timestamp (UTC) when event occurred |
| `uuid` | Unique identifier for this message/event |
| `sessionId` | Session this event belongs to |
| `version` | Claude version that generated this event (e.g., "2.1.1") |
| `userType` | User type: `"external"` |
| `cwd` | Current working directory when message was sent/received |
| `parentUuid` | ID of parent message (for threading conversations) |
| `isSidechain` | Whether this is a branch of the main conversation |
| `gitBranch` | Git branch active at time of message |

---

## Event Types

### 1. User Message: `"type": "user"`

User-submitted prompt or message.

```json
{
  "type": "user",
  "message": {
    "role": "user",
    "content": "create a comprehensive README.md and link all files"
  },
  "uuid": "8a2edc47-147a-464c-b5fb-683d9e53ecd9",
  "timestamp": "2026-01-09T14:02:46.229Z",
  "thinkingMetadata": {
    "level": "high",
    "disabled": false,
    "triggers": []
  },
  "todos": []
}
```

#### Fields

| Field | Type | Description |
|-------|------|-------------|
| `message.role` | string | Always `"user"` |
| `message.content` | string | The user's prompt text |
| `thinkingMetadata` | object | Extended thinking configuration |
| `thinkingMetadata.level` | string | `"high"`, `"medium"`, `"low"`, or null |
| `thinkingMetadata.disabled` | boolean | Whether thinking was disabled |
| `thinkingMetadata.triggers` | array | What triggered extended thinking |
| `todos` | array | Any TODOs/tasks extracted from the prompt |

---

### 2. Assistant Response: `"type": "assistant"`

Claude's response to a user message. **This is where token usage data is stored.**

```json
{
  "type": "assistant",
  "message": {
    "model": "claude-sonnet-4-5-20250929",
    "id": "msg_01YbWXGKJhwjeaqzn36cSP3N",
    "type": "message",
    "role": "assistant",
    "content": [
      {
        "type": "thinking",
        "thinking": "The user wants me to...",
        "signature": "base64-encoded-signature"
      },
      {
        "type": "text",
        "text": "Here's the comprehensive README.md..."
      },
      {
        "type": "tool_use",
        "id": "tool_123",
        "name": "write_file",
        "input": {
          "path": "README.md",
          "content": "..."
        }
      }
    ],
    "stop_reason": "end_turn|stop_sequence|tool_use",
    "stop_sequence": null,
    "usage": {
      "input_tokens": 10,
      "cache_creation_input_tokens": 35801,
      "cache_read_input_tokens": 0,
      "cache_creation": {
        "ephemeral_5m_input_tokens": 35801,
        "ephemeral_1h_input_tokens": 0
      },
      "output_tokens": 3,
      "service_tier": "standard"
    }
  },
  "requestId": "req_011CWwvnwANMMkTnp6b2o85G",
  "uuid": "93654cee-6c51-4e6f-afdf-65c650312017",
  "timestamp": "2026-01-09T14:02:52.977Z"
}
```

#### Main Fields

| Field | Type | Description |
|-------|------|-------------|
| `message.model` | string | Claude model used (e.g., `claude-sonnet-4-5-20250929`) |
| `message.id` | string | API message ID from Anthropic |
| `message.role` | string | Always `"assistant"` |
| `message.content` | array | Array of content blocks (see below) |
| `message.stop_reason` | string | Why response stopped: `end_turn`, `stop_sequence`, `tool_use`, `max_tokens` |
| `message.stop_sequence` | string | The sequence that triggered stop (usually null) |
| `requestId` | string | API request ID for debugging |

#### Content Block Types

Content is an array of objects with different types:

##### Thinking Block
```json
{
  "type": "thinking",
  "thinking": "Extended reasoning text...",
  "signature": "base64-encoded-signature"
}
```
- Contains Claude's internal reasoning when extended thinking is enabled
- Hidden from users by default

##### Text Block
```json
{
  "type": "text",
  "text": "Regular response text..."
}
```
- Human-readable text response

##### Tool Use Block
```json
{
  "type": "tool_use",
  "id": "call_123",
  "name": "write_file",
  "input": {
    "path": "filename.txt",
    "content": "..."
  }
}
```
- Claude calling a tool/function
- `name`: tool name
- `input`: tool arguments as JSON object

---

### Token Usage: `message.usage`

**Critical for analyzing session costs and efficiency:**

```json
"usage": {
  "input_tokens": 10,
  "cache_creation_input_tokens": 35801,
  "cache_read_input_tokens": 0,
  "cache_creation": {
    "ephemeral_5m_input_tokens": 35801,
    "ephemeral_1h_input_tokens": 0
  },
  "output_tokens": 3,
  "service_tier": "standard"
}
```

| Field | Description |
|-------|-------------|
| `input_tokens` | Tokens from current turn's input |
| `cache_creation_input_tokens` | Tokens written to cache (cached but charged) |
| `cache_read_input_tokens` | Tokens read from cache (cached and cheaper) |
| `cache_creation.ephemeral_5m_input_tokens` | Ephemeral cache tokens (5 minute TTL) |
| `cache_creation.ephemeral_1h_input_tokens` | Ephemeral cache tokens (1 hour TTL) |
| `output_tokens` | Tokens in Claude's response |
| `service_tier` | Billing tier (usually `"standard"`) |

**Token Cost Calculation:**
- `input_tokens`: Full price
- `cache_creation_input_tokens`: Full price (written to cache)
- `cache_read_input_tokens`: 10% of input price (cached hit)
- `output_tokens`: Full price (output is never cached)

---

### 3. System Event: `"type": "system"`

System-level metadata about the session progress.

```json
{
  "type": "system",
  "subtype": "turn_duration",
  "durationMs": 119773,
  "slug": "tender-mixing-honey",
  "timestamp": "2026-01-09T14:04:46.007Z",
  "uuid": "b802c1c5-6b8a-4c69-99db-42d8bee339d6",
  "isMeta": false
}
```

| Field | Description |
|-------|-------------|
| `subtype` | Type of system event (e.g., `turn_duration`) |
| `durationMs` | Duration of the turn in milliseconds |
| `slug` | Human-readable slug for the turn |
| `isMeta` | Whether this is metadata-only (not shown in UI) |

---

### 4. File History Snapshot: `"type": "file-history-snapshot"`

Snapshot of file states for recovery/reference.

```json
{
  "type": "file-history-snapshot",
  "messageId": "8a2edc47-147a-464c-b5fb-683d9e53ecd9",
  "snapshot": {
    "messageId": "8a2edc47-147a-464c-b5fb-683d9e53ecd9",
    "trackedFileBackups": {},
    "timestamp": "2026-01-09T14:02:46.234Z"
  },
  "isSnapshotUpdate": false
}
```

| Field | Description |
|-------|-------------|
| `messageId` | ID of message this snapshot is attached to |
| `snapshot.trackedFileBackups` | Map of file path → file contents for tracked files |
| `isSnapshotUpdate` | Whether this updates a previous snapshot |

---

## Analysis Examples

### Calculate Total Tokens Used in a Session

```python
import json

total_input = 0
total_cache_creation = 0
total_cache_read = 0
total_output = 0

with open('session.jsonl', 'r') as f:
    for line in f:
        event = json.loads(line)
        if event['type'] == 'assistant':
            usage = event['message'].get('usage', {})
            total_input += usage.get('input_tokens', 0)
            total_cache_creation += usage.get('cache_creation_input_tokens', 0)
            total_cache_read += usage.get('cache_read_input_tokens', 0)
            total_output += usage.get('output_tokens', 0)

# Calculate cost (at Claude 3.5 Sonnet pricing as of Jan 2026)
# Input: $3 per 1M tokens, Cache write: $3 per 1M, Cache read: $0.30 per 1M, Output: $15 per 1M
input_cost = (total_input + total_cache_creation) * 0.000003
cache_read_cost = total_cache_read * 0.0000003
output_cost = total_output * 0.000015
total_cost = input_cost + cache_read_cost + output_cost

print(f"Input tokens: {total_input:,}")
print(f"Cache creation tokens: {total_cache_creation:,}")
print(f"Cache read tokens: {total_cache_read:,}")
print(f"Output tokens: {total_output:,}")
print(f"Estimated cost: ${total_cost:.4f}")
```

### Extract All User Prompts

```python
import json

with open('session.jsonl', 'r') as f:
    for line in f:
        event = json.loads(line)
        if event['type'] == 'user':
            timestamp = event['timestamp']
            message = event['message']['content']
            print(f"[{timestamp}] {message[:100]}...")
```

### Find Claude Model Changes

```python
import json

models_used = {}

with open('session.jsonl', 'r') as f:
    for line in f:
        event = json.loads(line)
        if event['type'] == 'assistant':
            model = event['message'].get('model', 'unknown')
            if model not in models_used:
                models_used[model] = 0
            models_used[model] += 1

for model, count in models_used.items():
    print(f"{model}: {count} responses")
```

### Calculate Session Duration and Gaps

```python
import json
from datetime import datetime

timestamps = []

with open('session.jsonl', 'r') as f:
    for line in f:
        event = json.loads(line)
        ts = datetime.fromisoformat(event['timestamp'].replace('Z', '+00:00'))
        if event['type'] in ['user', 'assistant']:
            timestamps.append(ts)

if timestamps:
    total_duration = timestamps[-1] - timestamps[0]
    print(f"Session duration: {total_duration}")

    # Find gaps > 1 hour (resumptions)
    resumptions = 0
    for i in range(1, len(timestamps)):
        gap = timestamps[i] - timestamps[i-1]
        if gap.total_seconds() > 3600:
            resumptions += 1
            print(f"Gap of {gap} at {timestamps[i]}")

    print(f"Resumptions: {resumptions}")
```

---

## Key Insights

### What You Can Measure

1. **Token Usage & Cost**
   - Total input/output tokens per session
   - Cache hit rates
   - Estimated API costs

2. **Session Quality**
   - User prompt count vs assistant responses
   - Average response length
   - Turn duration (from system events)
   - Model performance over time

3. **Work Patterns**
   - Session duration and gaps (resumptions)
   - Time between messages
   - Git branch context
   - Working directory changes

4. **Claude's Behavior**
   - Tool usage patterns
   - Thinking vs output balance
   - Model versions used
   - Stop reasons (natural end vs timeout)

5. **File Impact**
   - Which files were modified
   - File history snapshots for recovery

---

## Tips for Exploration

### Using claudewatch to Explore Sessions

claudewatch now extracts and displays:
- **STARTED**: Session start time
- **LENGTH**: Total duration
- **USER**: Number of user prompts
- **INT**: Number of interruptions/resumptions (detected by 1+ hour gaps)

To add more analytics:
1. Parse `message.usage` for token tracking
2. Analyze `message.content` array for tool usage patterns
3. Cross-reference `cwd` changes to understand project context switching
4. Group by `gitBranch` to see work per branch

### Common Queries

```bash
# Count total messages in all sessions
find ~/.claude/projects -name "*.jsonl" -exec wc -l {} \; | awk '{s+=$1} END {print s}'

# Find sessions with token usage > 1M
find ~/.claude/projects -name "*.jsonl" -exec jq 'select(.type=="assistant" and .message.usage.input_tokens > 1000000)' {} \;

# Export all user prompts for analysis
find ~/.claude/projects -name "*.jsonl" -exec jq -r 'select(.type=="user") | .message.content' {} \;
```

---

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1 | 2026-01-30 | Current format. Includes: user messages, assistant responses with usage, system events, file snapshots |

See `version` field in `sessions-index.json` and individual events for format version.

---

## Related Files

- **[internal/monitor/session_parser.go](./internal/monitor/session_parser.go)** - Go code that parses JSONL
- **[internal/ui/model.go](./internal/ui/model.go)** - claudewatch SessionInfo structure
- **[CLAUDE.md](./CLAUDE.md)** - Claude Code CLI user instructions
