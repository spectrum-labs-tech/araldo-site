# araldo.dev

The website for [Araldo](https://github.com/spectrum-labs-tech/araldo), the open-source distribution
API: a landing page, and the app repo's documentation rendered as HTML.

The docs are not copied here. They live in the app repo, next to the code they describe, and the
build reads them from a checkout of it: `README.md`, `docs/*.md`, the decision records in
`docs/adr/`, `CONTRIBUTING.md`, `SECURITY.md` and the OpenAPI contract. Links between documents
become links between pages; links into the code go to GitHub. To change a doc, change it in the app
repo; the site picks it up within a day, or at once by running the Deploy workflow.

## Build it

You need Go (version in `go.mod`) and [Task](https://taskfile.dev), and the app repo checked out
beside this one.

```sh
task build                 # dist/ from ../araldo
task build APP=/path/to/araldo
task serve                 # build, then look at it on http://localhost:8000
task check                 # format, vet and tests
```

## Deploy

`.github/workflows/deploy.yaml` builds on every push to `main`, daily and by hand, and publishes
`dist/` to the Cloudflare Pages project `araldo-site` with the `CLOUDFLARE_API_TOKEN` and
`CLOUDFLARE_ACCOUNT_ID` secrets. Pull requests build and test but publish nothing.

## License

[AGPL-3.0-or-later](LICENSE), like Araldo itself.
