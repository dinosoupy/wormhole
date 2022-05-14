/*
Copyright © 2022 Anish Basu

*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/dinosoupy/wormhole/pkg/session/receiver"
)

// receiveCmd represents the receive command
var receiveCmd = &cobra.Command{
	Use:   "receive",
	Short: "Receive files by passing in the name you want to save the received file as argument",
	Long: `Receive command is used to receive files that have been sent to the device`,
	RunE: func(cmd *cobra.Command, args []string) error {
		output := args[0]
		if output == "" {
			return fmt.Errorf("output parameter missing")
		}
		f, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer f.Close()

		conf := receiver.Config{
			Stream: f,
		}

		session := receiver.Receiver(conf)
		return session.Start()
	},
}

func init() {
	rootCmd.AddCommand(receiveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// receiveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// receiveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
