# Release Readiness P0 Implementation Plan

> **Required sub-skill:** Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Make the repository safe for precommit, lint/test, commit, push, local install validation, and CI monitoring by fixing the highest-priority release blockers.

**Architecture:** Keep the Go CLI as the source of truth, but make Node wrapper/tests deterministic by using explicit binaries and isolated state. Harden install/test scripts so failures are process failures, not log-only warnings. Extend CI/pre-commit coverage to include Node/release checks.

**Tech Stack:** Go 1.24, Cobra CLI, Node.js >=18 scripts, npm lifecycle scripts, GitHub Actions, pre-commit local hooks.

---

### Task 1: Make release contract deterministic and failing for current bugs

**Files:**
- Modify: `scripts/release-contract-test.js`

**Step 1: Write the failing test behavior**
- Replace shell-string `execSync` with `spawnSync` argument arrays.
- Support per-case expectations: JSON stdout, text stdout, JSON stderr, expected exit code.
- Build a temporary Go binary with `-ldflags` setting `internal/cli.Version` to `package.json` version.
- Run through `node bin/asearch.js` with `ASEARCH_BIN` and isolated `ASEARCH_STATE_DIR`.

**Step 2: Run test to verify it fails**
Run: `npm run release:contract`
Expected: FAIL because `node bin/asearch.js --fakearg` exits `0` while contract expects non-zero.

**Step 3: Implement minimal production fix**
Modify `cmd/asearch/main.go` to exit non-zero when `cli.Execute()` returns an error.

**Step 4: Run test to verify it passes**
Run: `npm run release:contract`
Expected: PASS.

---

### Task 2: Strengthen smoke test and wrapper binary selection

**Files:**
- Modify: `bin/asearch.js`
- Modify: `scripts/smoke-test.js`

**Step 1: Write failing smoke assertions**
- Build/use a temporary versioned Go binary.
- Set `ASEARCH_BIN` and isolated `ASEARCH_STATE_DIR`.
- Assert every check, including CLI version equals `package.json` version.
- Exit non-zero on any failed assertion.

**Step 2: Run smoke test**
Run: `npm test`
Expected before wrapper fix: either fail or prove current wrapper cannot explicitly select test binary.

**Step 3: Implement minimal wrapper fix**
- Add `ASEARCH_BIN` override.
- Prefer package-local binary before state binary for deterministic dev/test behavior.
- Remove unsafe PATH fallback or make it opt-in/realpath-safe.
- Handle spawn errors and signal exits.

**Step 4: Run smoke test**
Run: `npm test`
Expected: PASS, version `0.6.0`.

---

### Task 3: Harden install script silent-failure cases

**Files:**
- Modify: `scripts/install.js`

**Step 1: Add install-script checks through existing contract/smoke coverage where practical**
- Syntax check must pass.
- Keep postinstall behavior unsupported platforms as non-fatal.

**Step 2: Implement minimal hardening**
- Use `os.tmpdir()` for scratch dirs.
- Check standard redirects and resolve relative redirect URLs.
- Check `tar` spawn result.
- Verify extracted binary exists.
- Throw on download/extract/missing-binary failures instead of logging success.
- If an existing binary exists, compare `asearch version` to package version; replace stale binary.

**Step 3: Run syntax and package checks**
Run: `node --check scripts/install.js && npm pack --dry-run`
Expected: PASS.

---

### Task 4: Add Node/release checks to hooks and CI

**Files:**
- Modify: `.pre-commit-config.yaml`
- Modify: `.github/workflows/ci.yml`
- Modify: `.githooks/pre-push`
- Modify: `package.json`

**Step 1: Add lint/check script**
Add `npm run check` to execute `node --check` on Node entry points.

**Step 2: Add gates**
- Pre-commit: run Go gates and Node gates when relevant paths change.
- CI: add Node setup and run `npm run check`, `npm test`, `npm run release:contract`, `npm pack --dry-run`.
- Pre-push: run same key gates and fail on gofmt issues.

**Step 3: Verify**
Run all local verification commands.

---

### Task 5: Final verification, commit, push, install, CI/CD watch

**Files:**
- All changed files.

**Step 1: Run verification**
Run:
```bash
go test ./...
go vet ./...
go build ./...
npm run check
npm test
npm run release:contract
npm pack --dry-run
pre-commit run --all-files
```

**Step 2: Commit and push**
If verification passes:
```bash
git add ...
git commit -m "chore: harden release readiness checks"
git push
```

**Step 3: Install and validate**
Install the pushed/new package path locally as appropriate, then run:
```bash
asearch version
asearch --help
asearch doctor
```

**Step 4: Wait for CI/CD and inspect failures**
Use GitHub CLI if available:
```bash
gh run list --limit 5
gh run watch <run-id>
gh run view <run-id> --log-failed
```
