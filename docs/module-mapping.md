# Dcat Admin Module Mapping

This project references the local `dcat-admin` source tree and keeps the same primary module vocabulary while translating implementation details into Go.

## Mapping

| Dcat Admin module | Go Admin Build equivalent | Status |
| --- | --- | --- |
| `Layout/Content` | `goadmin.App` layout shell + embedded templates | implemented |
| `Grid` | `goadmin/grid.Builder` + generic list rendering + configurable page/row actions | implemented |
| `Form` | `goadmin/form.Builder` + generic create/update flow + multiselect/switch/tags/date-range/datetime/upload/repeater fields | implemented |
| `Show` | `goadmin/show.Builder` + detail rendering | implemented |
| `Tree` | `goadmin/tree.Builder` + `TreeProvider` + GORM tree config | implemented |
| `Repositories` | `goadmin.Repository` + `goadmin/store/gormstore.Repository[T]` | implemented |
| `Models/AdminTablesSeeder` | `goadmin/auth` + `demo/seed` | implemented |
| `Auth/Permission/Menu` | `goadmin/auth.Service` + sidebar navigation + auth resource repositories | implemented |
| `Actions/Tools` | built-in CRUD actions + configurable grid page/row actions | partial |
| `Widgets` | dashboard cards, quick actions, and panel/list widgets | partial |
| `Extend/Scaffold/Console` | not yet ported | not implemented |
| `PJAX/LazyRenderable` | not yet ported | not implemented |

## Design differences

- Dcat Admin is Laravel-native; this project is `net/http` native.
- Dcat Admin couples deeply to Eloquent; this project keeps a smaller repository contract and ships a GORM adapter.
- Dcat Admin uses Bootstrap/AdminLTE; this project uses TailwindCSS-generated styles.
- Dcat Admin supports a much larger field/widget ecosystem; this project currently focuses on the reusable core modules.

## Current verification surface

- login/logout
- RBAC-aware menu tree
- dashboard metrics
- dashboard panels and quick actions
- grid search/filter/sort/pagination
- configurable grid page actions and row actions
- form create/update
- server-side validation with inline field errors
- upload validation, replacement cleanup, delete cleanup
- repeater/nested has-many style JSON-backed editor
- relation-backed nested editing demoed with article FAQs
- reusable GORM relation-backed repeater repository helper
- multiple relation-backed repeaters on one parent resource (demo: article FAQs + links)
- reusable repository hooks/extension API
- show/detail pages
- tree rendering
- gin + sqlite runnable demo

## Next expansion candidates

1. extension hooks, widget registry, and code generation
2. finer-grained action/tool/widget extension APIs
3. stronger upload adapters (cloud/object storage, image transforms, async processing)
4. more generic relation-backed nested editors beyond the article FAQ demo
