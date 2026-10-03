# araldo-site: notes for AI assistants

The araldo.dev website: `cmd/build` renders the landing page (`cmd/build/site/index.html`) and the
app repo's docs into `dist/`. Read README.md first.

- **This repo is public.** Never name private deployments, deploy repos, infrastructure repos,
  hostnames other than araldo.dev, or secrets paths.
- **Docs belong in the app repo.** Fix a doc there, not here; this repo only renders them.
- **Claims on the landing page must be true of the app's `main`.** Check the app repo before
  describing a feature, a platform or a status.
- **No JavaScript** unless a page truly needs it; pages work as plain HTML and CSS, in light and
  dark, at WCAG AA contrast, with one `h1` each.
- Every Go file starts with `// SPDX-License-Identifier: AGPL-3.0-or-later`. `task check` must pass.
