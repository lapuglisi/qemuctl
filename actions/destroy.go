package qemuctl_actions

import (
	"fmt"

	runtime "luizpuglisi.com/qemuctl/runtime"
)

type DestroyAction struct {
	machineName string
}

func (action *DestroyAction) Run(arguments []string) (err error) {
	var machine *runtime.Machine

	if len(arguments) < 1 {
		return fmt.Errorf("machine name is mandatory")
	}

	if action.machineName = arguments[0]; len(action.machineName) == 0 {
		return fmt.Errorf("machine name is mandatory")
	}

	machine = runtime.NewMachine(action.machineName)

	if !machine.Exists() {
		return fmt.Errorf("machine %s dos not exist", action.machineName)
	}

	if machine.IsStarted() {
		fmt.Printf("[qemuctl] \033[33mwarning\033[0m: machine '%s' is started, cannot destroy!\n", action.machineName)
		return nil
	} else {
		fmt.Printf("[qemuctl] destroying machine '%s'... ", action.machineName)
		machine.Destroy()
		fmt.Println("\033[32mok!\033[0m")
	}

	return nil
}
