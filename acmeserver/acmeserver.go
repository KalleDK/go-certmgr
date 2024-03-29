package acmeserver

import (
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/KalleDK/go-certapi/certapi"
	"github.com/KalleDK/go-certapi/certapi/certserver"
	"github.com/google/uuid"
	"gopkg.in/ini.v1"
	"software.sslmate.com/src/go-pkcs12"
)

type AuthDB interface {
	HasAccess(domain string, itemtype string, key certapi.APIKey) bool
}

func parseIniTime(section *ini.Section, name string) (t time.Time, err error) {
	key, err := section.GetKey(name)
	if err != nil {
		return
	}

	v, err := key.Int64()
	if err != nil {
		return
	}

	return time.Unix(v, 0), nil
}

func parseIniSerial(section *ini.Section, name string) (s string, err error) {
	key, err := section.GetKey(name)
	if err != nil {
		return
	}

	v := key.String()
	if err != nil {
		return
	}

	idx := strings.LastIndex(v, "/")
	return v[idx+1:], nil
}

func fromIni(b []byte) (c certapi.CertInfo, err error) {

	cfg, err := ini.Load(b)
	if err != nil {
		return
	}

	section := cfg.Section("")

	c.StartDate, err = parseIniTime(section, "Le_CertCreateTime")
	if err != nil {
		return
	}

	c.NextRenewTime, err = parseIniTime(section, "Le_NextRenewTime")
	if err != nil {
		return
	}

	c.Serial, err = parseIniSerial(section, "Le_LinkCert")
	if err != nil {
		return
	}

	return c, nil
}

type AcmeBackend struct {
	fs   fs.FS
	auth AuthDB
}

func (b *AcmeBackend) realname(domain string, itemtype string) (string, error) {
	switch itemtype {
	case certapi.Info:
		return fmt.Sprintf("%s.conf", domain), nil
	case certapi.Cert:
		return fmt.Sprintf("%s.cer", domain), nil
	case certapi.CertChain:
		return "fullchain.cer", nil
	case certapi.Key:
		return fmt.Sprintf("%s.key", domain), nil

	}
	return "", errors.New("file not found")
}

func (b *AcmeBackend) GetSerial(domain string, key certapi.APIKey) (serial string, err error) {
	sub, err := fs.Sub(b.fs, domain)
	if err != nil {
		return
	}

	name, err := b.realname(domain, certapi.Info)
	if err != nil {
		return
	}

	data, err := fs.ReadFile(sub, name)
	if err != nil {
		return
	}

	info, err := fromIni(data)
	if err != nil {
		return
	}

	return info.Serial, nil
}

func (b *AcmeBackend) GetItemInfo(domain string, itemtype string, key certapi.APIKey) (certinfo certapi.ItemInfo, err error) {
	if !b.auth.HasAccess(domain, itemtype, key) {
		err = errors.New("invalid API key")
		return
	}

	sub, err := fs.Sub(b.fs, domain)
	if err != nil {
		return
	}

	name, err := b.realname(domain, itemtype)
	if err != nil {
		return
	}

	stat, err := fs.Stat(sub, name)
	if err != nil {
		return
	}

	certinfo.ModTime = stat.ModTime()
	certinfo.Size = stat.Size()

	return certinfo, nil
}

func (b *AcmeBackend) getInfo(domain string) (cert certapi.Item, err error) {
	sub, err := fs.Sub(b.fs, domain)
	if err != nil {
		return
	}

	name, err := b.realname(domain, certapi.Info)
	if err != nil {
		return
	}

	stat, err := fs.Stat(sub, name)
	if err != nil {
		return
	}
	cert.ModTime = stat.ModTime()
	cert.Size = stat.Size()

	data, err := fs.ReadFile(sub, name)
	if err != nil {
		return
	}

	info, err := fromIni(data)
	if err != nil {
		return
	}

	cert.Data, err = json.Marshal(info)
	if err != nil {
		return
	}

	return cert, nil
}

func (b *AcmeBackend) getPEM(domain string, itemtype string) (cert certapi.Item, err error) {
	sub, err := fs.Sub(b.fs, domain)
	if err != nil {
		return
	}

	name, err := b.realname(domain, itemtype)
	if err != nil {
		return
	}

	stat, err := fs.Stat(sub, name)
	if err != nil {
		return
	}
	cert.ModTime = stat.ModTime()
	cert.Size = stat.Size()

	cert.Data, err = fs.ReadFile(sub, name)
	if err != nil {
		return
	}

	return cert, nil
}

