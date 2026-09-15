# oxzoo-wasm

An official ox deploy example: the page's greeting is computed inside a WebAssembly module compiled from Go (`GOOS=js GOARCH=wasm`), backed by a tiny stdlib-only Go API server. ox runs the install hooks in the release directory — one `go build` for the native server, one cross-compile for the wasm module with the greeting baked in via `-ldflags -X`, and a copy of the toolchain's `wasm_exec.js` glue — starts the compiled `./server` binary as a systemd process bound to `127.0.0.1:9115`, and configures nginx to serve `public/` statically while proxying only `/api` and `/health` to the Go process. Everything is driven by one `ox.toml` manifest at the repo root.

## Stack

| Component | Version | Purpose |
|---|---|---|
| Wasm module | Go 1.27.0 (`GOOS=js GOARCH=wasm`) | computes the greeting in the browser via `syscall/js` |
| API | Go stdlib `net/http` (Go 1.27.0) | `GET /api/greeting` and `GET /health`, binds `127.0.0.1:9115` |
| Frontend runtime | `wasm_exec.js` from the Go toolchain + vanilla JS | instantiates `greeting.wasm` and runs it |
| Backend build | `go` (apt `go`) | builds `./server` and `public/greeting.wasm` during the install hooks |
| Deploy | ox | `ox.toml` defines processes, frontend, domain |

## Environment flow

One variable, two paths:

**`GREETING_TAG`**

- **Runtime path (API):** `cmd/server/main.go` reads `os.Getenv("GREETING_TAG")` on each `/api/greeting` request, so a process restart with a new value changes the API line immediately.
- **Build-time path (wasm):** `wasm/main.go` declares the package-level variable `var greeting = "hello world oxzoo-wasm_dev"`, and the install hook builds with `-ldflags "-X 'main.greeting=hello world oxzoo-wasm_$GREETING_TAG'"`, baking the full greeting as one contiguous string into `public/greeting.wasm`. The quoting matters: the double quotes keep the value as a single `-ldflags` argument, and the single quotes keep the spaces intact inside the `-X` value. The wasm module never reads env at runtime, so a new tag requires a redeploy.

**`PORT`** is injected by the platform into the process environment (`environment_file`); `cmd/server/main.go` reads it with `os.Getenv("PORT")` and defaults to `9115`.

**Set `GREETING_TAG` in the ox Environment editor BEFORE the first deploy.** The wasm string is baked during the deploy build step, so changing it later requires a redeploy; the API value updates as soon as the process restarts. `.env.example` documents the variable with a placeholder; real values live in the ox dashboard, never in git.

## Deploy with ox

1. Add the repo in the ox dashboard: paste the clone URL `https://github.com/saurav-codes/oxzoo-wasm`.
2. In the Environment editor, set `GREETING_TAG` (for example `w3-01`).
3. Press **Deploy**. ox runs the three install hooks, starts `./server`, and waits for `http://127.0.0.1:9115/health` to return `ok`.

Expect the app at https://wasm.oxzoo.sorv.dev.

## Expected output

Visiting the domain shows the project heading plus the greeting, computed inside the wasm module:

```
oxzoo-wasm
hello world oxzoo-wasm_w3-01
```

with the note line under it explaining that the string is rendered by the wasm module and baked at build time. `curl https://wasm.oxzoo.sorv.dev/api/greeting` returns exactly:

```
hello world oxzoo-wasm_w3-01
```

as `text/plain`, and `/health` returns `ok`.

## How nginx fits

ox configures nginx with `spa = true`: it serves `public/` from the current release with `try_files $uri $uri/ /index.html`. Only `index.html`, `app.js`, and `style.css` ship in git; the two build artifacts `public/greeting.wasm` and `public/wasm_exec.js` are produced by the install hooks at deploy time and are gitignored. Only the `[frontend].api_paths` prefixes `/api` and `/health` are proxied to the web process on `127.0.0.1:9115`; everything else is static files.

## Local development

```bash
go build -o server ./cmd/server
GREETING_TAG=localtest GOOS=js GOARCH=wasm go build -ldflags "-X 'main.greeting=hello world oxzoo-wasm_$GREETING_TAG'" -o public/greeting.wasm ./wasm
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" public/wasm_exec.js
GREETING_TAG=localtest PORT=9115 ./server   # serves public/ and the API on 127.0.0.1:9115
```

Then open http://127.0.0.1:9115 — the server serves `public/` directly, no nginx needed locally. Pass env inline per the commands above; never commit a real `.env`.
