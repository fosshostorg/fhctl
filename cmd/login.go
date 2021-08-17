/*
Copyright © 2021 JcdeA <jcde@jcde.xyz>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"errors"
	"fmt"
	"net/mail"
	"os"

	"github.com/jcdea/aarch64-client-go"
	"github.com/jcdea/fhctl/check"
	"github.com/jcdea/fhctl/spinner"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to Fosshost",

	Run: func(cmd *cobra.Command, args []string) {
		validate := func(email string) error {
			_, err := mail.ParseAddress(email)
			if err != nil {
				return errors.New("invalid email")
			}
			return nil
		}
		if viper.ConfigFileUsed() != "" {
			choose := promptui.Select{
				Label: fmt.Sprintf("Configuration file found at %v", viper.ConfigFileUsed()),
				Items: []string{"use existing configuration", "overwite config file"},
			}
			i, _, err := choose.Run()
			check.CheckErr(err, "")

			switch i {
			case 0:
				println("using existing configuration.")
				os.Exit(0)
			}

		}

		emailPrompt := promptui.Prompt{
			Label:    "Your Email",
			Validate: validate,
		}
		email, err := emailPrompt.Run()
		check.CheckErr(err, "Prompt failed")

		PwPrompt := promptui.Prompt{
			Label: "Password",
			Mask:  '*',
		}
		pw, err := PwPrompt.Run()

		check.CheckErr(err, "Prompt failed")

		s, err := spinner.SpinnerWithMsg(spinner.SpinnerMsgs{Suffix: "Signing in",
			SuccessMsg: "Successfully logged in!\n",
			FailMsg:    "Failed to sign in. Please try again.\n"})
		check.CheckErr(err, "")

		s.Start()

		client := aarch64.NewClient("")

		resp, err := client.Login(email, pw)

		check.CheckErr(err, "")

		if resp.Meta.Success {

			s.Stop()
			configWriteSpinner, err := spinner.SpinnerWithMsg(spinner.SpinnerMsgs{
				Suffix:     "Writing configuration file",
				SuccessMsg: "Successfully saved configuration file!\n"})
			check.CheckErr(err, "")
			configWriteSpinner.Start()

			viper.Set("email", email)
			viper.Set("key", resp.Key)
			createOrWriteConfig(0600)

			configWriteSpinner.Stop()

		} else {
			s.StopFail()
		}

	},
}

func init() {
	rootCmd.AddCommand(loginCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// loginCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// loginCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
