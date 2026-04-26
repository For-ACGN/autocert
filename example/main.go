package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/For-ACGN/autocert"
	"github.com/For-ACGN/autocert/acme"
)

const letsEncryptTestURL = "https://acme-staging-v02.api.letsencrypt.org/directory"

var (
	domain string
	ipAddr string
	lAddr  string
	alpn01 bool
	http01 bool
	test   bool
)

func init() {
	flag.StringVar(&domain, "domain", "", "set domain for certificate")
	flag.StringVar(&ipAddr, "ipaddr", "", "set ip address for certificate")
	flag.StringVar(&lAddr, "addr", ":4000", "set http server address")
	flag.BoolVar(&alpn01, "alpn01", false, "force use alpn01 validate method")
	flag.BoolVar(&http01, "http01", false, "force use http01 validate method")
	flag.BoolVar(&test, "test", false, "use test certificate")
	flag.Parse()
}

func main() {
	config := autocert.Config{
		ForceALPN: alpn01,
		ForceHTTP: http01,
	}
	if domain != "" {
		config.Domains = []string{domain}
	}
	if ipAddr != "" {
		config.IPAddrs = []string{ipAddr}
	}
	if test {
		config.Client = &acme.Client{
			DirectoryURL: letsEncryptTestURL,
		}
	}
	config.TLSConfig = &tls.Config{
		NextProtos: []string{"h2", "http/1.1"},
	}

	listener, err := autocert.Listen("tcp", lAddr, &config)
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
