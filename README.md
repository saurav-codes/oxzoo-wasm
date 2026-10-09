# oxzoo-wasm

Deployed with [ox](https://deploywithox.com): deploy a repo to your own server with one command, no Docker. [Docs](https://deploywithox.com/docs) · [Stack guides](https://deploywithox.com/docs/guides)

An [ox](https://deploywithox.com) deploy example: the page's greeting is computed inside a WebAssembly module compiled from Go (`GOOS=js GOARCH=wasm`), backed by a tiny standard library Go server, deployed to your own Ubuntu server. ox builds the native server, cross-compiles the wasm module with the greeting baked in through `-ldflags -X`, copies the toolchain's `wasm_exec.js`, and runs `./server` under systemd behind Caddy. The server serves `public/` and the API itself.

## Stack

| Component | Version | Purpose |
|---|---|---|
| Wasm module | Go 1.24 (`GOOS=js GOARCH=wasm`) | computes the greeting in the browser via `syscall/js` |
| Server | Go standard library `net/http` | `GET /api/greeting`, `GET /health`, and the files in `public/` |
| Frontend runtime | `wasm_exec.js` from the Go toolchain + vanilla JS | instantiates `greeting.wasm` and runs it |

## ox.toml

```toml
# Go API that also serves public/, plus a Go WebAssembly module built with the greeting baked in.

[app]
start  = "./server"
health = "/health"

[build]
commands = [
  "go build -o server ./cmd/server",
  "GOOS=js GOARCH=wasm go build -ldflags \"-X 'main.greeting=hello world oxzoo-wasm_$GREETING_TAG'\" -o public/greeting.wasm ./wasm",
  "cp \"$(go env GOROOT)/lib/wasm/wasm_exec.js\" public/wasm_exec.js",
]
```

ox detects Go 1.24 from `go.mod` and installs it with mise. The three build commands run with bash in the release directory, with your variables in the environment, so `$GREETING_TAG` expands inside the `-ldflags` value. `public/greeting.wasm` and `public/wasm_exec.js` are build outputs and are gitignored.

## Environment flow

- **Run time (API):** `cmd/server/main.go` reads `GREETING_TAG` on each `/api/greeting` request.
- **Build time (wasm):** `wasm/main.go` declares `var greeting = "hello world oxzoo-wasm_dev"`, and the second build command overwrites it with `-X`, baking the full greeting into `public/greeting.wasm`. The double quotes keep the value as one `-ldflags` argument, and the single quotes keep its spaces. Changing the variable with `ox vars set` redeploys, which rebuilds the module.

## Deploy with ox

```sh
curl -fsSL https://deploywithox.com/install.sh | sh
ox login
ox new https://github.com/saurav-codes/oxzoo-wasm
printf 'GREETING_TAG=demo\n' | ox review oxzoo-wasm --from-file - --wait
```

The plan, offline:

```console
$ ox check .
ox check . (manifest: ox.toml)

  app.start                  ./server                                             declared
  app.health                 /health                                              declared
  build.commands[0]          go build -o server ./cmd/server                      declared
  build.commands[1]          GOOS=js GOARCH=wasm go build -ldflags "-X 'main.greeting=hello world oxzoo-wasm_$GREETING_TAG'" -o public/greeting.wasm ./wasm declared
  build.commands[2]          cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" public/wasm_exec.js declared
  tools.go                   1.24                                                 detected:go.mod

  Provided by ox: PORT, HOST, OX_ENV, OX_PROJECT, OX_RELEASE, OX_DATA_DIR, PUBLIC_URL, PUBLIC_HOST
  Set on the dashboard before the first deploy: GREETING_TAG
  hint: [build] commands replaces the detected build step "go build -o .ox/bin/app ./cmd/server" (from go.mod); add it to the list if the release still needs it

Ready to deploy.
```

## Expected output

```
oxzoo-wasm
hello world oxzoo-wasm_<GREETING_TAG>
```

The greeting line is rendered by the wasm module. `GET /api/greeting` returns the same text as `text/plain`, and `/health` returns `ok`.

## Local development

```sh
export GREETING_TAG=localtest
go build -o server ./cmd/server
GOOS=js GOARCH=wasm go build -ldflags "-X 'main.greeting=hello world oxzoo-wasm_$GREETING_TAG'" -o public/greeting.wasm ./wasm
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" public/wasm_exec.js
PORT=9115 ./server
```

Then open http://127.0.0.1:9115.
