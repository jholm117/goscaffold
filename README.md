# goscaffold

[![CI](https://github.com/jholm117/goscaffold/actions/workflows/ci.yml/badge.svg)](https://github.com/jholm117/goscaffold/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/jholm117/goscaffold)](https://github.com/jholm117/goscaffold/releases/latest)

Scaffold Go projects with composable layers (CLI, K8s controller, Helm chart) following consistent engineering standards.

## Install

```bash
brew install jholm117/tap/goscaffold
```

## Usage

```bash
# Create a CLI project
goscaffold init myapp --cli

# Create a K8s controller with Helm chart
goscaffold init mycontroller --controller --helm

# Create a project with everything
goscaffold init myproject --cli --controller --helm

# Add a layer to an existing project
cd myapp
goscaffold add controller
```

## Configuration

Create `~/.config/goscaffold/config.yaml`:

```yaml
module_prefix: github.com/youruser
homebrew_tap_token: "<GitHub PAT with contents:write on homebrew-tap repo>"
```

## Development

| Command | Description |
|---|---|
| `make build` | Build binary |
| `make test` | Run tests |
| `make lint` | Run golangci-lint |
| `make setup-hooks` | Install pre-push hook |
| `make help` | Show all targets |
