package kb

import (
	"fmt"
	"net/http"
	"path"
	"strings"
)

type FarmConfig struct {
	Domain    string
	ClientDir string
	Database  string
	Server    []struct {
		Provider string
	}
}

type Farm struct {
	Domain    string
	ClientDir string
	Context   Context
	// fq domain -> server
	Servers map[string]*Server
}

func NewFarm(conf FarmConfig, context Context) (*Farm, error) {
	farm := &Farm{
		Domain:    conf.Domain,
		ClientDir: conf.ClientDir,
		Context:   context,
		Servers:   make(map[string]*Server),
	}

	for _, sconf := range conf.Server {
		sdomain := strings.ToLower(sconf.Provider) + "." + conf.Domain
		server, err := NewServer(sconf.Provider, sdomain, conf.Database)
		if err != nil {
			return nil, fmt.Errorf("error with \"%s\": %s", sconf.Provider, err)
		}
		farm.Servers[sdomain] = server
	}

	return farm, nil
}

func (farm *Farm) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Host == "login."+farm.Domain {
		farm.Context.ServeHTTP(w, r)
		return
	}

	// allow same domain-origin requests
	origin := r.Header.Get("Origin")
	if origin == farm.Domain || strings.HasSuffix(origin, "."+farm.Domain) {
		w.Header().Set("Access-Control-Allow-Methods", "PUT, GET, POST, PATCH, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Origin", origin)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	if !farm.Context.LoggedIn(w, r) {
		farm.Context.RequestCredentials(w, r)
		return
	}

	if r.URL.Path == "/" {
		servefile(w, r, farm.ClientDir, "index.html")
		return
	}

	if strings.HasPrefix(r.URL.Path, "/client/") {
		servefile(w, r, farm.ClientDir, strings.TrimPrefix(r.URL.Path, "/client"))
		return
	}

	if server, ok := farm.Servers[r.Host]; ok {
		server.ServeHTTP(w, r)
		return
	}

	http.NotFound(w, r)
}

func servefile(w http.ResponseWriter, r *http.Request, dir, upath string) {
	if !strings.HasPrefix(upath, "/") {
		upath = "/" + upath
	}
	upath = path.Clean(upath)
	http.ServeFile(w, r, path.Join(dir, upath[1:]))
}