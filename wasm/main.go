// oxzoo-wasm wasm module: renders the page greeting in the browser.
// The full string is baked at build time via -ldflags "-X 'main.greeting=...'".
package main

import "syscall/js"

var greeting = "hello world oxzoo-wasm_dev"

func main() {
	js.Global().Get("document").Call("getElementById", "greeting").Set("textContent", greeting)
	<-make(chan struct{})
}
