# Dcat Admin Module Mapping

This project references the local `dcat-admin` source tree and keeps the same primary module vocabulary while translating implementation details into Go.

## Mapping

| Dcat Admin module | Go Admin Build equivalent | Status |
| --- | --- | --- |
| `Layout/Content` | `goadmin.App` layout shell + embedded templates | implemented |
| `Grid` | `goadmin/grid.Builder` + 25 displayers + configurable page/row actions | implemented |
| `Form` | `goadmin/form.Builder` + 50+ field types + validation | implemented |
| `Show` | `goadmin/show.Builder` + detail rendering | implemented |
| `Tree` | `goadmin/tree.Builder` + `TreeProvider` + drag-and-drop + batch actions | implemented |
| `Repositories` | `goadmin.Repository` + `goadmin/store/gormstore.Repository[T]` + hooks | implemented |
| `Models/AdminTablesSeeder` | `goadmin/auth` + `demo/seed` | implemented |
| `Auth/Permission/Menu` | `goadmin/auth.Service` + RBAC + audit logging + login logs | implemented |
| `Actions/Tools` | built-in CRUD actions + QuickEdit + ContextMenu + grid page/row actions | implemented |
| `Widgets` | 22 components: checkbox, radio, table, terminal, etc. | implemented |
| `Extend/Scaffold/Console` | not yet ported | not implemented |

## Design differences

- Dcat Admin is Laravel-native; this project is `net/http` native.
- Dcat Admin couples deeply to Eloquent; this project keeps a smaller repository contract and ships a GORM adapter.
- Dcat Admin uses Bootstrap/AdminLTE; this project uses TailwindCSS-generated styles.
- Dcat Admin supports a much larger field/widget ecosystem; this project currently focuses on the reusable core modules.

## Current verification surface

- login/logout with account lockout protection
- RBAC-aware menu tree with permission checks
- dashboard metrics with charts
- dashboard panels and quick actions
- grid search/filter/sort/pagination
- 25 grid displayers (badge, label, image, switch, expand, modal, etc.)
- configurable grid page actions and row actions
- QuickEdit inline editing
- ContextMenu actions
- form create/update with 50+ field types
- server-side validation with inline field errors
- upload validation, replacement cleanup, delete cleanup
- repeater/nested has-many style JSON-backed editor
- relation-backed nested editing demoed with article FAQs
- reusable GORM relation-backed repeater repository helper
- multiple relation-backed repeaters on one parent resource (demo: article FAQs + links)
- reusable repository hooks/extension API with audit logging
- show/detail pages with relations
- tree rendering with drag-and-drop
- tree batch operations
- 22 widget components (checkbox, radio, table, terminal, etc.)
- gin + sqlite runnable demo

## Next expansion candidates

1. password policy (strength validation, expiration)
2. code generator/scaffolding CLI
3. finer-grained action/tool/widget extension APIs
4. stronger upload adapters (cloud/object storage, image transforms, async processing)
5. more generic relation-backed nested editors beyond the article FAQ demo
