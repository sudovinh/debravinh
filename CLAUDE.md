# CLAUDE.md

## What this is

debravinh.com — a linktree-style personal site for Debra & Vinh. Go + Echo v4 server with two static HTML pages (`/` and `/aboutus`), everything embedded in the binary.

## Architecture

- `main.go` — the whole server. `newServer()` registers all middleware and routes; `main()` only does config + start/graceful shutdown. Pages and assets are compiled in via `go:embed` (no filesystem reads at runtime).
- Routes: `/` (landing), `/aboutus`, `/assets/*` (static), `/robots.txt`, anything else 307-redirects to `/`.
- `web/views/*.html` — plain HTML, no templating. `web/assets/css/style.css` — single shared stylesheet for both pages. No JS frameworks, icons are inline SVG.
- Tests in `main_test.go` cover every route plus the security headers. Write tests first when changing routes/behavior.

## Commands

Tooling comes from Flox (`flox activate`, manifest committed in `.flox/`).

- `make run` — serve on :8080 (`PORT` env to override)
- `make check` — vet + test + build (run before pushing)
- `make vuln` — govulncheck scan
- `make docker-build` / `make docker-run`
- `make terraform-plan` / `make terraform-apply` — uses `op run` with `terraform/.env.terraform` (needs a signed-in 1Password CLI)

## Deployment

- PR → `ci.yaml` (vet/test/build/govulncheck + docker smoke test) and `deploy-preview.yaml` (DO preview app, URL commented on the PR). Closing the PR deletes the preview.
- Merge to main → `deploy-app.yaml` runs `digitalocean/app_action/deploy@v2`, which deploys from source using `.do/app.yaml`. DigitalOcean builds the repo `Dockerfile`. No container registry.
- The DO app is named `debravinh-com`; the `debravinh.com` DNS zone lives in DO and is attached via the app spec.
- **Spec lives in two places:** `.do/app.yaml` (used by app_action) and `terraform/main.tf` (owns the infra; existing app + domain were imported, not recreated). Keep them in sync when changing instance size, health checks, domains, etc.

## Gotchas

- **Port must stay 8080** — the DO app spec, Dockerfile EXPOSE, and the server default all assume it.
- Preview apps cost money while their PR is open; `delete-preview.yaml` cleans up on close.
- HSTS is only sent on HTTPS requests (`X-Forwarded-Proto: https` behind DO's proxy) — that's correct behavior, don't "fix" it.
- `curl -I` returns 405 (Echo doesn't register HEAD); use `curl -s -D - -o /dev/null` to inspect headers.
- A leftover `go run` child process can keep port 8080 after killing the parent — `pkill debravinh`.

## Conventions

- Branches: `vinhn.feat_<topic>`. Commits: `feat(path):` / `conf(path):`, short lowercase subject, human tone, brief prose body.
- Keep the security posture intact when changing code: CSP/secure headers middleware, body limit, server timeouts, SHA-pinned least-privilege workflows, tests asserting headers.
- YAGNI — plain HTML/CSS, no databases, no JS frameworks.
