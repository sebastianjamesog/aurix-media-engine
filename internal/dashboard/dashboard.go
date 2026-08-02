package dashboard

import (
	"embed"
	"net/http"
)

//go:embed dashboard.html
var content embed.FS

// RegisterRoutes registers the `/dashboard` route on the given HTTP Mux.
func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /dashboard", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data, err := content.ReadFile("dashboard.html")
		if err != nil {
			http.Error(w, "Dashboard template missing", http.StatusInternalServerError)
			return
		}
		w.Write(data)
	})
}
