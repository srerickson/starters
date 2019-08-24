package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const configFileName = "api"

var secret, addr, tlsCrt, tlsKey, prefix, staticDir, staticPrefix string
var port int
var serveTLS bool

func apiHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`what`))
	})
}

func config() error {
	pflag.StringVar(&addr, "addr", "127.0.0.1", "server IP address")
	pflag.IntVar(&port, "port", 8080, "server port")
	pflag.BoolVar(&serveTLS, "tls", true, "wheather to run server using TLS")
	pflag.StringVar(&tlsCrt, "tls-crt", "localhost.crt", "path to TLS cert")
	pflag.StringVar(&tlsKey, "tls-key", "localhost.key", "path to TLS key")
	pflag.StringVar(&prefix, "prefix", "/api", "api prefix")
	pflag.StringVar(&staticDir, "static-dir", "static", "path to static directory")
	pflag.StringVar(&staticPrefix, "static-prefix", "/static", "path to static directory")
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)
	viper.BindEnv("API_SECRET", "secret") // environment variable
	viper.SetConfigName(configFileName)
	viper.AddConfigPath(".")                     // path to look for the config file in
	if err := viper.ReadInConfig(); err != nil { // Handle errors reading the config file
		return err
	}
	secret = viper.GetString("secret")
	if secret == "" {
		return errors.New("secret seed not set")
	}
	return nil
}

func main() {

	if err := config(); err != nil {
		log.Fatalf("config error: %s", err)
	}

	root := mux.NewRouter()

	// API Routes
	apiRoute := root.PathPrefix(prefix).Subrouter()
	apiRoute.Use(authMiddleware)
	apiRoute.Handle("", apiHandler())

	// Static File Routes
	root.PathPrefix(staticPrefix).Handler(http.StripPrefix(staticPrefix, http.FileServer(http.Dir(staticDir))))

	if serveTLS {
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
		return
	}
	addrPort := fmt.Sprintf("%s:%d", addr, port)
	log.Printf("listening on %s", addrPort)
	log.Fatal(http.ListenAndServe(addrPort, root))
}
