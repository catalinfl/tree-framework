package tree

import (
	"log"
	"net/http"
)

type ServerConfig struct {
	Port                string
	AutomaticMiddleware bool
}

func (r *Mux) StartExecuting(sc ...ServerConfig) {
	var cfg ServerConfig
	if len(sc) > 0 {
		cfg = sc[0]
	}

	if cfg.Port == "" {
		cfg.Port = ":8080"
	}

	log.Println("Starting server on port", cfg.Port)
	r.setMiddlewareAutomatically(cfg.AutomaticMiddleware)
	r.trees = r.buildTrees()

	err := http.ListenAndServe(cfg.Port, r)
	if err != nil {
		log.Printf("[ERROR] Failed to start server. Error %s", err)
	}
}
