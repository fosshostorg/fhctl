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
	"bufio"
	"crypto/ed25519"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"log"

	"net"
	"os"
	"strings"
	"time"

	"github.com/jcdea/aarch64-client-go"
	"github.com/jcdea/edkey"
	"github.com/jcdea/fhctl/check"
	"github.com/jcdea/fhctl/request"
	"github.com/jcdea/fhctl/types"
	"github.com/manifoldco/promptui"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
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

// if err != nil {
// 	if err.Error() != "exit status 130" {
// 		check.CheckErr(err, "")
// 	}
// }

// SSH into VM
func sshVM(vm aarch64.VM) error {
	err := initSSHKeys()
	check.CheckErr(err, "")

	if !ipv6able() {
		auth := []ssh.AuthMethod{
			ssh.Password(vm.Password),
		}
		client, err := jumpSSH(auth, vm)
		// Client has not copied ssh id yet

		check.CheckErr(err, "hello ")

		session, err := client.NewSession()
		if err != nil {
			err = copySSHKeys(vm)
			check.CheckErr(err, "")
			session.Close()
			sshVM(vm)

		}
		defer session.Close()
		println("i am starting the shell")

		session.Stderr = os.Stderr
		session.Stdout = os.Stdout
		session.RequestPty("xterm", 40, 40, ssh.TerminalModes{
			ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
			ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
		})

		in, _ := session.StdinPipe()
		// Forward user commands to the remote shell
		if err := session.Shell(); err != nil {

			check.CheckErr(err, "")
			log.Fatalf("failed to start shell: %s", err)
		}

		for {
			reader := bufio.NewReader(os.Stdin)
			str, _ := reader.ReadString('\n')
			fmt.Fprint(in, str)
		}

	}
	return nil

}

func copySSHKeys(vm aarch64.VM) (err error) {
	pubkeyPath := viper.GetString("ssh.pubKeyPath")
	pubKey, err := ioutil.ReadFile(pubkeyPath)
	check.CheckErr(err, "")

	var auth []ssh.AuthMethod

	check.CheckErr(err, "")
	auth = []ssh.AuthMethod{
		ssh.Password(vm.Password),
	}
	sshClient, err := jumpSSH(auth, vm)
	if err != nil {
		return err
	}

	session, err := sshClient.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown host"
	} else {
		hostname = fmt.Sprintf("host %v", hostname)
	}

	err = session.Run(fmt.Sprintf("echo \"%v %v %v\" >> ~/.ssh/authorized_keys", strings.TrimRight(string(pubKey), "\n"), "Generated by fhctl for ", hostname))

	check.CheckErr(err, "")
	return nil
}

func jumpSSH(auth []ssh.AuthMethod, vm aarch64.VM) (client *ssh.Client, err error) {
	jumpConfig := &ssh.ClientConfig{
		User: "jump",
		Auth: auth,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			// TODO: fix this thing to properly check host
			if askIsHostTrusted(hostname, key) {
				return nil
			} else {
				return errors.New("host is not trusted")
			}
		},
		Timeout: time.Second * 10,
	}

	bClient, err := ssh.Dial("tcp", fmt.Sprintf("%v%v.infra.aarch64.com:22", vm.PoP, vm.Host), jumpConfig)
	if err != nil {
		log.Fatal(err)
	}

	addr := strings.Split(vm.Address, "/")[0]

	// Dial a connection to the service host, from the jump server
	conn, err := bClient.Dial("tcp6", fmt.Sprintf("[%v]:22", addr))
	if err != nil {
		log.Fatal(err)
	}

	finalConfig := &ssh.ClientConfig{
		User: "root",
		Auth: auth,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			// TODO: fix this thing to properly check host
			if askIsHostTrusted(hostname, key) {
				return nil
			} else {
				return errors.New("host is not trusted")
			}
		},
		Timeout: time.Second * 30,
	}
	ncc, chans, reqs, err := ssh.NewClientConn(conn, addr, finalConfig)
	if err != nil {
		return nil, err
	}

	return ssh.NewClient(ncc, chans, reqs), nil
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

func initSSHKeys() (err error) {

	privkeyPath := viper.GetString("ssh.privKeyPath")
	pubkeyPath := viper.GetString("ssh.pubKeyPath")

	// _, err = os.Stat(privkeyPath)
	// if !os.IsExist(err) {
	// 	if _, err := os.Create(privkeyPath); err != nil {
	// 		cobra.CheckErr(err)
	// 	}
	// }

	// _, err = os.Stat(pubkeyPath)
	// if !os.IsExist(err) {
	// 	if _, err := os.Create(pubkeyPath); err != nil {
	// 		cobra.CheckErr(err)
	// 	}
	// }

	// Generate a new private/public keypair for OpenSSH
	pubKey, privKey, _ := ed25519.GenerateKey(nil)
	publicKey, _ := ssh.NewPublicKey(pubKey)

	pemKey := &pem.Block{
		Type:  "OPENSSH PRIVATE KEY",
		Bytes: edkey.MarshalED25519PrivateKey(privKey),
	}
	privateKey := pem.EncodeToMemory(pemKey)
	authorizedKey := ssh.MarshalAuthorizedKey(publicKey)

	_ = ioutil.WriteFile(privkeyPath, privateKey, 0600)
	_ = ioutil.WriteFile(pubkeyPath, authorizedKey, 0644)

	return nil
}

func askIsHostTrusted(host string, key ssh.PublicKey) bool {

	// reader := bufio.NewReader(os.Stdin)

	// fmt.Printf("Unknown Host: %s \nFingerprint: %s \n", host, ssh.FingerprintSHA256(key))
	// fmt.Print("Would you like to add it? type yes or no: ")

	// a, err := reader.ReadString('\n')

	// if err != nil {
	// 	log.Fatal(err)
	// }

	// return strings.ToLower(strings.TrimSpace(a)) == "yes"
	return true
}
