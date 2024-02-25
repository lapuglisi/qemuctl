package qemuctl_actions

import (
	"fmt"
	"strings"
)

type HelpAction struct {
}

func (action *HelpAction) Run(arguments []string) (err error) {
	actions := GetAllActionStrings()

	fmt.Println()
	fmt.Printf("usage:\n")
	fmt.Printf("  qemuctl {%s}", strings.Join(actions, " | "))
	fmt.Println()
	return nil
}
