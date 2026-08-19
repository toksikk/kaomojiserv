# Repository Notes

## Runtime

- This is one Go 1.21 `main` package; `main.go` wires every route and `observability.go` wraps the single `http.ServeMux` with access logging and Prometheus metrics.
- The HTML template and kaomoji list are runtime files, not embedded. Run from the repository root or pass `-template` and `-kaomojis`; startup panics if either path is unavailable. An empty kaomoji file also panics when selecting the initial item.
- Local development needs an unprivileged port: `go run . -port 8080`. The defaults are port `80`, timeout `60`, `kaomojis.txt`, and `kaomoji_template.html`.
- Rotation is lazy and shared across clients: only a request handled by `/` checks the timeout and chooses a new kaomoji. `/raw` and `/api` return the current selection without rotating it.
- The root `"/"` ServeMux pattern is also the fallback for unknown paths. Preserve route labels from `ServeMux.Handler`; observability tests expect labels such as `/api` rather than raw URL paths.
- `make build` injects `main.version` and `main.builddate` with linker flags and writes `bin/kaomojiserv`. Plain `go build`/`go run` leaves build metadata empty or derived from Go build info.

## Verification

- Full tests: `go test ./...`
- Focus one test: `go test -run '^TestObserveHTTPLogsAccessAndRecordsMetrics$' .`
- Match the pre-commit Go checks with `gofmt -w <changed.go files>`, `golangci-lint run`, and `go test ./...`. If `go.mod` or `go.sum` changed, also run `go mod verify`.
- Template tests read `kaomoji_template.html` from the working directory and assert exact JavaScript/accessibility markers; run tests from the repository root after template changes.

## Content And Commits

- Add one kaomoji per line in `kaomojis.txt`; blank lines are not filtered and become selectable entries.
- Commit messages follow Scoped Commits: `scope(optional issue): change description`. The mandatory lowercase `scope` names the affected subsystem or module, the issue is optional, and type prefixes are not used. Examples: `ui: add guide` and `observability(#42): record request metrics`.
