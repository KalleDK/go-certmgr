/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/KalleDK/go-certmgr/acmeserver"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	defaultBaseDir  = "./data"
	defaultAuthFile = "./data/auth.json"
	defaultPort     = 8000
	defaultUUID     = "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		key := viper.GetString("serve.ssl.key")
		fmt.Println(key)
		cert := viper.GetString("serve.ssl.cert")

		uid := uuid.MustParse(viper.GetString("serve.uuid"))
		auth := acmeserver.SimpleAuthFromJson(viper.GetString("serve.auth"))
		subfs := os.DirFS(viper.GetString("serve.basedir"))
		handler := acmeserver.NewHandler(subfs, auth, uid)

		addr := fmt.Sprintf(":%d", viper.GetUint32("serve.port"))

		if (key != "") && (cert != "") {
			http.ListenAndServeTLS(addr, cert, key, handler)
		} else {
			http.ListenAndServe(addr, handler)
		}

	},
}

func init() {
	serveCmd.Flags().StringP("uuid", "u", defaultUUID, "UUID")
	serveCmd.Flags().StringP("basedir", "b", defaultBaseDir, "Path")
	serveCmd.Flags().Uint32P("port", "p", defaultPort, "Port")
	serveCmd.Flags().StringP("auth", "a", defaultAuthFile, "Auth")
	serveCmd.Flags().String("ssl.cert", "", "Certificate")
	serveCmd.Flags().String("ssl.key", "", "Certificate Key")

	serveCmd.Flags().VisitAll(func(f *pflag.Flag) {
		viper.BindPFlag("serve."+f.Name, f)
	})

	rootCmd.AddCommand(serveCmd)

}
