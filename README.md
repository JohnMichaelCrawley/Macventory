<div align="center">

  <img src="Assets/logo.png" alt="Macventory logo" width="180">

  # Macventory

  **A clear, portable inventory of the software installed on your Mac.**

  [![Platform](https://img.shields.io/badge/platform-macOS-111827?style=flat-square)](#requirements)
  [![Go](https://img.shields.io/badge/Go-1.23%2B-00ADD8?style=flat-square&logo=go&logoColor=white)](#requirements)
  [![CI](https://img.shields.io/badge/CI-macOS-2088FF?style=flat-square&logo=githubactions&logoColor=white)](.github/workflows/ci.yml)
[![Licence](https://img.shields.io/badge/licence-PolyForm%20Noncommercial%201.0.0-22C55E?style=flat-square)](LICENSE)

</div>

## Overview

Macventory is a command-line tool that creates an inventory of the software installed on your Mac.

It records macOS applications, Homebrew packages, global language packages, developer tools, Docker resources and executables, then generates a timestamped Markdown report on your Desktop to save and review.

> [!IMPORTANT]
> Macventory is an inventory tool, not a backup or migration utility. It does not preserve applications, personal data, application settings, credentials or licences.

## Features

- Generates one portable, timestamped Markdown report on your Desktop
- Detects installed macOS applications and their versions
- Records Homebrew formulae, casks, taps and services
- Produces an embedded Brewfile for use as a reinstall reference
- Detects global npm, pnpm, Yarn, pipx, uv, RubyGems and Cargo packages
- Records Docker images, containers and volume names
- Captures major developer-tool and SDK versions
- Detects Visual Studio Code and Cursor extensions when their CLIs are available
- Lists executables from common user-controlled binary directories
- Redacts hardware serial numbers, UUIDs and provisioning identifiers
- Continues safely when optional tools are unavailable

## Requirements

- macOS
- Go 1.23 or newer when building from source
- Optional collectors:
  - Homebrew
  - `mas`
  - Docker
  - Visual Studio Code CLI
  - Cursor CLI
  - Language-specific package managers

Macventory still produces a report when optional collectors are unavailable.

## Installation

### Build from source

```bash
git clone https://github.com/johncrawley/macventory.git
cd macventory/Codebase

go build -o macventory .