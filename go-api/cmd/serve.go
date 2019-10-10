package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	_ "github.com/lib/pq"
	"github.com/srerickson/starters/goapi"
	"github.com/srerickson/starters/goapi/resource"
	"gopkg.in/yaml.v2"
)

var configPath string // path to config file
var verbose bool

func init() {
	flag.StringVar(&configPath, "c", "api.yml", "path to config file (.yml)")
	flag.BoolVar(&verbose, "v", false, "verbose")
}

func apiHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`what`))
	})
}

func main() {
	flag.Parse()

	// Check config and open db connection
	config, err := goapi.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("config error: %s", err)
	}
	config.Verbose = verbose
	db, err := sql.Open("postgres", config.DatabaseURL)
	if err != nil {
		log.Fatalf("database error: %s", err)
	}

	// Routing
	root := chi.NewRouter()

	resource.Init(db, `api.resources`)
	root.Mount("/resources", resource.Handler())

	if config.Verbose {
		rawCfg, err := yaml.Marshal(&config)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("#---- server config ----")
		fmt.Printf(string(rawCfg))
		fmt.Println("#---- /config ----")
	}

	// configure the server
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", config.ServerNet, config.ServerPort),
		Handler: root,
		TLSConfig: &tls.Config{
			MinVersion:               tls.VersionTLS12,
			CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			},
		},
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
	}

	// start the server
	start := func() error {
		if config.ServeTLS {
			return srv.ListenAndServeTLS(config.TLSCrt, config.TLSKey)
		}
		return srv.ListenAndServe()
	}
	log.Printf("listening on %s (TLS=%v)", srv.Addr, config.ServeTLS)
	log.Fatal(start())
}
