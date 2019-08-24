package main

import (
	"crypto/tls"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
	"github.com/volatiletech/sqlboiler/boil"

	"github.com/gorilla/mux"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const configFileName = "api"

var secret, dbURL, addr, tlsCrt, tlsKey, prefix, staticDir, staticPrefix string
var port int
var serveTLS, verbose bool
var db *sql.DB

func apiHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`what`))
	})
}

func config() error {
	// command line args
	pflag.StringVar(&addr, "addr", "127.0.0.1", "server IP address")
	pflag.IntVar(&port, "port", 8080, "server port")
	pflag.BoolVar(&serveTLS, "tls", true, "wheather to run server using TLS")
	pflag.StringVar(&tlsCrt, "tls-crt", "localhost.crt", "path to TLS cert")
	pflag.StringVar(&tlsKey, "tls-key", "localhost.key", "path to TLS key")
	pflag.StringVar(&prefix, "prefix", "/api", "api prefix")
	pflag.StringVar(&staticDir, "static-dir", "static", "path to static directory")
	pflag.StringVar(&staticPrefix, "static-prefix", "/static", "path to static directory")
	pflag.BoolVar(&verbose, "v", false, "verbose")

	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)
	// environment variables
	viper.BindEnv("API_SECRET", "secret") // environment variable
	viper.BindEnv("API_DB_URL", "dbURL")
	// load config file
	viper.SetConfigName(configFileName)
	viper.AddConfigPath(".")                     // path to look for the config file in
	if err := viper.ReadInConfig(); err != nil { // Handle errors reading the config file
		return err
	}
	// check required settings
	secret = viper.GetString("secret")
	if secret == "" {
		return errors.New("secret not set")
	}
	dbURL = viper.GetString("dbURL")
	if dbURL == "" {
		return errors.New("database URL not set")
	}
	return nil
}

func main() {
	// Check config and open db connection
	if err := config(); err != nil {
		log.Fatalf("config error: %s", err)
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("database error: %s", err)
	}
	boil.SetDB(db)

	// Routing
	root := mux.NewRouter()
	// api routes
	apiRoute := root.PathPrefix(prefix).Subrouter()
	apiRoute.Use(authMiddleware)
	apiRoute.Handle("", apiHandler())
	// static director routes
	root.PathPrefix(staticPrefix).Handler(http.StripPrefix(staticPrefix, http.FileServer(http.Dir(staticDir))))

	if verbose {
		fmt.Println(viper.AllSettings())
	}

	// start the server
	if !serveTLS {
		addrPort := fmt.Sprintf("%s:%d", addr, port)
		log.Printf("listening on %s", addrPort)
		log.Fatal(http.ListenAndServe(addrPort, root))
		return
	}
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
		Addr:         fmt.Sprintf("%s:443", addr),
		Handler:      root,
		TLSConfig:    cfg,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
	}
	log.Printf("listening on %s (w/ TLS)", addr)
	log.Fatal(srv.ListenAndServeTLS(tlsCrt, tlsKey))
}
