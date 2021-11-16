package utils

import (
	"errors"
	"strings"
)

func MustOrgRepo(in string) (string, string) {
	o, r, err := OrgRepo(in)
	if err != nil {
		panic(err)
	}
	return o, r
}

func OrgRepo(in string) (string, string, error) {
	s := strings.Split(in, "/")
	if len(s) != 2 || s[0] == "" || s[1] == "" {
		return "", "", errors.New("invalid org/repo format")
	}
	return s[0], s[1], nil
}
