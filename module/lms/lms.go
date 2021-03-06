package lms

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gorilla/mux"
	"github.com/raintreeinc/knowledgebase/kb"
)

var _ kb.Module = &Module{}

type Module struct {
	server *kb.Server
	router *mux.Router
}

// New LMS module that acts as a limited LRS
func New(server *kb.Server) *Module {
	mod := &Module{
		server: server,
		router: mux.NewRouter(),
	}
	mod.init()
	return mod
}

// Info
func (mod *Module) Info() kb.Group {
	return kb.Group{
		ID:          "lms",
		Name:        "LMS",
		Public:      true,
		Description: "Learning managament system",
	}
}

func (mod *Module) init() {
	// create temp folder for uploads
	path, _ := os.Getwd()
	_ = os.Mkdir(filepath.FromSlash(path+"/temp/"), 666)
	mod.createUser()

	mod.router.HandleFunc("/lms=lesson", mod.handler).Methods("GET")
	mod.router.HandleFunc("/lms=/uploadContent/", mod.getLessonList).Methods("GET")  // list all existing lessons
	mod.router.HandleFunc("/lms=/uploadContent/", mod.uploadContent).Methods("POST") // create new lesson
	mod.router.HandleFunc("/lms=/uploadVideo/", mod.uploadVideo).Methods("POST")
	mod.router.HandleFunc("/lms=/uploadVideo/", mod.getSignedVideoLink).Methods("GET")
	mod.router.HandleFunc("/lms=/deleteVideo/", mod.deleteVideo).Methods("POST")
}

type lessonData struct {
	LessonID string
	URI      string
}

// todo
// Handle HTTP request to either static file server or REST server (URL start with "api/")
func (mod *Module) handler(w http.ResponseWriter, r *http.Request) {
	// todo: full heigh, need to wrap to div?
	const lessonTemplate = `
<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>-</title>
	</head>
	<body>
		<iframe 
			src="{{.URI}}" 
			width="100%" 
			height="670px" 
			frameborder="0" 
			allowfullscreen="true"
			referrerpolicy="same-origin">
		</iframe >	
	</body>
</html>`
	if strings.HasPrefix(r.URL.RawQuery, "id=") {
		// todo: validate empty & existence, extract to func
		lessonID := strings.Replace(r.URL.RawQuery, "id=", "", 1)
		bucket := getEnvWithDefault("AWS_KB_BUCKET", "rt-knowledge-base-dev")
		uri := "https://" + bucket + ".s3.amazonaws.com/H5P/lessons/" + lessonID + "/template.html"

		w.Header().Set("Content-Type", "application/json") //MIME to application/json
		w.WriteHeader(http.StatusOK)                       //status code 200, OK
		//w.Write([]byte(lessonID))                          //body text

		lesson := lessonData{
			LessonID: lessonID,
			URI:      uri,
		}

		t, err := template.New("webpage").Parse(lessonTemplate)
		check(err)

		err = t.Execute(w, lesson)
		check(err)

		// todo: send back iframe;
		// start by sending back just 1 existing page from DB
	} else {
		// define your static file directory
		staticFilePath := "./client/LMS.html"
		http.ServeFile(w, r, staticFilePath)
	}
}

//  Create default user for LMS uploads
func (mod *Module) createUser() {
	name := "lmsuser"
	_, err := mod.server.Database.Context("admin").Users().ByID(kb.Slugify(name))

	if err == kb.ErrUserNotExist {
		user := kb.User{
			AuthID:       name,
			AuthProvider: "guest",
			ID:           kb.Slugify(name),
			Email:        "lmsuser@raintreeinc.com",
			Name:         name,
			MaxAccess:    kb.Reader,
		}

		_ = mod.server.Database.Context("admin").Users().Create(user)
	}
}

func (mod *Module) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mod.router.ServeHTTP(w, r)
}

// Pages
func (mod *Module) Pages() []kb.PageEntry {
	return []kb.PageEntry{{
		Slug:     "lms=lms",
		Title:    "LMS module.",
		Synopsis: "LMS module.",
	}}
}

func (mod *Module) getLessonList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ListLessonsFromBucket(w)
}

func (mod *Module) uploadContent(w http.ResponseWriter, r *http.Request) {
	err, fileNameWithPath := saveFileFromHttpRequestToServer(r)
	if err != nil {
		kb.WriteResult(w, err)
		return
	}

	if uploadError, uploadedFilePath := uploadFileFromServerToS3(fileNameWithPath); uploadError == nil {
		fmt.Fprintf(w, uploadedFilePath)
	} else {
		fmt.Fprintf(w, uploadError.Error())
	}

	_ = os.Remove(fileNameWithPath)
}

// todo: return nil if req. values are empty?
func (mod *Module) uploadVideo(w http.ResponseWriter, r *http.Request) {
	err, fileNameWithPath := saveFileFromHttpRequestToServer(r)
	if err != nil {
		kb.WriteResult(w, err)
		return
	}

	environment := r.FormValue("environment")
	clientID := r.FormValue("clientID")
	guid := r.FormValue("guid")
	if uploadError, uploadedFilePath := uploadVideoFileFromServerToS3(fileNameWithPath, clientID, environment, guid); uploadError == nil {
		fmt.Fprintf(w, uploadedFilePath)
	} else {
		fmt.Fprintf(w, uploadError.Error())
	}

	_ = os.Remove(fileNameWithPath)
}

func (mod *Module) getSignedVideoLink(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, getSignedLink(r.FormValue("key"), "rt-kb-videos"))
}

func (mod *Module) deleteVideo(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, deleteVideoFileFromS3(r.FormValue("key"), "rt-kb-videos"))
}

func check(err error) {
	if err != nil {
		println(err)
	}
}
