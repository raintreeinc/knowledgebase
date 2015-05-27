package kbpage

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/raintreeinc/knowledgebase/kb"
	"github.com/raintreeinc/knowledgebase/kbserver"
)

var _ kbserver.System = &System{}

type System struct {
	server *kbserver.Server
	router *mux.Router
}

func New(server *kbserver.Server) *System {
	sys := &System{
		server: server,
		router: mux.NewRouter(),
	}
	sys.init()
	return sys
}

func (sys *System) Info() kbserver.Group {
	return kbserver.Group{
		ID:          "page",
		Name:        "Page",
		Public:      true,
		Description: "Displays page listing and information.",
	}
}

//TODO
func (sys *System) Pages() []kb.PageEntry { return nil }

func (sys *System) init() {
	m := sys.router
	m.HandleFunc("/page:pages", sys.pages).Methods("GET")
	m.HandleFunc("/page:recent-changes", sys.recentChanges).Methods("GET")
	m.HandleFunc("/page:search", sys.search).Methods("GET")
}

func (sys *System) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sys.router.ServeHTTP(w, r)
}

func (sys *System) pages(w http.ResponseWriter, r *http.Request) {
	user, err := sys.server.CurrentUser(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	index := sys.server.IndexByUser(user.ID)

	entries, err := index.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	kbserver.WriteJSON(w, r, &kb.Page{
		Owner: "page",
		Slug:  "page:pages",
		Title: "Pages",
		Story: kb.StoryFromEntries(entries),
	})
}

func (sys *System) search(w http.ResponseWriter, r *http.Request) {
	user, err := sys.server.CurrentUser(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	index := sys.server.IndexByUser(user.ID)

	q := r.URL.Query().Get("q")
	entries, err := index.Search(q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	kbserver.WriteJSON(w, r, &kb.Page{
		Owner: "page",
		Slug:  "page:search",
		Title: "Search \"" + q + "\"",
		Story: kb.StoryFromEntries(entries),
	})
}

func (sys *System) recentChanges(w http.ResponseWriter, r *http.Request) {
	user, err := sys.server.CurrentUser(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	index := sys.server.IndexByUser(user.ID)

	entries, err := index.RecentChanges(30)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	kbserver.WriteJSON(w, r, &kb.Page{
		Owner: "page",
		Slug:  "page:recent-changes",
		Title: "Recent Changes",
		Story: kb.StoryFromEntries(entries),
	})
}