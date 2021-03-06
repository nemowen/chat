package main

import (
	"crypto/md5"
	"fmt"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/common"
	"github.com/stretchr/objx"
	"io"
	"log"
	"net/http"
	"strings"
)

type ChatUser interface {
	UniqueID() string
	AvatarURL() string
}

type chatUser struct {
	common.User
	uniqueID string
}

func (u chatUser) UniqueID() string {
	return u.uniqueID
}

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
			log.Fatalln("在尝试获取提供者时出错", provider,
				"-", err)
		}
		loginUrl, err := provider.GetBeginAuthURL(nil, nil)
		if err != nil {
			log.Fatalln("在获取GetBeginAuthURL时出错",
				provider, "-", err)
		}
		w.Header().Set("Location", loginUrl)
		w.WriteHeader(http.StatusTemporaryRedirect)
	case "callback":
		provider, err := gomniauth.Provider(provider)
		if err != nil {
			log.Fatalln("在尝试获取提供者时出错", provider, "-", err)
		}
		creds, err := provider.CompleteAuth(objx.MustFromURLQuery(r.URL.RawQuery))
		if err != nil {
			log.Fatalln("在尝试完成验证时出错",
				provider, "-", err)
		}
		user, err := provider.GetUser(creds)
		if err != nil {
			log.Fatalln("在尝试从获取用户数据时出错", provider, "-", err)
		}
		chatUser := &chatUser{User: user}
		m := md5.New()
		io.WriteString(m, strings.ToLower(user.Name()))
		chatUser.uniqueID = fmt.Sprintf("%x", m.Sum(nil))
		avatarURL, err := avatars.GetAvatarURL(chatUser)
		if err != nil {
			log.Fatalln("GetAvatarURL时出错", err)
		}
		// 取得名称差使用Base64编码,将其写入到cookie中
		authCookieValue := objx.New(map[string]interface{}{
			"userid":     chatUser.uniqueID,
			"name":       user.Name(),
			"avatar_url": avatarURL,
		}).MustBase64()
		http.SetCookie(w, &http.Cookie{
			Name:  "auth",
			Value: authCookieValue,
			Path:  "/"})
		w.Header()["Location"] = []string{"/chat"}
		w.WriteHeader(http.StatusTemporaryRedirect)

	default:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Auth action %s not supported", action)
	}
}
