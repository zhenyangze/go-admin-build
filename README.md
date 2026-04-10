# Go Admin Build

`Go Admin Build` is a Dcat Admin inspired admin framework implemented in Go.

It keeps the page-builder workflow of `Grid`, `Form`, `Show`, `Tree`, `Auth`, `RBAC`, and `Menu`, but the runtime architecture is Go-native:

- core app is `net/http` first
- demo mounts through `gin`
- persistence uses `gorm`
- demo database uses `sqlite`
- UI is built with `tailwindcss`

## What is included

- reusable package under `goadmin/`
- built-in auth/RBAC/menu schema under `goadmin/auth`
- generic GORM repository under `goadmin/store/gormstore`
- DSL builders for `grid`, `form`, `show`, and `tree`
- runnable demo under `cmd/demo`
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
PORT=8091 go run ./cmd/demo
```

Open `http://127.0.0.1:8091/admin`

- username: `admin`
- password: `admin`

## Demo modules

- `/admin/users`
- `/admin/roles`
- `/admin/articles`
- `/admin/categories`
- `/admin/categories/tree`

## Embedding in another Go app

The framework itself is framework-agnostic because it exposes a standard `http.Handler`.

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
