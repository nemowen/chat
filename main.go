package main

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sync"
)

type templateHandle struct {
	once     sync.Once
	fileName string
	templ    *template.Template
}

func (t *templateHandle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.fileName)))
	})
	t.templ.Execute(w, nil)
}

func main() {
	http.Handle("/", &templateHandle{fileName: "chat.html"})
	if err := http.ListenAndServe(":1987", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
