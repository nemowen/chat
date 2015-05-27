package main

import (
	"flag"
	"github.com/nemowen/trace"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/facebook"
	"github.com/stretchr/gomniauth/providers/github"
	"github.com/stretchr/gomniauth/providers/google"
	"github.com/stretchr/objx"
	"html/template"
	"log"
	"net/http"
	_ "net/http/pprof"
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
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates",
			t.fileName)))
	})
	data := map[string]interface{}{
		"Host": r.Host,
	}
	if authCookie, err := r.Cookie("auth"); err == nil {
		data["UserData"] = objx.MustFromBase64(authCookie.Value)
	}

	t.templ.Execute(w, data)
}

var avatars Avatar = TryAvatars{FileSystemAvatar{}, AuthAvatar{}, GravatarAvatar{}}

func main() {

	var addr = flag.String("addr", ":1987", "The addr of the application.")
	flag.Parse()

	// set up gomniauth
	gomniauth.SetSecurityKey("23refwrtwt34tgthtyjngnhjgyiyujnbfsdfsd")
	gomniauth.WithProviders(
		facebook.New("key", "secret", "http://localhost:1987/auth/callback/facebook"),
		github.New("223d6897398e1a2bdb2c", "6681159769dc9c480feb796c1b251d0fe39fb6ae",
			"http://localhost:1987/auth/callback/github"),
		google.New("key", "secret", "http://localhost:1987/auth/callback/google"),
	)

	r := newRoom()
	r.tracer = trace.New(os.Stdout) // using our new trace
	http.Handle("/login", &templateHandle{fileName: "login.html"})
	http.HandleFunc("/logout", logout)
	http.Handle("/chat", MustAuth(&templateHandle{fileName: "chat.html"}))
	http.Handle("/upload", &templateHandle{fileName: "upload.html"})
	http.HandleFunc("/uploader", uploadHandler)
	http.HandleFunc("/auth/", loginHandler)
	http.Handle("/room", r)

	http.Handle("/avatars/", http.StripPrefix("/avatars/", http.FileServer(
		http.Dir("./avatars"))))
	go r.run()

	log.Println("Starting web server on port", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   "auth",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	w.Header()["Location"] = []string{"/chat"}
	w.WriteHeader(http.StatusTemporaryRedirect)
}
