# Security Audit — Network & Data Exfiltration Analysis

**Date:** 2026-03-04  
**Scope:** All outbound network calls and data exfiltration possibilities  
**Auditor:** GitHub Copilot

---

## Summary

**No direct HTTP/HTTPS calls exist in the Go source code.** There are no imports of `net/http`, `net`, or any HTTP client library in any `.go` file. All outbound network activity is delegated exclusively to the **Azure CLI subprocess** via `os/exec`.

---

## 1. Outbound Network Calls — `internal/azure/azure.go`

All three functions use `exec.Command("az", ...)` to shell out to the Azure CLI binary. The CLI itself then makes authenticated HTTPS calls to Azure endpoints.

| Function | CLI Command Executed | Azure Endpoint Reached | Data Sent | Data Returned |
|---|---|---|---|---|
| `AzShowKeyvault` | `az keyvault show --name … --subscription …` | `management.azure.com` | keyvault name, subscription ID | keyvault metadata JSON |
| `AzGetSecrets` | `az keyvault secret list --vault-name … --subscription …` | `<vaultname>.vault.azure.net` | vault name, subscription ID | secret names + IDs JSON |
| `AzShowSecret` | `az keyvault secret show --vault-name … --name … --subscription …` | `<vaultname>.vault.azure.net` | vault name, secret name, subscription ID | **full secret value JSON** |

**No other outbound network calls exist in the codebase.**

---

## 2. Data Exfiltration Risk Assessment

### 🟡 MEDIUM — Unencrypted disk cache of secret metadata

**File:** `internal/cache/cache.go`  
`WriteKeyvault` and `WriteSecrets` write raw JSON responses to `bin/cache/<name>-kv.json` and `bin/cache/<name>-secrets.json`. Secret **names** and **IDs** are written to disk in plaintext. While `.gitignore` excludes `*bin/`, these files persist on disk between runs with world-readable permissions (`0644`).

---

### 🟡 MEDIUM — Secret values held in unbounded in-memory map

**File:** `internal/azure/azure.go`  
`AzShowSecret` stores full secret values (including the `value` field from the Azure API JSON response) in `SecretsStow map[string]string` for the lifetime of the process. There is no expiry, no zeroing on close, and no cap on the number of secrets cached. This map is a target for memory-dump attacks.

```go
// secret value is stored indefinitely, never cleared
az.SecretsStow[subscriptionId+vaultName+secretName] = string(out)
```

---

### 🟡 MEDIUM — Error output silently discarded on all CLI calls

**File:** `internal/azure/azure.go` — `AzShowKeyvault`, `AzGetSecrets`, `AzShowSecret`  
All three `exec.Command` calls use `CombinedOutput()` but discard the error return value (`out, _ := ...`). A failed or tampered CLI call (e.g. a rogue `az` binary on `$PATH`) would silently return empty/malformed output with no alerting.

---

### 🟢 LOW — `exec.Command` args are not shell-interpolated

The use of variadic `exec.Command("az", "keyvault", "secret", "show", ...)` args means **command injection is not possible** even if a secret name or vault name contained shell metacharacters. This is correctly implemented.

---

### 🟢 LOW — No telemetry, analytics, or beacon URLs

A full-text search found zero hardcoded URLs, no `http.Get`/`http.Post`, no websocket calls, and no embedded callback endpoints in any source file.

---

### 🟢 LOW — `golang.org/x/net` is transitive only

The `go.sum` references to `golang.org/x/net` are indirect dependencies pulled in by `tcell`/`tview` for terminal handling. No application code imports or uses it.

---

## 3. Complete Network Surface

```
vaulty process
    └── os/exec → az CLI binary (Azure CLI)
                    ├── HTTPS → management.azure.com  (keyvault metadata)
                    └── HTTPS → *.vault.azure.net     (secrets list & values)
```

There is no other outbound network surface.

---

## 4. Recommendations

| Priority | Recommendation | File |
|---|---|---|
| High | **Zero the `SecretsStow` map on exit** — overwrite each entry or set `az.SecretsStow = nil` when the application stops to reduce the memory-dump window for secret values. | `internal/azure/azure.go` |
| High | **Evict secret values after display** — remove the entry from `SecretsStow` when the user navigates away from the secret details view (i.e. in `CloseSecretDetailsView`). | `internal/ui/ui.go` |
| Medium | **Restrict `bin/cache/` file permissions** — `os.WriteFile` uses `0644`; change to `0600` to prevent other local users from reading cached secret metadata. | `internal/cache/cache.go` |
| Medium | **Check for `az` binary authenticity** — resolve the full path of the `az` binary at startup using `exec.LookPath` and log a warning if it is not in an expected system location, to guard against a rogue binary on `$PATH`. | `internal/azure/azure.go` |
| Low | **Handle exec errors explicitly** — replace `out, _ :=` with proper `out, err :=` checks on all three CLI calls to detect and surface failures. | `internal/azure/azure.go` |
