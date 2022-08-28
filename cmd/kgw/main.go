package main

import (
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/digitalcircle-com-br/kgw/cmd/kgw/k8s"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/acme/autocert"
	"gopkg.in/yaml.v3"
)

//go:embed version
var version string

var lastCfg []byte

var mux *http.ServeMux = http.NewServeMux()

func buildMux() error {
	err := yaml.Unmarshal(lastCfg, cfg)
	if err != nil {
		return err
	}
	if cfg.LogLevel != "" {
		lv, err := logrus.ParseLevel(cfg.LogLevel)
		if err != nil {
			lv = logrus.TraceLevel
		}
		logrus.SetLevel(lv)
	}

	mux = k8s.BuildMux()
	go func() {
		for {
			time.Sleep(time.Second * 15)
			mux = k8s.BuildMux()
		}
	}()
	return nil
}

var allowedHeaders = "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization,X-CSRF-Token"

func handleCors(w http.ResponseWriter, r *http.Request) {
	if origin := r.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
		w.Header().Set("Access-Control-Expose-Headers", "Authorization")
		w.Header().Add("Access-Control-Allow-Credentials", "true")
	}
}

var cfg *Config = &Config{}

func initLog() {
	envll := os.Getenv("LOG_LEVEL")
	if envll == "" {
		logrus.SetLevel(logrus.InfoLevel)
		return
	}
	ll, err := logrus.ParseLevel(envll)
	if err != nil {
		logrus.SetLevel(logrus.TraceLevel)
		logrus.Warnf("Log level %s is unknown - falling back to trace", envll)
		return
	}
	logrus.SetLevel(ll)
}

func run() error {

	initLog()

	logrus.Infof("Version: %s", version)

	err := detectConfigOnce()
	if err != nil {
		return err
	}
	go detectConfig()
	var s = http.Server{
		Addr: cfg.Addr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handleCors(w, r)
			mux.ServeHTTP(w, r)
		}),
	}

	mux.HandleFunc("/__test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("It works!!\n\n"))
		r.Write(w)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		r.Write(w)
	})

	prom := http.Server{
		Addr:    ":8081",
		Handler: promhttp.Handler(),
	}

	go func() {
		logrus.Infof("Running Prom at 8081")
		err := prom.ListenAndServe()
		if err != nil {
			logrus.Fatalf("error running prom server: %s", err.Error())
		}
	}()

	switch {

	case cfg.Acme != nil && cfg.Acme.Enabled:
		logrus.Debugf("Going ACME/TLS mode - :443")
		m := &autocert.Manager{
			Cache:  autocert.DirCache("/kgw/ca"),
			Prompt: autocert.AcceptTOS,
			Email:  "caroot@digitalcircle.com.br",
		}

		s.TLSConfig = m.TLSConfig()
		err = s.ListenAndServeTLS("", "")
	case cfg.Secure:
		logrus.Debugf("Going TLS mode - :443")
		_, err = os.Stat("/kgw/ca/cert")
		if err != nil {
			return fmt.Errorf("could not load cert: %s", err.Error())
		}
		_, err = os.Stat("/kgw/ca/key")
		if err != nil {
			return fmt.Errorf("could not load key: %s", err.Error())
		}
		err = s.ListenAndServeTLS("/kgw/ca/cert", "/kgw/ca/key")
	default:
		logrus.Debugf("Going PLAIN mode - :80")
		err = s.ListenAndServe()
	}

	return err
}

func main() {
	err := run()
	if err != nil {
		panic(err)
	}
}
