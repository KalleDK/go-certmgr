package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/KalleDK/go-certapi/certapi"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const SETTINGSPATH = "/etc/certmgr.conf"

type Settings struct {
	ID         uuid.UUID
	CertHome   string
	Key        certapi.APIKey
	ServerCert string
	ServerKey  string
}

func loadSettings() (*Settings, error) {
	path := SETTINGSPATH
	if len(os.Args) > 1 {
		path = os.Args[1]
	}
	path = filepath.Clean(path)

	fp, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	var settings Settings
	if err := json.NewDecoder(fp).Decode(&settings); err != nil {
		return nil, err
	}

	settings.CertHome = filepath.Clean(settings.CertHome)

	return &settings, nil
}

func serveFavicon(c *gin.Context) {
	const favicon = `<svg
xmlns="http://www.w3.org/2000/svg"
viewBox="0 0 16 16">

<text x="0" y="14">🔒</text>
</svg>`

	c.Header("Content-Type", "image/svg+xml")
	c.Writer.WriteString(favicon)
}

func main() {

	settings, err := loadSettings()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("CertHome: %s\n", settings.CertHome)

	ch := certapi.CertHome{
		Path: settings.CertHome,
		Key:  settings.Key,
	}

	r := gin.Default()

	r.GET("/favicon.ico", serveFavicon)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": fmt.Sprintf("pong: %s", settings.ID),
		})
	})

	certs := r.Group("/cert/:domain", func(c *gin.Context) {
		domain := c.Param("domain")
		c.Set("domain", domain)
	})

	certs.GET("/", func(c *gin.Context) {
		domain := c.GetString("domain")
		info, err := ch.Info(domain)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(200, info)
		}
	})

	certs.GET("/next", func(c *gin.Context) {
		domain := c.GetString("domain")
		info, err := ch.Info(domain)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.String(200, "%d", info.NextRenewTime.Unix())
		}
	})

	certs.GET("/serial", func(c *gin.Context) {
		domain := c.GetString("domain")
		info, err := ch.Info(domain)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.String(200, "%s", info.Serial)
		}
	})

	certs.GET("/key", func(c *gin.Context) {
		domain := c.GetString("domain")
		keystr := c.GetHeader("Authorization")
		if len(keystr) < len("Bearer ") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid api key 0"})
			return
		}
		keystr = keystr[len("Bearer "):]

		keyfile, err := ch.KeyFile(domain, keystr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.File(keyfile)
		}

	})

	certs.GET("/certificate", func(c *gin.Context) {
		domain := c.GetString("domain")
		p, err := ch.Cert(domain)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.File(p)
		}
	})

	certs.GET("/fullchain", func(c *gin.Context) {
		domain := c.GetString("domain")
		p, err := ch.Full(domain)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.File(p)
		}
	})

	if settings.ServerCert == "" {
		r.Run(":80")
	} else {
		r.RunTLS(":443", settings.ServerCert, settings.ServerKey)
	}

}
