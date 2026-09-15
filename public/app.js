const go = new Go();
WebAssembly.instantiateStreaming(fetch("greeting.wasm"), go.importObject).then((r) => go.run(r.instance));
