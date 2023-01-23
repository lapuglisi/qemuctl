package qemuctl_actions

import (
	"fmt"
	"log"
	"os"
	"path"

	qemuctl_helpers "github.com/lapuglisi/qemuctl/helpers"
)

type CompletionAction struct {
	shellName string
}

func (action *CompletionAction) Run(arguments []string) (err error) {
	var completionString string
	var completion qemuctl_helpers.QemuctlCompletion = qemuctl_helpers.QemuctlCompletion{}

	action.shellName = ""
	if len(arguments) > 0 {
		action.shellName = arguments[0]
	} else {
		action.shellName = path.Base(os.Getenv("SHELL"))
	}

	log.Printf("[CompletionAction] using shell name '%s'\n", action.shellName)
	completionString = completion.GetCompletion(action.shellName)

	log.Printf("[CompletionAction] got completion string: [%s]\n", completionString)

	fmt.Print(completionString)

	return nil
}
