package main

import (
	_ "embed"
	"io/fs"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/acme/autocert"
	"gopkg.in/yaml.v3"
)

//go:embed version
var version string

var lastCfg []byte

var mux *http.ServeMux = http.NewServeMux()

func newForwader(cr ConfigRoute) http.HandlerFunc {
	url, err := url.Parse(cr.Target)
	if err != nil {
		logrus.Warnf("error setting up fwd: %s", err.Error())
	}
	ret := httputil.NewSingleHostReverseProxy(url)
	if cr.StripPath {
		return func(w http.ResponseWriter, r *http.Request) {
			http.StripPrefix(cr.Path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ret.ServeHTTP(w, r)
			})).ServeHTTP(w, r)

		}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ret.ServeHTTP(w, r)
	}

	// src := r.Path
	// dst := r.Target
	// strip := r.StripPath

	// return func(w http.ResponseWriter, r *http.Request) {
	// 	start := time.Now()
	// 	defer func() {
	// 		dur := time.Since(start)
	// 		log.Printf("Total Processing time: %s", dur.String())
	// 	}()
	// 	if src == "" {
	// 		src = "/"
	// 	}
	// 	host := strings.Split(dst, "://")[1]
	// 	r.Header.Add("x-proxied-by-lb", "true")
	// 	r.Host = host
	// 	r.RequestURI = ""

	// 	// if src != "/" {
	// 	npath := r.URL.Path
	// 	if strip {
	// 		npath = strings.Replace(r.URL.Path, src, "/", 1)
	// 	}
	// 	npath = strings.ReplaceAll(npath, "//", "/")

	// 	strings.TrimSuffix(dst, "/")

	// 	strings.TrimPrefix(npath, "/")

	// 	nurlstr := dst + "/" + npath

	// 	if r.URL.RawQuery != "" {
	// 		nurlstr = nurlstr + "?" + r.URL.RawQuery
	// 	}

	// 	log.Printf("url: %s", nurlstr)

	// 	nurl, err := url.Parse(nurlstr)

	// 	if err != nil {
	// 		http.Error(w, err.Error(), 500)
	// 		log.Printf("error: %s", err.Error())
	// 		return
	// 	}

	// 	log.Printf("Redirecting: [%s] %s=>%s", src, r.URL.String(), nurl.String())

	// 	r.URL = nurl

	// 	res, err := http.DefaultClient.Do(r)
	// 	if err != nil {
	// 		http.Error(w, err.Error(), 500)
	// 		log.Printf("error: %s", err.Error())
	// 		return
	// 	}
	// 	if res.StatusCode != 200 {
	// 		log.Printf("Error calling %s: %s", r.URL.String(), res.Status)
	// 	}
	// 	for k, v := range res.Header {
	// 		for _, vv := range v {
	// 			w.Header().Add(k, vv)
	// 		}
	// 	}
	// 	w.WriteHeader(res.StatusCode)
	// 	defer res.Body.Close()
	// 	defer r.Body.Close()
	// 	io.Copy(w, res.Body)
	// }
}

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
	os.WriteFile("/kgw/ca/cert", []byte(cfg.Cert), 0600)
	os.WriteFile("/kgw/ca/key", []byte(cfg.Key), 0600)
	nmux := http.NewServeMux()
	for _, v := range cfg.Routes {
		func(ar ConfigRoute) {
			nmux.HandleFunc(v.Path, newForwader(ar))
		}(v)
	}
	mux = nmux
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

func dumpMounts() {
	logrus.Debug("Mounted /kgw files:")
	filepath.WalkDir("/kgw", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		logrus.Debugf(" - %s", path)
		return nil
	})
}

func run() error {

	initLog()
	dumpMounts()

	logrus.Infof("Version: %s", version)

	err := detectConfigOnce()
	if err != nil {
		return err
	}

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
		http.NotFound(w, r)
	})

	go detectConfig()

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
