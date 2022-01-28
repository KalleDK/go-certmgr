package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/KalleDK/go-certmgr/certmgr/certmgr"
	"github.com/adrg/xdg"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const binaryName = "certmgr"

var cfgFile string

var (
	defaultConfigDir = filepath.Join(xdg.ConfigHome, binaryName)
	defaultVarDir    = filepath.Join(xdg.StateHome, binaryName)
	defaultConfExt   = "toml"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   binaryName,
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },

}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type Env struct {
	flags *pflag.FlagSet
	v     *viper.Viper
}

func (e Env) Register(name, shorthand string, value interface{}, usage string) {

	switch val := value.(type) {
	case bool:
		if shorthand == "" {
			e.flags.Bool(name, val, usage)
		} else {
			e.flags.BoolP(name, shorthand, val, usage)
		}
	case string:
		if shorthand == "" {
			e.flags.String(name, val, usage)
		} else {
			e.flags.StringP(name, shorthand, val, usage)
		}
	case uint:
		e.flags.UintP(name, shorthand, val, usage)
	case uint16:
		e.flags.Uint16P(name, shorthand, val, usage)
	}
	e.v.SetDefault(name, value)
	e.v.BindPFlag(name, e.flags.Lookup(name))
}

func NewEnv(f *pflag.FlagSet) Env {
	e := Env{
		flags: f,
		v:     viper.New(),
	}
	e.v.SetTypeByDefaultValue(true)
	e.v.SetEnvPrefix(binaryName)
	e.v.AutomaticEnv()
	return e
}

func (e Env) Unmarshal(v interface{}, funcs ...interface{}) error {
	if len(funcs) == 0 {
		return e.v.Unmarshal(v)
	}

	fcs := []mapstructure.DecodeHookFunc{}
	for _, f := range funcs {
		fcs = append(fcs, f.(mapstructure.DecodeHookFunc))
	}

	return e.v.Unmarshal(v, viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(fcs...)))
}

var sf certmgr.SettingsFlags
var ef Env

func init() {
	if runtime.GOOS == "windows" {
		defaultVarDir = filepath.Join(defaultVarDir, "certs")
	}
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file")

	ef = NewEnv(rootCmd.PersistentFlags())
	sf = certmgr.NewSettingsFlags(ef, defaultVarDir)

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		ef.v.SetConfigFile(cfgFile)
	} else {
		ef.v.SetConfigName(binaryName)

		// Find home directory.
		ef.v.AddConfigPath(defaultConfigDir)
		for _, p := range xdg.ConfigDirs {
			ef.v.AddConfigPath(p)
		}
	}

	if err := ef.v.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", ef.v.ConfigFileUsed())
	}

}
