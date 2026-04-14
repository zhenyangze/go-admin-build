# Go Admin Build

`Go Admin Build` is a Dcat Admin inspired admin framework implemented in Go.

**This repository contains:**
- Demo application and documentation
- Dcat Admin reference source

**Core framework is at:** https://github.com/zhenyangze/goadmin

## Repository Structure

- `examples/demo/` - Demo application using the goadmin framework
- `docs/` - Documentation (architecture, module mapping, runbook)
- `dcat-admin/` - Reference PHP Dcat Admin source

## Quick Start

Install the core framework:
```bash
go get github.com/zhenyangze/goadmin
```

Build CSS:
```bash
npm install
npm run build:css
```

Run demo:
```bash
cd examples/demo
go run ./cmd/demo
```

## What is included

- reusable package under `goadmin/` (module: `github.com/zhenyangze/goadmin`)
- built-in auth/RBAC/menu schema under `goadmin/auth`
- audit logging with `LoginLog` and `AuditLog` (via repository hooks)
- generic GORM repository under `goadmin/store/gormstore`
- DSL builders for `grid`, `form`, `show`, and `tree`
- **25 grid displayers**: badge, label, image, progress, qr, switch, checkbox, radio, select, expand, modal, table, tree, etc.
- **22 widget components**: checkbox, radio, table, lazytable, terminal, chart, tab, dropdown, alert, etc.
- configurable grid page/row actions, QuickEdit, and ContextMenu
- server-side form validation with inline error feedback
- relation-backed nested editing demoed via `Article -> FAQs`
- reusable GORM relation-backed repeater repository helper for parent/child admin forms
- one parent resource can now host multiple relation-backed repeaters (demo: `Article -> FAQs + Links`)
- reusable GORM hooks/extension API for create/update/delete pipelines
- dashboard widgets with metric cards, quick actions, and list panels
- runnable demo under `examples/demo/` (module: `github.com/zhenyangze/go-admin-build/examples/demo`)
- server-side form validation with inline error feedback for required fields and date ranges
- architecture notes in [docs/architecture.md](/Users/zhenyangze/Documents/Company/git/go-admin-build/docs/architecture.md)
- Dcat mapping notes in [docs/module-mapping.md](/Users/zhenyangze/Documents/Company/git/go-admin-build/docs/module-mapping.md)

## Quick start

Install frontend build tooling once:

```bash
npm install
npm run build:css
```

Run tests:

```bash
go test ./...
```

Start the demo:

```bash
./scripts/run_demo.sh
```

Or run directly:

```bash
go run ./examples/demo/cmd/demo
```

Use a fresh SQLite database if you want to validate newly added demo schema/seed data:

```bash
DEMO_DB_PATH=tmp/demo/fresh-admin.db PORT=8091 DEMO_RESET=1 go run ./examples/demo/cmd/demo
```

Uploaded demo files are stored locally under `tmp/demo/uploads` and served back through `/admin/uploads/...`.
The current upload implementation supports extension/size checks, single-file replacement cleanup, delete cleanup, and multi-file CSV storage for lightweight demos.

Open `http://127.0.0.1:8091/admin`

- username: `admin`
- password: `admin`

## Demo modules

- `/admin/users`
- `/admin/roles`
- `/admin/permissions`
- `/admin/menus`
- `/admin/audit-logs`
- `/admin/login-logs`
- `/admin/articles`
- `/admin/projects`
- `/admin/tickets`
- `/admin/reports`
- `/admin/categories` (with tree view)
- `/admin/menus/tree`

## Embedding in another Go app

The framework package is `github.com/zhenyangze/goadmin` (independent of the demo).

```go
import "github.com/zhenyangze/goadmin"
```

It exposes a standard `http.Handler` and can be mounted in any framework:

```go
adminApp, _, err := demo.Build("tmp/demo/admin.db")
if err != nil {
	log.Fatal(err)
}

router := gin.Default()
router.Any("/admin", gin.WrapH(adminApp.Handler()))
router.Any("/admin/*any", gin.WrapH(adminApp.Handler()))
```

The same pattern can be used in `chi`, `echo`, or plain `net/http`.

## Demo verification

Run a smoke test against a fresh demo instance:

```bash
./scripts/demo_smoke.sh
```

See [docs/demo-runbook.md](/Users/zhenyangze/Documents/Company/git/go-admin-build/docs/demo-runbook.md) for the demo walkthrough.

The dashboard now acts as a guided demo landing page, including:

- module guide
- suggested walkthrough
- framework coverage summary
- demo helpers / reset-export guidance
- capability matrix