func (b *AcmeBackend) getPFX(domain string) (cert certapi.Item, err error) {
	sub, err := fs.Sub(b.fs, domain)
	if err != nil {
		return
	}

	key, err := func() (key crypto.PrivateKey, err error) {
		name, err := b.realname(domain, certapi.Key)
		if err != nil {
			return
		}

		data, err := fs.ReadFile(sub, name)
		if err != nil {
			return
		}

		block, _ := pem.Decode(data)
		if block == nil || block.Type != "RSA PRIVATE KEY" {
			return nil, errors.New("failed to decode PEM block containing public key")
		}

		return x509.ParsePKCS1PrivateKey(block.Bytes)
	}()
	if err != nil {
		return
	}

	pub, err := func() (pub *x509.Certificate, err error) {
		name, err := b.realname(domain, certapi.Cert)
		if err != nil {
			return
		}

		stat, err := fs.Stat(sub, name)
		if err != nil {
			return
		}
		cert.ModTime = stat.ModTime()
		cert.Size = stat.Size()

		data, err := fs.ReadFile(sub, name)
		if err != nil {
			return
		}

		block, _ := pem.Decode(data)
		if block == nil || block.Type != "CERTIFICATE" {
			err = errors.New("failed to decode PEM block containing public key")
			return
		}

		return x509.ParseCertificate(block.Bytes)
	}()
	if err != nil {
		return
	}

	cert.Data, err = pkcs12.Encode(rand.Reader, key, pub, nil, pkcs12.DefaultPassword)
	if err != nil {
		return
	}

	return cert, nil
}

func (b *AcmeBackend) getPFXChain(domain string) (cert certapi.Item, err error) {
	sub, err := fs.Sub(b.fs, domain)
	if err != nil {
		return
	}

	key, err := func() (key crypto.PrivateKey, err error) {
		name, err := b.realname(domain, certapi.Key)
		if err != nil {
			return
		}

		data, err := fs.ReadFile(sub, name)
		if err != nil {
			return
		}

		block, _ := pem.Decode(data)
		if block == nil || block.Type != "RSA PRIVATE KEY" {
			return nil, errors.New("failed to decode PEM block containing public key")
		}

		return x509.ParsePKCS1PrivateKey(block.Bytes)
	}()
	if err != nil {
		return
	}

	pub, cas, err := func() (pub *x509.Certificate, cas []*x509.Certificate, err error) {
		name, err := b.realname(domain, certapi.Cert)
		if err != nil {
			return
		}

		stat, err := fs.Stat(sub, name)
		if err != nil {
			return
		}
		cert.ModTime = stat.ModTime()
		cert.Size = stat.Size()

		data, err := fs.ReadFile(sub, name)
		if err != nil {
			return
		}

		block, rest := pem.Decode(data)
		if block == nil || block.Type != "CERTIFICATE" {
			err = errors.New("failed to decode PEM block containing public key")
			return
		}

		pub, err = x509.ParseCertificate(block.Bytes)
		if err != nil {
			return
		}

		cas = []*x509.Certificate{}
		for len(rest) > 0 {
			block, rest = pem.Decode(data)
			if block == nil || block.Type != "CERTIFICATE" {
				err = errors.New("failed to decode PEM block containing public key")
				return
			}
			var ca *x509.Certificate
			ca, err = x509.ParseCertificate(block.Bytes)
			if err != nil {
				return
			}
			cas = append(cas, ca)
		}

		return pub, cas, nil
	}()
	if err != nil {
		return
	}

	cert.Data, err = pkcs12.Encode(rand.Reader, key, pub, cas, pkcs12.DefaultPassword)
	if err != nil {
		return
	}

	return cert, nil
}

func (b *AcmeBackend) GetItem(domain string, itemtype string, key certapi.APIKey) (cert certapi.Item, err error) {
	if !b.auth.HasAccess(domain, itemtype, key) {
		err = errors.New("invalid API key")
		return
	}

	switch itemtype {
	case certapi.Info:
		return b.getInfo(domain)
	case certapi.PKCS12:
		return b.getPFX(domain)
	case certapi.PKCS12Chain:
		return b.getPFXChain(domain)
	default:
		return b.getPEM(domain, itemtype)
	}
}

func NewHandler(subfs fs.FS, auth AuthDB, uid uuid.UUID) http.Handler {
	return certserver.NewCertHandler(
		uid,
		&AcmeBackend{
			auth: auth,
			fs:   subfs,
		},
	)
}
