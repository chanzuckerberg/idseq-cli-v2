package cmd

import (
	"fmt"

	"github.com/chanzuckerberg/czid-cli/pkg/auth0"
	"github.com/spf13/cobra"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with CZID",
	Long: `Log into Chan Zuckerberg ID so you can upload samples.
This will either open a web page or provide you with
a link to a web page if you use the --headless
option. Once you log in on that web page on any
device (not necessarily the one you ran the command on)
you will be authorized to upload samples to your
CZID account.

By default you will remain authenticated for a short
time. If you would like to obtain a secret that
allows you to stay persistently authenticated use the
--persistent option. If you do this a long lived
secret will be added to your configuration file
so please exercise caution when handling this
file. If you suspect your secret has been
comprimised, please reach out to CZID support
at https://chanzuckerberg.zendesk.com/hc/en-us/requests/new.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		headless, err := cmd.Flags().GetBool("headless")
		if err != nil {
			return err
		}
		persistent, err := cmd.Flags().GetBool("persistent")
		if err != nil {
			return err
		}
		err = auth0.DefaultClient.Login(headless, persistent)
		if err != nil {
			return err
		} else {
			fmt.Print("Thanks for logging in! Just a friendly reminder: To not overwhelm CZ ID, please limit your uploads to less than 500 samples per upload, and not more than 1,000 samples per week.\n")
		}

		return nil
	},
}

func init() {
	RootCmd.AddCommand(loginCmd)
	loginCmd.PersistentFlags().Bool("headless", false, "don't open the login form in a browser")
	loginCmd.PersistentFlags().Bool("persistent", false, "remain logged in on this device (see description)")
}
