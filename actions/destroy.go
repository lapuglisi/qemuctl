package qemuctl_actions

import (
	"flag"
	"fmt"

	runtime "github.com/lapuglisi/qemuctl/runtime"
)

type DestroyAction struct {
	machineName  string
	forceDestroy bool
}

func (action *DestroyAction) Run(arguments []string) (err error) {
	var machine *runtime.Machine
	var flagSet *flag.FlagSet = flag.NewFlagSet("qemuctl destroy", flag.ExitOnError)

	flagSet.BoolVar(&action.forceDestroy, "force", false, "destroys machine even if is started")

	err = flagSet.Parse(arguments)
	if err != nil {
		return err
	}

	action.machineName = flagSet.Args()[0]

	if len(action.machineName) == 0 {
		return fmt.Errorf("machine name is mandatory")
	}

	machine = runtime.NewMachine(action.machineName)

	if !machine.Exists() {
		return fmt.Errorf("machine %s dos not exist", action.machineName)
	}

	if machine.IsRunning() || machine.IsStarted() {
		if action.forceDestroy {
			fmt.Printf("[qemuctl] \033[33mwarning\033[0m: force destroying machine '%s'\n", action.machineName)

			killAction := KillAction{}
			killAction.Run([]string{machine.Name})
		} else {
			fmt.Printf("[qemuctl] \033[33mwarning\033[0m: machine '%s' is started, cannot destroy!\n", action.machineName)
			return nil
		}
	}

	fmt.Printf("[qemuctl] destroying machine '%s'... ", action.machineName)
	machine.Destroy()
	fmt.Println("\033[32mok!\033[0m")

	return nil
}
