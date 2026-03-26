<p align="center">
  <a href="https://goreportcard.com/report/github.com/knuspii/kepfi"><img src="https://goreportcard.com/badge/github.com/knuspii/kepfi" alt="Go Report Card" /></a>
  <a href="https://github.com/knuspii/kepfi/actions/workflows/go.yml"><img src="https://github.com/knuspii/kepfi/actions/workflows/go.yml/badge.svg" alt="Build" /></a>
  <a href="https://github.com/knuspii/kepfi/stargazers"><img src="https://img.shields.io/github/stars/knuspii/kepfi?style=social" alt="GitHub Stars" /></a>
  <br>
  <img src="https://img.shields.io/badge/Platform-Linux-blue?logo=linux&logoColor=white" alt="Platform" />
</p>

<div align="center">
  <h1><code>kepfi</code></h1>
  <p>A smart alternative to rm with a recovery bin and storage tracking.</p>
  <img src="assets/preview.png" width="200" height="200" alt="Preview">
</div>

## 🚀 Features & Usage
```
Usage: kepfi [option]

Options:
  -l           Shows a detailed table of kepfi trashed items
  -r <file>    Restores a file/folder back to its original location
  -t <file>    Move a file/folder to /tmp/
  -ps <file>   Purge specific file/folder in kepfi trash
  -pa          Purge all files/folders from kepfi trash
  -f           Force action (no confirmation)
  -at <HH:MM>  Schedule a one-time purge at a specific time
  -v           Display version

  Examples:
  kepfi file.txt        Move file.txt to kepfi trash
  kepfi -r file.txt     Restore file.txt to its original path
  kepfi -at 22:30       Schedule a background purge for 22:30
```

## 📥 Easy Install
```
curl -sSL https://raw.githubusercontent.com/knuspii/kepfi/main/install.sh | sudo bash
```
You can also download kepfi from the [Releases](https://github.com/knuspii/kepfi/releases) \
[![Download](https://img.shields.io/github/downloads/knuspii/kepfi/total?color=green)](https://github.com/knuspii/kepfi/releases)

## 📂 Directory Structure
```
~/.local/share/kepfi/
├── trash/           # This is where your 'deleted' files actually live
└── metadata.json    # The "brain" that remembers original paths and timestamps
```

> ### 💀 `rm` is mid. `kepfi` is the glow-up.
> Using `rm` in 2026 is low-key traumatic. It’s giving "I accidentally deleted my entire project..."
