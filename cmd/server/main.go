// oxzoo-wasm API: stdlib-only server behind nginx. GREETING_TAG is read at
// runtime; the wasm module's copy is baked at build time (see ox.toml).
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9115"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/greeting", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintf(w, "hello world oxzoo-wasm_%s", os.Getenv("GREETING_TAG"))
	})
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte("ok"))
	})
	mux.Handle("GET /", http.FileServer(http.Dir("public")))

	if err := http.ListenAndServe("127.0.0.1:"+port, mux); err != nil {
		log.Fatal(err)
	}
}
