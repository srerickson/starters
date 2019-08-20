package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

const staticDir = `static`
const secretEnv = `API_SECRET`

var users = map[string]string{
	`user1`: `secret1`,
}

func apiHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`what`))
	})
}

func main() {
	if os.Getenv(secretEnv) == "" {
		log.Fatalf("environment variable not set: %s", secretEnv)
	}
	root := mux.NewRouter()
	apiRoute := root.PathPrefix(`/api`).Subrouter()
	apiRoute.Use(authMiddleware)
	apiRoute.Handle(``, apiHandler())

	// Static Files
	root.PathPrefix("/static").Handler(http.StripPrefix("/static", http.FileServer(http.Dir(staticDir))))

	// TLS config
	cfg := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
	}
	srv := &http.Server{
		Addr:         ":443",
		Handler:      root,
		TLSConfig:    cfg,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
	}
	log.Fatal(srv.ListenAndServeTLS("localhost.crt", "localhost.key"))

	// log.Fatal(http.ListenAndServe(`127.0.0.1:8080`, root))

}
