# CLAUDE.md

## Environment: WSL2

This project runs on WSL2 (Ubuntu) under Windows. Some tools aren't on PATH by default.

### Node.js / npm

Node is managed via **nvm**. It must be sourced before use:

```bash
export NVM_DIR="$HOME/.nvm" && [ -s "$NVM_DIR/nvm.sh" ] && . "$NVM_DIR/nvm.sh"
```

Without this, `node`/`npm`/`npx` won't be found. The git hooks (`.githooks/pre-commit`) already handle this for the web build step.

### Playwright

Playwright requires system libraries for Chromium. Install them with:

```bash
sudo apt-get install -y libnspr4 libnss3 libcups2t64 libxkbcommon0 libatspi2.0-0t64 libxdamage1 libgbm1 libpango-1.0-0 libcairo2 libasound2t64
```

Then install Playwright browsers: `npx playwright install chromium`

### Go

Go is available on PATH normally (no special sourcing needed).

## Project Structure

- **Go backend**: `cmd/bot/` (Discord bot), `cmd/web/` (web server), `cmd/worker/` (background worker)
- **React frontend**: `web/` — built with Vite + Tailwind CSS
- **Go embed**: `cmd/web/main.go` embeds `cmd/web/dist/` (the built frontend). The pre-commit hook builds the frontend into this path when `web/` files are staged.

## Pre-commit Hook

Located at `.githooks/pre-commit`. Runs selectively based on staged files:
- `.go` files → gofmt + `go build ./cmd/bot/...`
- `web/` files → `npx vite build --outDir ../cmd/web/dist --emptyOutDir` + `go build -o /dev/null ./cmd/web`

## Deployment
- **Release**: `make release v=X.Y.Z` — tags and pushes, triggering GitHub Actions GoReleaser.

## Build Version

The frontend logs the git commit hash and date to the browser console at startup (injected via Vite `define` in `vite.config.ts`). Use this to verify which version is deployed.
