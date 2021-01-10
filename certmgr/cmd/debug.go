package cmd

import (
	"fmt"

	"github.com/KalleDK/go-certmgr/certmgr/certmgr"
	"github.com/spf13/cobra"
)

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
		var seti certmgr.Settings
		err := sf.Unmarshal(&seti)
		fmt.Println(err)
		fmt.Printf("%+v\n", seti)

	},
}

func init() {
	rootCmd.AddCommand(debugCmd)
}
