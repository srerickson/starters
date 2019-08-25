package main

import (
	"crypto/tls"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/volatiletech/sqlboiler/boil"
	"gopkg.in/yaml.v2"
)

const (
	envVarSecret      = "API_SECRET"
	envVarDatabaseURL = "API_DB_URL"
)

// Config holds main config options for the server
type Config struct {
	Secret       string
	DatabaseURL  string          `yaml:"database"`
	ServerNet    string          `yaml:"net"`
	ServerPort   int             `yaml:"port"`
	ServeTLS     bool            `yaml:"tls"`
	TLSCrt       string          `yaml:"tls-crt"`
	TLSKey       string          `yaml:"tls-key"`
	Prefix       string          `yaml:"api-path-prefix"`
	StaticPrefix string          `yaml:"static-path-prefix"`
	StaticDir    string          `yaml:"static-dir"`
	Auths        []Auth          `yaml:"auths"`
	authMap      map[string]Auth `yaml:"-"`
	Verbose      bool
}

// Auth is an authorization object
type Auth struct {
	ID        string `yaml:"id"`
	KeyDigest string `yaml:"key-digest"`
	Roles     []string
}

var defaults = Config{
	ServerNet:    "127.0.0.1",
	ServerPort:   8080,
	DatabaseURL:  "postgres://localhost",
	TLSCrt:       "server.crt",
	TLSKey:       "server.key",
	Prefix:       "/api",
	StaticDir:    "static",
	StaticPrefix: "/static",
}
var config Config     // global config object
var configPath string // path to config file

func init() {
	flag.StringVar(&configPath, "c", "api.yml", "path to config file (.yml)")
	flag.BoolVar(&config.Verbose, "v", false, "verbose")
}

func loadConfig() error {
	config = defaults
	flag.Parse()
	cfgRaw, err := ioutil.ReadFile(configPath)
	if err != nil {
		return err
	}
	if err = yaml.Unmarshal(cfgRaw, &config); err != nil {
		return err
	}
	// check environment for Secret and Database settings
	if ev := os.Getenv(envVarSecret); ev != "" {
		config.Secret = ev
	}
	if ev := os.Getenv(envVarDatabaseURL); ev != "" {
		config.DatabaseURL = ev
	}
	// check required settings
	if config.Secret == "" {
		return errors.New("secret not set")
	}
	if config.DatabaseURL == "" {
		return errors.New("database URL not set")
	}
	if config.ServeTLS {
		config.ServerPort = 443
	}
	return nil
}

func apiHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`what`))
	})
}

func main() {
	// Check config and open db connection
	if err := loadConfig(); err != nil {
		log.Fatalf("config error: %s", err)
	}
	db, err := sql.Open("postgres", config.DatabaseURL)
	if err != nil {
		log.Fatalf("database error: %s", err)
	}
	boil.SetDB(db)

	// Routing
	root := mux.NewRouter()
	// api routes
	apiRoute := root.PathPrefix(config.Prefix).Subrouter()
	apiRoute.Use(authMiddleware)

	apiRoute.Handle("", apiHandler())
	// static director routes
	root.PathPrefix(config.StaticPrefix).Handler(http.StripPrefix(config.StaticPrefix, http.FileServer(http.Dir(config.StaticDir))))

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

// func gracefullShutdown(server *http.Server, quit <-chan os.Signal, done chan<- bool) {
// 	<-quit
// 	log.Println("Server is shutting down...")

// 	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
// 	defer cancel()

// 	server.SetKeepAlivesEnabled(false)
// 	if err := server.Shutdown(ctx); err != nil {
// 		log.Fatalf("Could not gracefully shutdown the server: %v\n", err)
// 	}
// 	close(done)
// }
