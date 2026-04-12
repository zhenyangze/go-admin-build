# Demo Runbook

## Goal

Provide a single runnable admin demo that proves the framework works end-to-end with:

- Gin mounting
- SQLite persistence
- GORM-backed resources
- auth / RBAC / menu
- grid / form / show / tree
- uploads
- nested relation-backed repeaters

## Recommended commands

### Start a fresh demo

```bash
./scripts/run_demo.sh
```

By default this uses:

- `PORT=8091`
- `DEMO_DB_PATH=tmp/demo/fresh-admin.db`
- `DEMO_RESET=1`

### Smoke test the demo

```bash
./scripts/demo_smoke.sh
```

## Login

- URL: `http://127.0.0.1:8091/admin`
- username: `admin`
- password: `admin`

## Suggested manual walkthrough

1. Login and verify the dashboard cards/panels render.
   - `Quick Access`
   - `Recent Articles`
   - `Module Guide`
   - `Suggested Walkthrough`
   - `Framework Coverage`
   - `Demo Helpers`
   - `Capability Matrix`
   - `Operational Overview`
2. Open `Demo Center` and verify:
   - fresh reset command
   - export shortcut
   - module verification links
3. Open `Users`, `Roles`, `Permissions`, and `Menus`.
4. Open `Articles` and verify:
   - image preview
   - gallery count
   - FAQ count
   - links count
   - tags badges
5. Open `Projects` and verify:
   - seeded milestones render
   - status badges render
   - detail page shows milestone list
6. Open `Audit Logs` and verify:
   - list page is read-only
   - detail page shows actor/resource/detail
7. Open `Tickets` and verify:
   - belongs-to project name renders
   - status/priority render
   - detail page shows assignee/project/due date
8. Open `Reports` and verify:
   - read-only metric/report snapshots
   - trend badge rendering
   - detail page shows metric/dimension/value
9. Create a new article with:
   - image upload
   - gallery upload
   - FAQ rows
   - link rows
10. Create a new project with milestones.
11. Create a new ticket and assign it to a project.
12. Open `Categories` tree and `Menus` tree.

## Notes

- Demo uploads are stored adjacent to the selected DB path in an `uploads/` folder.
- Setting `DEMO_RESET=1` removes the selected SQLite DB and the sibling uploads directory before startup.
