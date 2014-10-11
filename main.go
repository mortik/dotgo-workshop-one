package main

import (
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
	"gopkg.in/unrolled/render.v1"
)

type Content struct {
	Token    string
	Markdown template.HTML
}

func FormHandler(rw http.ResponseWriter, req *http.Request, r *render.Render) {
	content := &Content{
		Token: req.URL.Query().Get("token"),
	}

	r.HTML(rw, http.StatusOK, "form", content)
}

func MarkdownHandler(rw http.ResponseWriter, req *http.Request, r *render.Render) {
	markdown := blackfriday.MarkdownCommon([]byte(req.FormValue("body")))
	parsedMarkdown := string(bluemonday.UGCPolicy().SanitizeBytes(markdown))

	content := &Content{
		Token:    req.URL.Query().Get("token"),
		Markdown: template.HTML(parsedMarkdown),
	}

	r.HTML(rw, http.StatusOK, "markdown", content)
}

func AuthHandler(rw http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	if req.URL.Query().Get("token") == "secret" {
		next(rw, req)
	} else {
		http.Error(rw, "Not Authorized", http.StatusUnauthorized)
	}
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := render.New(render.Options{
		Layout:     "base",
		Extensions: []string{".html"},
	})

	n := negroni.New(
		negroni.NewRecovery(),
		negroni.NewLogger(),
		negroni.NewStatic(http.Dir("public")),
		negroni.HandlerFunc(AuthHandler),
	)

	router := mux.NewRouter()

	router.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		FormHandler(rw, req, r)
	})

	router.HandleFunc("/markdown", func(rw http.ResponseWriter, req *http.Request) {
		MarkdownHandler(rw, req, r)
	}).Methods("POST")

	n.UseHandler(router)

	n.Run(":" + port)
}
