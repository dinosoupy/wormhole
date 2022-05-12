/*
Copyright © 2022 Anish Basu

*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/dinosoupy/wormhole/pkg/session"
	"github.com/dinosoupy/wormhole/pkg/session/common"
	"github.com/dinosoupy/wormhole/pkg/session/sender"
)

// sendCmd represents the send command
var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send any flie by passing in filename as the argument",
	Long: `The send command is used to send files to a receiver`,
	Run: func(cmd *cobra.Command, args []string) {
		fileToSend := args[0]
		if fileToSend == "" {
			return fmt.Errorf("file parameter missing")
		}
		f, err := os.Open(fileToSend)
		if err != nil {
			return err
		}
		defer f.Close()
		conf := sender.Config{
			Stream: f,
			Configuration: common.Configuration{
				OnCompletion: func() {
				},
			},
		}
		session := sender.New(f)
		return session.Start()
	},
}

func init() {
	rootCmd.AddCommand(sendCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// sendCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// sendCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
