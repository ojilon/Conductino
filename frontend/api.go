package main

import (
	"log"
	"net/http"
)

func StartIPCServer(cfg *Config){
	mux := http.NewServeMux()

	// 1. Single File Server for ALL frontend assets under ./web
    // This makes everything under /ui/ and /states/ accessible.
    mux.Handle("/", http.FileServer(http.Dir("./web")))

    // 2. Hand off the JSON API endpoints to the handlers package
    api := handlers.NewBackendClient(cfg.IPC.BackendURL)
    

    log.Printf("[Go IPC] listening on %s", cfg.IPC.FrontendListen)
    if err := http.ListenAndServe(cfg.IPC.FrontendListen, mux); err != nil {
        log.Fatalf("IPC server crashed: %v", err)
    }

}