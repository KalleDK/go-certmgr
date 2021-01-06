package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"reflect"

	"github.com/KalleDK/go-certapi/certapi"
	"github.com/KalleDK/go-certmgr/certmgr/certmgr"
	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func stringToUUIDHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(uuid.UUID{}) {
			return data, nil
		}
		uid, err := uuid.Parse(data.(string))

		if err != nil {
			return uuid.UUID{}, fmt.Errorf("failed parsing uuid %w", err)
		}

		return uid, nil
	}
}

func stringToKeyHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(certapi.APIKey{}) {
			return data, nil
		}
		var key certapi.APIKey
		err := key.UnmarshalText([]byte(data.(string)))

		if err != nil {
			return certapi.APIKey{}, fmt.Errorf("failed parsing uuid %w", err)
		}

		return key, nil
	}
}

func loadSettings(setting *certmgr.Settings) error {
	err := viper.Unmarshal(setting,
		viper.DecodeHook(
			mapstructure.ComposeDecodeHookFunc(
				stringToUUIDHookFunc(),
				stringToKeyHookFunc(),
			),
		),
	)
	if err != nil {
		return err
	}
	if setting.ServerPort == 0 {
		if setting.ServerCert != "" {
			setting.ServerPort = 443
		} else {
			setting.ServerPort = 80
		}
	}
	return nil
}

// debugCmd represents the debug command
var debugCmd = &cobra.Command{
	Use:   "debug",
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
		b := &bytes.Buffer{}
		dec := json.NewEncoder(b)
		dec.SetIndent("", "  ")

		if err := dec.Encode(settings); err != nil {
			log.Fatal(err)
		}

		fmt.Print(b.String())

	},
}

func init() {
	rootCmd.AddCommand(debugCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// debugCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// debugCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
