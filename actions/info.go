package qemuctl_actions

import (
	"fmt"

	runtime "github.com/lapuglisi/qemuctl/runtime"
)

type InfoAction struct {
	machineName string
}

func (action *InfoAction) Run(arguments []string) (err error) {
	action.machineName = arguments[0]

	if len(action.machineName) == 0 {
		return fmt.Errorf("machine name is mandatory")
	}

	machine := runtime.NewMachine(action.machineName)
	if !machine.Exists() {
		return fmt.Errorf("machine '%s' does not exist", action.machineName)
	}

	fmt.Println("")
	fmt.Printf("[machine information for '%s']\n", machine.Name)
	fmt.Println("{")
	fmt.Printf("  BiosFile .......... %s\n", machine.BiosFile)
	fmt.Printf("  QEMU PID .......... %d\n", machine.QemuPid)
	fmt.Printf("  SSH Local Port .... %d\n", machine.SSHLocalPort)
	fmt.Printf("  Status ............ %s\n", machine.Status)
	fmt.Printf("  Command Line ...... %s\n", machine.CommandLine)
	fmt.Println("}")
	fmt.Println("")
	return nil
}
