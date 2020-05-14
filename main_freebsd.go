package main

import (
	"os"
	"os/user"
	"strconv"
)

func (c Config) UID() (uid int, err error) {
	var usr *user.User
	if c.CertUser != "" {
		usr, err = user.Lookup(c.CertUser)
	} else {
		usr, err = user.Current()
	}
	if err != nil {
		return 0, err
	}
	uid, err = strconv.Atoi(usr.Uid)
	if err != nil {
		return 0, err
	}
	return uid, nil
}

func (c Config) GID() (gid int, err error) {
	var usr *user.User
	if c.CertUser != "" {
		usr, err = user.Lookup(c.CertUser)
	} else {
		usr, err = user.Current()
	}
	if err != nil {
		return 0, err
	}
	gid, err = strconv.Atoi(usr.Gid)
	if err != nil {
		return 0, err
	}
	return gid, nil
}

func Chown(f string, uid, gid int) error {
	return os.Lchown(f, uid, gid)
}
