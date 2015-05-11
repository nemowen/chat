package main

import (
	"flag"
	"github.com/nemowen/trace"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/facebook"
	"github.com/stretchr/gomniauth/providers/github"
	"github.com/stretchr/gomniauth/providers/google"
	"html/template"
	"log"
	"net/http"
	"os"
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

	// set up gomniauth
	gomniauth.SetSecurityKey("some long key")
	gomniauth.WithProviders(
		facebook.New("key", "secret",
			"http://localhost:8080/auth/callback/facebook"),
		github.New("223d6897398e1a2bdb2c", "6681159769dc9c480feb796c1b251d0fe39fb6ae",
			"http://localhost:8080/auth/callback/github"),
		google.New("key", "secret",
			"http://localhost:8080/auth/callback/google"),
	)

	r := newRoom()
	r.tracer = trace.New(os.Stdout) // using our new trace

	http.Handle("/chat", MustAuth(&templateHandle{fileName: "chat.html"}))
	//http.Handle("/assets", http.StripPrefix("/assets", http.FileServer(http.Dir("/path/to/assets/"))))
	http.Handle("/login", &templateHandle{fileName: "login.html"})
	http.HandleFunc("/auth/", loginHandler)
	http.Handle("/room", r)
	go r.run()

	log.Println("Starting web server on port", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
