# Contributing

## Branching and pull requests

`main` is the production-ready branch. Do not push directly to it. `dev` is the
persistent integration and testing branch; all changes must reach `dev` before
they can be promoted to `main`.

1. Update `dev`, then create a feature branch from it:

   ```sh
   git switch dev
   git pull --ff-only
   git switch -c feature/short-description
   ```

2. Implement and test the change locally.
3. Open a pull request from the feature branch into `dev` and merge only after
   its required CI checks pass.
4. Verify the integrated change on `dev`. Pushes to `dev` run the same CI checks
   as its pull requests.
5. When the integration version is ready for production, open a pull request
   from `dev` into `main`.
6. Merge the `dev`-to-`main` pull request only after all required checks pass.
   Production container publishing runs only from `main`.

Use the pull-request template to document testing, UI evidence, database or
migration effects, and deployment considerations.

## Required local checks

Run these before opening a pull request:

```sh
test -z "$(gofmt -l $(git ls-files '*.go'))"
go vet ./...
go test -run '^$' ./...
go test ./...
go build -trimpath ./...
docker build --tag amex-grocery-splitter-se:ci .
```

There is currently no database, Prisma schema, or migration tooling in this
repository, so no database validation command is required. Add the appropriate
validation to CI and this list if that tooling is introduced.

## GitHub branch protection (manual setup)

Repository branch-protection settings are managed in GitHub and must be enabled
there after the `dev` branch is first published.

### `main`

- Require a pull request before merging; require at least one approval if the
  repository has reviewers.
- Require status checks to pass before merging. Select the `Quality checks` CI
  check and require the branch to be up to date before merging.
- Restrict who can push to the branch, so direct pushes are blocked. Do not add
  routine contributors to the bypass list.
- Block force pushes and branch deletion.

### `dev`

- Require status checks to pass before merging pull requests. Select the
  `Quality checks` CI check and require the branch to be up to date before
  merging.
- Block force pushes and branch deletion.

In GitHub, configure these under **Settings → Branches** (or the corresponding
repository ruleset). Use branch patterns `main` and `dev` and ensure the rules
apply to administrators as appropriate for the repository.
