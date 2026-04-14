# Go Admin Framework Architecture

## Goal

Build a Go package inspired by Dcat Admin's developer experience while keeping the runtime architecture idiomatic for Go:

- `net/http` first so it can be mounted by Gin, Chi, Echo, Fiber adapters, or plain stdlib.
- `gorm` as the default persistence adapter, but not required by the core package.
- `tailwindcss` for the UI layer instead of the original Bootstrap/AdminLTE stack.
- resource-oriented DSL for `Grid`, `Form`, `Show`, `Tree`, `Layout`, `Auth`, `RBAC`, and `Menu`.

## Dcat Admin parity targets

The local `dcat-admin` source shows these major capability areas:

1. page shell and layout
2. grid/list builder
3. form builder
4. show/detail builder
5. tree builder
6. repository abstraction
7. auth, permission, role, menu
8. actions/tools/widgets/extension points

This implementation focuses on delivering a complete, runnable vertical slice for the most important modules:

- layout shell
- auth/login/session
- RBAC-based menu visibility
- resource registration
- grid with search/filter/sort/pagination/actions
- form with typed fields and validation hints
- show/detail page
- tree page
- generic GORM repository
- Gin demo backed by SQLite

## Package layout

```text
goadmin/
  app.go               # app bootstrap, routing, rendering, session/auth glue
  config.go            # config and app options
  contracts.go         # repository/auth/navigation contracts
  resource.go          # resource registration and page metadata
  auth/                # built-in GORM auth/RBAC/menu models + service
  audit/               # audit logging hooks and models
  form/                # form builder DSL with 50+ field types
  grid/                # grid builder DSL with 25 displayers
  show/                # show builder DSL
  store/gormstore/     # generic GORM CRUD/tree repository
  theme/               # theme defaults and UI tokens
  tree/                # tree builder DSL
  widgets/             # UI widgets (22 components)
    checkbox/          # checkbox widget
    radio/             # radio widget
    table/             # table widget
    lazytable/         # lazy loading table widget
    terminal/          # terminal/console widget
    ...
```

## Why `net/http` first

Making the app itself implement `http.Handler` keeps the package portable:

- Gin demo mounts with `gin.WrapH(adminApp.Handler())`
- Other frameworks can mount the same handler or proxy to it
- tests can use `httptest` directly

## Constraints

- Do not mirror Dcat Admin's PHP internals one-to-one; mirror the module surface and developer workflow.
- Keep the first version dependency-light: only `gin`, `gorm`, `sqlite`, `tailwindcss`, and secure password hashing support.
- Keep the first diff reviewable and demoable end-to-end.

## Demo scope

The demo should prove:

1. login/logout works
2. sidebar menu is permission-aware
3. dashboard renders
4. grid pages render and support search/filter/sort/pagination
5. form pages can create/update records
6. show pages render record details
7. tree page renders hierarchical data

## Known non-goals for this iteration

- full plugin marketplace
- pjax/asynchronous partial-refresh UX parity
- code generator/scaffolding CLI parity

## Recently Implemented

### Grid Displayers (25 total)
All major displayers are now implemented:
- **Visual**: Badge, Label, Image, ProgressBar, QRCode
- **Interactive**: SwitchDisplay, Checkbox, Radio, Select, SwitchGroup
- **Input**: Input, Textarea, Editable
- **Navigation**: Link, Downloadable, Button, DropdownActions
- **Layout**: Expand, Modal, Limit, Copyable
- **Data**: Table, Tree, DialogTree, Orderable

### Widgets (22 total)
Dashboard and form widgets:
- **Form**: Checkbox, Radio, Form
- **Display**: Table, LazyTable, Card, Box
- **Feedback**: Alert, Callout, Tooltip, Modal
- **Content**: Markdown, Code, Terminal
- **Navigation**: Tab, Dropdown, Tree
- **Utilities**: Async, Lazy, Dump, DarkModeSwitcher
