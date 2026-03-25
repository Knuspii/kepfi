<p align="center">
  <a href="https://go.dev/"><img src="https://img.shields.io/static/v1?label=Made%20with&message=Go&logo=go&color=007ACC" alt="Golang" /></a>
  <a href="https://goreportcard.com/report/github.com/knuspii/kepfi"><img src="https://goreportcard.com/badge/github.com/knuspii/kepfi" alt="Go Report Card" /></a>
  <a href="https://github.com/knuspii/kepfi/actions/workflows/go.yml"><img src="https://github.com/knuspii/kepfi/actions/workflows/go.yml/badge.svg" alt="Build" /></a>
  <a href="https://github.com/knuspii/kepfi/stargazers"><img src="https://img.shields.io/github/stars/knuspii/kepfi?style=social" alt="GitHub Stars" /></a>
  <br>
  <img src="https://img.shields.io/github/license/knuspii/kepfi" />
  <img src="https://img.shields.io/badge/Platform-Linux-blue?logo=linux&logoColor=white" alt="Platform" />
</p>

<div align="center">
  <h1><code>kepfi</code></h1>
  A smart alternative to rm with a recovery bin and storage tracking.
</div>

## 🚀 Features & Usage
**Usage: `kepfi <flag>`**
| Flag | Action | Description |
| :--- | :--- | :--- |
| `<name>` | **Move Trash** | Moves files/folders to the `kephi trash`. |
| `-at <HH:MM>` | **Schedule Purge** | Launches a background process to clear `kephi trash` at a specific time. |
| `-r <name>` | **Restore** | Moves a file/folder back to its original location. |
| `-l` | **List Trash** | Shows a detailed table of `kepfi trashed` items. |
| `-t` | **Temp Move** | Moves files/folders to `/tmp/` instead of the `kepfi trash`. |
| `-pa` | **Purge All** | Permanently deletes everything in the `kepfi trash` directory. |
| `-ps <name>`| **Purge Specific** | Permanently deletes one specific file/folder from the `kephi trash`. |
| `-f` | **Force** | Skips the "y/N" confirmation prompt for purge actions. |
| `-v` | **Version** | Shows current version and credits. |

## 📥 Easy Install
```
curl -sSL https://raw.githubusercontent.com/Knuspii/kepfi/main/install.sh | sudo bash
```

## 📂 Directory Structure
```
~/.local/share/kepfi/
├── trash/           # This is where your 'deleted' files actually live
└── metadata.json    # The "brain" that remembers original paths and timestamps
```
