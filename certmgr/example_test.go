// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package certmgr_test

import (
	"fmt"
	"log"
	"net/http"

	"github.com/For-ACGN/autocert/certmgr"
)

func ExampleNewListener() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, TLS user! Your config: %+v", r.TLS)
	})
	log.Fatal(http.Serve(certmgr.NewListener("example.com"), mux))
}

func ExampleManager() {
	m := &certmgr.Manager{
		Cache:      certmgr.DirCache("secret-dir"),
		Prompt:     certmgr.AcceptTOS,
		Email:      "example@example.org",
		HostPolicy: certmgr.HostWhitelist("example.org", "www.example.org"),
	}
	s := &http.Server{
		Addr:      ":https",
		TLSConfig: m.TLSConfig(),
	}
	s.ListenAndServeTLS("", "")
}
