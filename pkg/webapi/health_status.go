package webapi

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func (api *apiVigie) addHealthEndpoints(router *mux.Router) {

	router.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]bool{"todo": true})
	})

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {

		_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
		return

		// an example HTTP handler
		if api.vigie.Health() == "Ready" {
			_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
		} else {
			http.Error(w, http.StatusText(503), 503)
		}
	})

}
