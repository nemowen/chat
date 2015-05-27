package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	//"io/ioutil"
	//"path"
)

// 当Avatar不能提供一个URL的时候
var ErrNoAvatarURL = errors.New("chat: Unable to get an avatar URL.")

type Avatar interface {
	GetAvatarURL(c ChatUser) (string, error)
}

type AuthAvatar struct{}

func (_ AuthAvatar) GetAvatarURL(c ChatUser) (string, error) {
	url := c.AvatarURL()
	if len(url) > 0 {
		return url, nil
	}
	return "", ErrNoAvatarURL

}

type GravatarAvatar struct{}

func (_ GravatarAvatar) GetAvatarURL(c ChatUser) (string, error) {
	return "//www.gravatar.com/avatar/" + c.UniqueID(), nil
}

type FileSystemAvatar struct{}

func (_ FileSystemAvatar) GetAvatarURL(c ChatUser) (string, error) {
	if files, err := ioutil.ReadDir("avatars"); err == nil {
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			if match, _ := path.Match(c.UniqueID()+"*", file.Name()); match {
				return "/avatars/" + file.Name(), nil
			}
		}
	}
	return "", ErrNoAvatarURL
}

type TryAvatars []Avatar

func (a TryAvatars) GetAvatarURL(c ChatUser) (string, error) {
	for _, avatar := range a {
		if url, err := avatar.GetAvatarURL(c); err == nil {
			return url, nil
		}
	}
	return "", ErrNoAvatarURL
}
