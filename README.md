# Vaulty

[![Security Audit](https://img.shields.io/badge/security-audited-green)](SECURITY_AUDIT.md)

An Azure Keyvault TUI, written in Golang using [tview](https://github.com/rivo/tview/tree/master). 

![Screenshot](vaulty.gif)

Vaulty is under active development and is subject to change. 

## Features

- Browse multiple Azure Key Vaults and their secrets from a single TUI
- Secrets list auto-focuses on startup — navigate without touching the mouse
- Switch between vaults instantly with number keys (`1`–`9`)
- Fuzzy search across secrets with `/`
- View a secret's value with `d` or `Enter` — a loading overlay is shown while fetching
- Re-fetch the latest secret value at any time with `r` (bypasses in-memory cache)
- Copy the secret value directly to your clipboard with `y`
- Sync all vault metadata and secret lists from Azure with `Ctrl+R`
- Secret values are cached in-memory per session; list metadata is cached on disk

## Keybindings

| Key | Action |
|-----|--------|
| `d` / `Enter` | Show secret value |
| `b` / `Esc` | Close secret view |
| `y` | Copy secret value to clipboard |
| `r` | Re-fetch secret (bypasses cache) |
| `/` | Search secrets |
| `1`–`9` | Switch vault |
| `Ctrl+R` | Reload all vaults & secrets from Azure |
| `q` | Quit |

## Installation

### Pre-built binaries

Download the latest binary for your platform from the [Releases](https://github.com/Bizzaro/vaulty/releases) page.

Create your config file at `~/.vaulty.conf`:

``` yaml
Keyvaults:
  - Name: <keyvault-name>
    Subscription: <keyvault-subscription-id>
  - Name: <keyvault-name>
    Subscription: <keyvault-subscription-id>
```

Then run the binary:

``` bash
./vaulty
```

### Building from source

1. Clone the repository
2. Create `~/.vaulty.conf`
   - You must configure at least one Keyvault. Remove any unused Keyvaults from configuration.

``` yaml
Keyvaults:
  - Name: <keyvault-name>
    Subscription: <keyvault-subscription-id>
  - Name: <keyvault-name>
    Subscription: <keyvault-subscription-id>
  - Name: <keyvault-name>
    Subscription: <keyvault-subscription-id>
```
3. Build and run the executable

``` bash
cd vaulty
make && ./bin/vaulty
```

# Roadmap
- Bug fixing
- Global search
- Secret modification/deletion
- Certificates management
- Keys management
- Themes







