package certmgr

import (
	"github.com/KalleDK/go-certapi/certapi"
	"github.com/google/uuid"
)

type Settings struct {
	ID         uuid.UUID
	CertHome   string
	ApiKey     certapi.APIKey
	ServerCert string
	ServerKey  string
	ServerPort uint
}
