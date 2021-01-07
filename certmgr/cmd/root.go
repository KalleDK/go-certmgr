package cmd

import (
	"fmt"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "certmgr",
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

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file  (default is $HOME/.certmgr.yaml)")

	rootCmd.PersistentFlags().String("id", "", "UUID for server")
	viper.BindPFlag("id", rootCmd.PersistentFlags().Lookup("id"))

	rootCmd.PersistentFlags().StringP("certhome", "d", "", "Cert home")
	viper.BindPFlag("certhome", rootCmd.PersistentFlags().Lookup("certhome"))

	rootCmd.PersistentFlags().StringP("apikey", "k", "", "API Key")
	viper.BindPFlag("apikey", rootCmd.PersistentFlags().Lookup("apikey"))

	rootCmd.PersistentFlags().UintP("port", "p", 0, "Server Port")
	viper.BindPFlag("serverport", rootCmd.PersistentFlags().Lookup("port"))

	rootCmd.PersistentFlags().String("cert", "", "Server Cert")
	viper.BindPFlag("servercert", rootCmd.PersistentFlags().Lookup("cert"))

	rootCmd.PersistentFlags().String("certkey", "", "Server Key")
	viper.BindPFlag("serverkey", rootCmd.PersistentFlags().Lookup("certkey"))

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".flaf" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".certmgr")
	}

	viper.SetEnvPrefix("certmgr")
	viper.AutomaticEnv() // read in environment variables that match

	fmt.Println(viper.ConfigFileUsed())
	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
