# araldo-site: notes for AI assistants

The araldo.dev websites: `cmd/build` renders the landing page (`cmd/build/site/index.html`) into
`dist/site` and the app repo's docs into `dist/docs` (docs.araldo.dev). Read README.md first.

- **This repo is public.** Never name private deployments, deploy repos, infrastructure repos,
  secrets paths, or hostnames outside araldo.dev. The hosted plan is public: the landing page sells
  it, links to app.araldo.dev and account.araldo.dev, and shows its prices, which must match the
  plans the hosted service bills (a price change lands here too).
- **Docs belong in the app repo.** Fix a doc there, not here; this repo only renders them.
- **Claims on the landing page must be true of the app's `main`.** Check the app repo before
  describing a feature, a platform or a status.
- **No JavaScript** unless a page truly needs it; pages work as plain HTML and CSS, in light and
  dark, at WCAG AA contrast, with one `h1` each.
- Every Go file starts with `// SPDX-License-Identifier: AGPL-3.0-or-later`. `task check` must pass.
