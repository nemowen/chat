package main

import (
	"fmt"
	"github.com/stretchr/gomniauth"
	"log"
	"net/http"
	"strings"
)

type authHandle struct {
	next http.Handler
}

func (a *authHandle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if _, err := r.Cookie("auth"); err == http.ErrNoCookie {
		// 未验证,将定位到/login页面
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else if err != nil {
		// 其它错误
		panic(err.Error())
	} else {
		// 验证成功, 继续处理下面的handler
		a.next.ServeHTTP(w, r)
	}
}

func MustAuth(h http.Handler) http.Handler {
	return &authHandle{next: h}
}

// format: /auth/{action}/{provider}
func loginHandler(w http.ResponseWriter, r *http.Request) {
	segs := strings.Split(r.URL.Path, "/")
	action := segs[2]
	provider := segs[3]
	switch action {
	case "login":
		provider, err := gomniauth.Provider(provider)
		if err != nil {
			log.Fatalln("Error when trying to get provider", provider,
				"-", err)
		}
		loginUrl, err := provider.GetBeginAuthURL(nil, nil)
		if err != nil {
			log.Fatalln("Error when trying to GetBeginAuthURL for",
				provider, "-", err)
		}
		w.Header.Set("Location", loginUrl)
		w.WriteHeader(http.StatusTemporaryRedirect)
	default:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Auth action %s not supported", action)
	}
}
