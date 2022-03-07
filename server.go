package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"net/http/cgi"
	"os"
	"os/signal"
	"path/filepath"
)

func main() {
	var (
		rootFlag  = flag.String("root", "./werc", "root werc.")
		addrFlag  = flag.String("addr", ":8000", "server address.")
		sitesFlag = flag.String("sites", "sites", "sites directory name.")
		root      string
		addr      string
		sites     string
	)
	flag.Parse()
	addr = *addrFlag
	sites = *sitesFlag
	root, err := filepath.Abs(*rootFlag)
	if err != nil {
		log.Fatalln("!!!", err)
		return
	}

	srvStart(root, addr, sites)
}

func srvStart(root, addr, sites string) {
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, os.Kill)
	srv := &http.Server{
		Addr:    addr,
		Handler: newHandlerWerc(root, sites),
	}
	go func() {
		log.Printf("*** Werc at %s listening on %s ***\n", root, addr)
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalln("!!!", err)
		}
	}()
	<-sig
	if err := srv.Shutdown(context.Background()); err != nil {
		log.Println(err)
	}
}

type HandlerWerc struct {
	root  string
	sites string
	cgi   *cgi.Handler
}

func newHandlerWerc(root string, sites string) *HandlerWerc {
	return &HandlerWerc{
		root:  root,
		sites: sites,
		cgi: &cgi.Handler{
			Path: filepath.Join(root, "bin", "werc.rc"),
			Dir:  filepath.Join(root, "bin"),
			InheritEnv: []string{
				"PLAN9",
			},
			PathLocationHandler: nil,
		},
	}
}

func (h *HandlerWerc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if host, _, err := net.SplitHostPort(r.Host); err == nil {
		r.Host = host
	}
	fn := filepath.Join(h.root, h.sites, r.Host, r.URL.Path)
	if fi, err := os.Stat(fn); err == nil && !fi.IsDir() {
		http.ServeFile(w, r, fn)
		return
	}
	h.cgi.ServeHTTP(w, r)
}
