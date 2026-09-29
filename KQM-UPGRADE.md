# KQM update notes

The develop branch starts at upstream main commit
`00ff2396f2b79cbc05995799c765b06e181e292c`. The KQM UI and database use separate
packages. No upstream UI source package is replaced.

- `ui/packages/kqmsim`: KQM simulator and result viewer.
- `ui/packages/kqmembed`: share preview.
- `ui/packages/kqmworkers`: simulator hosting, share links, images, and WASM storage.
- `ui/packages/kqmdb`: database, submission form, review queue, archive import, and upgrade runner.

The simulator keeps the current upstream character and reaction implementations.
The manual character port restores Alyosha, Linnea, Lohen, Sandrone, Vesna,
Vodyanitsa, and Zibai. It also restores Albedo, Klee, and Venti Hexerei changes,
Diona Revelation changes, and configurable reaction bonuses. These characters
retain the assumptions and limits from the old fork.

The current upstream code already includes the other overlapping fork changes.
Do not overwrite current upstream implementations with old generated files.
`scripts/kqm-generate.py` restores the extra keys and ICD entries after the data
pipeline. If upstream adds one of these characters, review and remove the local
implementation before the next update. Keep UI names in the KQM packages.

## Local checks

From the repository root:

```sh
go test ./internal/... ./pkg/gcs/... ./pkg/reactable/... ./pkg/simulation/...
cd ui
pnpm install --frozen-lockfile
pnpm --filter '@gcsim/kqm*' typecheck
pnpm --filter @gcsim/kqmworkers test
pnpm --filter @gcsim/kqmdb test:worker
pnpm build:kqmsim
pnpm build:kqmdb
```

For simulator development, start `pnpm --filter @gcsim/kqmworkers exec wrangler
dev --local --port 8788`, then `pnpm --filter @gcsim/kqmsim dev`. The local UI
uses the local share store. Database setup and rerun instructions are in
[the database README](ui/packages/kqmdb/README.md).

## Deployment

Commit the tested changes first. Build both the native simulator and WASM from
that clean commit. The WASM build copies the runtime from the selected Go version.
The deploy script uploads the new WASM file to R2 before it updates the UI.
`main.wasm` is excluded from static assets because it exceeds the static file
size limit. Existing WASM versions and share data stay in their current stores.

Save the current Worker version IDs and export D1 before a deployment. Validate
the database upgrade on a copy of that export. Use KQM account credentials; clear
unrelated Cloudflare environment overrides on this machine. Deploy from `ui`:

```sh
env -u CLOUDFLARE_API_TOKEN -u CLOUDFLARE_ACCOUNT_ID pnpm deploy:kqmsim
env -u CLOUDFLARE_API_TOKEN -u CLOUDFLARE_ACCOUNT_ID pnpm deploy:kqmdb
```

The manual GitHub workflow needs a `CF_API_TOKEN` Actions secret for the KQM
account. It does not import database archives or publish rerun results. Those
operations require a separate backup and the commands in the database README.

For a UI fault, roll the Worker back to the saved version ID. A Worker rollback
does not restore D1 or R2 data. Keep the database backup and rerun history, and
restore only affected records after checking for later review decisions.
