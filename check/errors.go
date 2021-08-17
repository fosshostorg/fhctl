package check

import (
	"fmt"
	"os"
)

func CheckErr(err error, msg string) {
	if err != nil {
		if err.Error() == "^C" {
			os.Exit(0)
		}
		if msg != "" {
			fmt.Printf("%v: %v\n", msg, err)
			return
		}
		if msg == "" {
			fmt.Printf("Error: %v\n", err)
			return
		}
	}
}
