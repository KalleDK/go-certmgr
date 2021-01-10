package certmgr

import (
	"fmt"
	"reflect"

	"github.com/KalleDK/go-certapi/certapi"
	"github.com/google/uuid"
)

type SSLSettings struct {
	Cert string
	Key  string
}

type Settings struct {
	ID       uuid.UUID
	CertHome string
	APIKey   certapi.APIKey
	SSL      SSLSettings
	Port     uint
	Debug    bool
}

type Environment interface {
	Register(name, shorthand string, value interface{}, usage string)
	Unmarshal(v interface{}, funcs ...interface{}) error
}

type SettingsFlags struct {
	Env Environment
}

func NewSettingsFlags(e Environment, defaultVarDir string) SettingsFlags {
	e.Register("id", "i", "", "UUID")
	e.Register("certhome", "d", defaultVarDir, "Cert Home")
	e.Register("port", "p", uint16(0), "Server Port if 0 then 80 or 443 is choosen based on ssl")
	e.Register("apikey", "k", string(""), "APIKey for auth")
	e.Register("ssl.cert", "", string(""), "SSL Cert")
	e.Register("ssl.key", "", string(""), "SSL Key")
	e.Register("debug", "", bool(true), "Debug")
	return SettingsFlags{Env: e}
}

func (sf SettingsFlags) Unmarshal(s *Settings) error {
	err := sf.Env.Unmarshal(s, stringToAPIKey, stringToUUID)
	if err != nil {
		return err
	}

	if s.Port == 0 {
		if s.SSL.Cert == "" || s.SSL.Key == "" {
			s.Port = 80
		} else {
			s.Port = 443
		}
	}

	return nil
}

func stringToAPIKey(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {

	if f.Kind() != reflect.String {
		return data, nil
	}

	if t != reflect.TypeOf(certapi.APIKey{}) {
		return data, nil
	}

	if len(data.(string)) == 0 {
		return certapi.APIKey{}, nil
	}

	var key certapi.APIKey

	if err := key.UnmarshalText([]byte(data.(string))); err != nil {
		return certapi.APIKey{}, fmt.Errorf("failed parsing uuid %w", err)
	}

	return key, nil
}

func stringToUUID(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.String {
		return data, nil
	}

	if t != reflect.TypeOf(uuid.UUID{}) {
		return data, nil
	}

	if len(data.(string)) == 0 {
		return uuid.UUID{}, nil
	}

	uid, err := uuid.Parse(data.(string))

	if err != nil {
		return uuid.UUID{}, fmt.Errorf("failed parsing uuid %w", err)
	}

	return uid, nil
}
