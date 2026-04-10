# Dcat Admin Module Mapping

This project references the local `dcat-admin` source tree and keeps the same primary module vocabulary while translating implementation details into Go.

## Mapping

| Dcat Admin module | Go Admin Build equivalent | Status |
| --- | --- | --- |
| `Layout/Content` | `goadmin.App` layout shell + embedded templates | implemented |
| `Grid` | `goadmin/grid.Builder` + generic list rendering | implemented |
| `Form` | `goadmin/form.Builder` + generic create/update flow | implemented |
| `Show` | `goadmin/show.Builder` + detail rendering | implemented |
| `Tree` | `goadmin/tree.Builder` + `TreeProvider` + GORM tree config | implemented |
| `Repositories` | `goadmin.Repository` + `goadmin/store/gormstore.Repository[T]` | implemented |
| `Models/AdminTablesSeeder` | `goadmin/auth` + `demo/seed` | implemented |
| `Auth/Permission/Menu` | `goadmin/auth.Service` + sidebar navigation | implemented |
| `Actions/Tools` | built-in row actions and page actions | partial |
| `Widgets` | dashboard cards and layout panels | partial |
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
- grid search/filter/sort/pagination
- form create/update
- show/detail pages
- tree rendering
- gin + sqlite runnable demo

## Next expansion candidates

1. permission and menu management pages backed by the built-in auth tables
2. richer field types such as uploads, switches, tags, date/time pickers
3. relation editors for many-to-many and nested forms
4. extension hooks, widget registry, and code generation
