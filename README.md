# debravinh

[![CI](https://github.com/sudovinh/debravinh/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/sudovinh/debravinh/actions/workflows/ci.yaml)
[![Deploy](https://github.com/sudovinh/debravinh/actions/workflows/deploy-app.yaml/badge.svg?branch=main)](https://github.com/sudovinh/debravinh/actions/workflows/deploy-app.yaml)

Source for [debravinh.com](https://debravinh.com) — a linktree-style landing page for Debra & Vinh, plus an [/aboutus](https://debravinh.com/aboutus) page with our story.

## Stack

- **Go + [Echo v4](https://echo.labstack.com/)** — single-file server, both pages and all assets are compiled in via `go:embed` (no filesystem reads at runtime)
- **Plain HTML/CSS** — no JS frameworks, inline SVG icons, one shared stylesheet
- **DigitalOcean App Platform** — builds from the repo `Dockerfile`, spec lives in [`.do/app.yaml`](.do/app.yaml)
- **Terraform** — manages the DO app and the `debravinh.com` domain ([`terraform/`](terraform/))
- **GitHub Actions** — CI, PR preview deploys, and production deploys

## Local development

Tooling is managed with [Flox](https://flox.dev) (`flox activate` drops you into a shell with go, terraform, doctl, op, and gh). Or bring your own Go ≥ 1.26.

```shell
make run          # start the server on :8080 (PORT env var to override)
make test         # run the test suite
make check        # vet + test + build, run before pushing
make vuln         # govulncheck dependency scan
make docker-build # build the production image locally
make docker-run   # run it on :8080
```

## CI/CD

```
PR opened ──> CI (vet/test/build/govulncheck + docker smoke test)
          └─> preview app deployed on DO, URL commented on the PR
PR closed ──> preview app deleted
merge to main ──> deploy-app workflow ──> DigitalOcean rebuilds & deploys
```

Deploys use [`digitalocean/app_action/deploy@v2`](https://docs.digitalocean.com/products/app-platform/how-to/deploy-from-github-actions/) with the app spec from `.do/app.yaml`. There's no container registry — DigitalOcean builds the Dockerfile from source.

## Security

- Strict security headers (CSP, HSTS, X-Frame-Options, nosniff, referrer/permissions policy) set in middleware and asserted in tests
- Request body limit and server timeouts
- `govulncheck` runs in CI; GitHub Actions are SHA-pinned with least-privilege permissions
- Container runs as `nobody` on an up-to-date Alpine base

## Infrastructure

The DO app (`debravinh-com`) and the `debravinh.com` domain are managed in [`terraform/`](terraform/). Secrets come from 1Password via `op run` — nothing sensitive is stored in the repo.

```shell
make terraform-init
make terraform-plan
make terraform-apply
```
