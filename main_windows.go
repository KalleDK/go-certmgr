package main

import (
	"os/user"
)

func (c Config) UID() (uid string, err error) {
	var usr *user.User
	if c.CertUser != "" {
		usr, err = user.Lookup(c.CertUser)
	} else {
		usr, err = user.Current()
	}
	if err != nil {
		return "", err
	}
	return usr.Uid, nil
}

func (c Config) GID() (gid string, err error) {
	var usr *user.User
	if c.CertUser != "" {
		usr, err = user.Lookup(c.CertUser)
	} else {
		usr, err = user.Current()
	}
	if err != nil {
		return "", err
	}
	return usr.Gid, nil
}

func Chown(f string, uid, gid string) error {
	return nil
}
