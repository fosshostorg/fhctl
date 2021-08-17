package check

import (
	"os"

	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

func CheckSignin() {
	if viper.Get("key") == nil {
		println(promptui.IconBad + (" Please log in using `fhctl login` first.\n"))
		os.Exit(0)
	}
}
