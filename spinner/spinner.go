package spinner

import (
	"time"

	"github.com/theckman/yacspin"
)

func SpinnerWithMsg(m SpinnerMsgs) (*yacspin.Spinner, error) {
	var spinnerCfg = yacspin.Config{
		Frequency:         100 * time.Millisecond,
		CharSet:           yacspin.CharSets[11],
		Suffix:            " " + m.Suffix,
		SuffixAutoColon:   true,
		StopCharacter:     "✔",
		StopFailCharacter: "✗",
		StopFailMessage:   m.FailMsg,
		StopMessage:       m.SuccessMsg,
		StopFailColors:    []string{"fgRed"},
		StopColors:        []string{"fgGreen"},
	}
	return yacspin.New(spinnerCfg)
}

type SpinnerMsgs struct {
	Suffix string

	SuccessMsg string
	FailMsg    string
}
