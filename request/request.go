package request

import (
	"github.com/jcdea/aarch64-client-go"

	"github.com/jcdea/fhctl/check"
	"github.com/jcdea/fhctl/spinner"
	"github.com/spf13/viper"
)

func GetProjects() (response aarch64.ProjectsResponse, err error) {
	check.CheckSignin()
	s, err := spinner.SpinnerWithMsg(spinner.SpinnerMsgs{
		Suffix:     "Retrieving project list",
		SuccessMsg: "Success",
		FailMsg:    "Failed retrieving project list",
	})
	check.CheckErr(err, "")

	s.Start()
	client := aarch64.NewClient(viper.GetString("key"))

	resp, err := client.Projects()
	if err != nil {
		s.StopFail()
		return aarch64.ProjectsResponse{}, err
	}
	s.Stop()
	return resp, nil
}
