# Bluesky Suggest List Members

A small CLI tool that helps you discover new accounts to add to lists in Bluesky.

Given a Bluesky handle and list name, the tool will:
- Retrieve members of the list
- Analyze who those members follow
- Suggest new accounts to add — those accounts followed by multiple list members
- Output the suggestions, sorted by popularity

## 🚀 Installation
(Comming soon)

## Usage
1. retrieve lists
   
`SuggestListMembers --username foo.bsky.social`

3. suggest list members

`SuggestListMembers --username example.bsky.social --list coffeeLovers`

### Flags

| Flag    | Description |
| -------- | ------- |
| `--username` | A Bluesky handle |
| `--list` | (Optional) Name of the list to suggest new accounts for     |
| `-j` Or `-json` | Enable JSON output |
