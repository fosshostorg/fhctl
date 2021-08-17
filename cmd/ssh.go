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
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"

	"github.com/jcdea/aarch64-client-go"
	"github.com/jcdea/fhctl/check"
	"github.com/jcdea/fhctl/request"
	"github.com/jcdea/fhctl/types"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// sshCmd represents the ssh command
var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "SSH into a vm",

	Run: func(cmd *cobra.Command, args []string) {
		// We already check if user is signed in at GetProjects

		if len(args) != 0 && args[0] != "" {
			err := sshCmdWithAlias(args)
			if err != nil {
				println(fmt.Sprintf("No alias found for\"%v\"\n", args[0]))
			}
		}

		projectResp, err := request.GetProjects()
		check.CheckErr(err, "Failed to retrieve project list")

		var projectNames []string
		for _, item := range projectResp.Projects {
			projectNames = append(projectNames, item.Name)
		}

		check.CheckErr(err, "")
		selectProject := promptui.Select{
			Label: "Select project",
			Items: projectNames,
		}
		selectedProjectIndex, _, err := selectProject.Run()
		check.CheckErr(err, "")

		selectedProject := projectResp.Projects[selectedProjectIndex]

		var VMNames []string
		for _, item := range selectedProject.VMs {
			VMNames = append(VMNames, item.Id)

		}

		check.CheckErr(err, "")
		selectVM := promptui.Select{
			Label: "Select VM",
			Items: VMNames,
		}
		selectedVMIndex, _, err := selectVM.Run()
		check.CheckErr(err, "")

		sshVM(selectedProject.VMs[selectedVMIndex])

	},
}

// Search for alias, then ssh if alias is found.
// if not found: returns not found error
func sshCmdWithAlias(args []string) error {
	var vms []aarch64.VM

	project, err := types.SearchProjectAlias(args[0])
	if err != nil {
		return err
	}

	println(fmt.Sprintf("Using alias %v=%v\n", args[0], project.Id))

	projectResp, err := request.GetProjects()
	check.CheckErr(err, "Failed to retrieve project list")

	for _, item := range projectResp.Projects {
		if item.Id == project.Id {
			vms = item.VMs

		}

	}
	var VMNames []string
	for _, item := range vms {
		VMNames = append(VMNames, item.Id)
	}

	selectVM := promptui.Select{Label: "Select VM", Items: VMNames}
	index, _, err := selectVM.Run()
	check.CheckErr(err, "")

	sshVM(vms[index])
	return nil

}

func init() {
	rootCmd.AddCommand(sshCmd)
}

// SSH into VM
func sshVM(vm aarch64.VM) {
	if ipv6able() {
		cmd := exec.Command("ssh", strings.Split(fmt.Sprintf("root@%v", vm.Address), "/")[0])
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err != nil {
			if err.Error() != "exit status 130" {
				check.CheckErr(err, "")
			}
		}

	} else {
		fmt.Printf("establishing connection to %v through a SSH jump server...\n\n", strings.Split(fmt.Sprintf("root@%v", vm.Address), "/")[0])
		cmd := exec.Command("ssh", "-J", fmt.Sprintf("jump@%v%v.infra.aarch64.com", vm.PoP, vm.Host), strings.Split(fmt.Sprintf("root@%v", vm.Address), "/")[0])
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run() // add error checking

		if err != nil {
			if err.Error() != "exit status 130" {
				check.CheckErr(err, "")
			}
		}

	}

}

// Does the client support ipv6?
func ipv6able() bool {
	_, err := net.Dial("tcp", "2606:4700:4700::1111")
	if err != nil {
		println("ipv6 not available")
		return false
	}
	return true
}
