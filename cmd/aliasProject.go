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
	"strings"

	"github.com/jcdea/fhctl/check"
	"github.com/jcdea/fhctl/request"
	"github.com/jcdea/fhctl/types"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// aliasProjectCmd represents the aliasProject command
var aliasProjectCmd = &cobra.Command{
	Use:   "project",
	Short: "Alias a project",

	Run: func(cmd *cobra.Command, args []string) {
		projectResp, err := request.GetProjects()

		var projectNames []string
		for _, item := range projectResp.Projects {
			projectNames = append(projectNames, item.Name)
		}
		check.CheckErr(err, "")
		selectProject := promptui.Select{
			Label: "Select project",
			Items: projectNames,
		}
		selectedIndex, selected, err := selectProject.Run()
		check.CheckErr(err, "")

		validate := func(aliasName string) error {

			if strings.Contains(aliasName, ":") {
				return errors.New("invalid name")
			}
			return nil
		}
		aliasPrompt := promptui.Prompt{
			Label:    fmt.Sprintf("Name to alias \"%v\" to", selected),
			Validate: validate,
		}

		aliasName, err := aliasPrompt.Run()
		check.CheckErr(err, "")

		viper.Set("alias."+aliasName, types.AliasItem{
			Id:   projectResp.Projects[selectedIndex].Id,
			Type: types.Project,
		})

		createOrWriteConfig(0600)
		println(promptui.IconGood + " Successfully aliased resource!")

	},
}

func init() {
	aliasCmd.AddCommand(aliasProjectCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// aliasProjectCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// aliasProjectCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
