package dispatch

import (
	"html"
	"net/http"
	"strings"

	"github.com/raintreeinc/knowledgebase/kb"
)

var _ kb.Module = &Module{}

type Module struct {
	group  kb.Group
	server *kb.Server
}

func New(group kb.Group, server *kb.Server) *Module {
	mod := &Module{
		group:  group,
		server: server,
	}
	return mod
}

func (mod *Module) Info() kb.Group {
	return mod.group
}

func (mod *Module) Pages() []kb.PageEntry { return []kb.PageEntry{} }

func (mod *Module) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	groupID, titleID, pageID := kb.TokenizeLink3(r.URL.Path)

	if groupID != mod.group.ID {
		http.Error(w, "Invalid owner specified:\nexpected "+string(mod.group.ID)+".",
			http.StatusBadRequest)
		return
	}

	_, index, ok := mod.server.IndexContext(w, r)
	if !ok {
		return
	}

	entries, err := index.ByTitle(titleID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	page := &kb.Page{Slug: pageID}
	if len(entries) > 0 {
		page.Title = entries[0].Title
		page.Story.Append(kb.Tags(entries[0].Tags...))
		page.Story.Append(kb.Paragraph(entries[0].Synopsis))
	} else {
		page.Title = kb.SlugToTitle(titleID)
	}

	if len(entries) == 0 {
		page.Story.Append(kb.Paragraph("No pages."))
	} else {
		page.Story.Append(kb.HTML("<h2>Versions</h2>"))

		prefix := string(mod.group.ID + "/")
		for _, entry := range entries {
			if !strings.HasPrefix(string(entry.Slug), prefix) {
				continue
			}
			groupID, _ := kb.TokenizeLink(string(entry.Slug))
			title := strings.TrimPrefix(string(groupID), prefix)

			page.Story.Append(kb.Entry(
				html.EscapeString(title),
				"",
				entry.Slug,
			))
		}
	}

	page.WriteResponse(w)
}
