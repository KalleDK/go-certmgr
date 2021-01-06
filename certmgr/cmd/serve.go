package cmd

import (
	"bytes"
	"fmt"
	"log"
	"net/http"

	"github.com/KalleDK/go-certapi/certapi"
	"github.com/KalleDK/go-certmgr/certmgr/certmgr"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
)

func serveFavicon(c *gin.Context) {
	const favicon = `<svg
xmlns="http://www.w3.org/2000/svg"
viewBox="0 0 16 16">

<text x="0" y="14">🔒</text>
</svg>`

	c.Header("Content-Type", "image/svg+xml")
	c.Writer.WriteString(favicon)
}

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		var settings certmgr.Settings
		if err := loadSettings(&settings); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("CertHome: %s\n", settings.CertHome)
		fmt.Printf("Port: %d\n", settings.ServerPort)

		ch := certapi.CertHome{
			Path: settings.CertHome,
			Key:  settings.ApiKey,
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

		{
			certs.GET("", func(c *gin.Context) {
				mime := c.NegotiateFormat("application/json", "text/plain")
				if mime == "" {
					mime = "application/json"
				}
				domain := c.GetString("domain")
				info, err := ch.Info(domain)
				switch mime {
				case "text/plain":
					{
						if err != nil {
							c.AbortWithError(http.StatusBadRequest, err)
						} else {
							resp := ini.Empty()
							sec := resp.Section(ini.DefaultSection)
							sec.NewKey("StartDate", fmt.Sprintf("%d", info.StartDate.Unix()))
							sec.NewKey("NextRenewTime", fmt.Sprintf("%d", info.NextRenewTime.Unix()))
							sec.NewKey("Serial", info.Serial)

							buf := &bytes.Buffer{}
							_, err := resp.WriteTo(buf)
							if err != nil {
								c.AbortWithError(http.StatusInternalServerError, err)
								return
							}

							c.String(200, buf.String())
							return

						}
					}
				case "application/json":
					{
						if err != nil {
							c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
						} else {
							c.JSON(200, info)
						}
					}
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
		}

		if settings.ServerCert == "" {
			r.Run(fmt.Sprintf(":%d", settings.ServerPort))
		} else {
			r.RunTLS(fmt.Sprintf(":%d", settings.ServerPort), settings.ServerCert, settings.ServerKey)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
