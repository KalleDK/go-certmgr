package acmeserver

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/KalleDK/go-certapi/certapi"
)

type SimpleAuth struct {
	Domains map[certapi.APIKey]string
}

func (s SimpleAuth) HasAccess(domain string, t certapi.CertType, key certapi.APIKey) bool {
	apidomain, ok := s.Domains[key]
	return ok && (domain == apidomain)
}

func SimpleAuthFromJson(path string) (s SimpleAuth) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println(err)
	}
	json.Unmarshal(data, &s.Domains)
	return
}
