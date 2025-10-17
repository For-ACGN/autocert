package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/For-ACGN/autocert"
)

var (
	domain string
	ipAddr string
	lAddr  string
)

func init() {
	flag.StringVar(&domain, "domain", "", "set domain for certificate")
	flag.StringVar(&ipAddr, "ip", "", "set ip address for certificate")
	flag.StringVar(&lAddr, "l", ":4000", "set http server address")
	flag.Parse()
}

func main() {
	cfg := autocert.Config{}
	if domain != "" {
		cfg.Domains = []string{domain}
	}
	if ipAddr != "" {
		cfg.IPAddrs = []string{ipAddr}
	}

	listener, err := autocert.Listen("tcp", lAddr, &cfg)
	checkError(err)
	fmt.Println("bind listener successfully")

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.RemoteAddr)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Hello ACME!"))
	})
	server := http.Server{
		Handler: mux,
	}
	_ = server.Serve(listener)
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
