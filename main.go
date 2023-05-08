package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

func main() {
	defaultURL, err := url.Parse("http://localhost:9000")
	if err != nil {
		log.Fatal(fmt.Errorf("error parsing default URL: %v", err))
	}

	nodeURL, err := url.Parse("http://localhost:9100")
	if err != nil {
		log.Fatal(fmt.Errorf("error parsing node URL: %v", err))
	}

	googleURL, err := url.Parse("https://google.com")
	if err != nil {
		log.Fatal(fmt.Errorf("error parsing google's URL: %v", err))
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if q, found := strings.CutPrefix(r.URL.Path, "/google"); found {
			proxy := &httputil.ReverseProxy{
				Rewrite: func(r *httputil.ProxyRequest) {
					r.SetURL(googleURL)
					r.Out.URL.Path = "/" + q
				},
			}
			proxy.ServeHTTP(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/go") {
			w.Write([]byte(fmt.Sprintf("I'm Go!\r\n[%v]\n", r.URL.Path)))
			return
		}
		if strings.HasPrefix(r.URL.Path, "/node") {
			httputil.NewSingleHostReverseProxy(nodeURL).ServeHTTP(w, r)
			return
		}
		httputil.NewSingleHostReverseProxy(defaultURL).ServeHTTP(w, r)
	})
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
