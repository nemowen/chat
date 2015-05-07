package main

import (
	"flag"
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
	t.templ.Execute(w, r)
}

func main() {
	var addr = flag.String("addr", ":1987", "The addr of the application.")
	flag.Parse()

	r := newRoom()
	http.Handle("/", &templateHandle{fileName: "chat.html"})
	http.Handle("/room", r)
	go r.run()

	log.Println("Starting web server on port", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
