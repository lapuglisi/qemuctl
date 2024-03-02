package main

import (
	"fmt"
	"log"
	"os"

	actions "github.com/lapuglisi/qemuctl/actions"
	runtime "github.com/lapuglisi/qemuctl/runtime"
)

func usage() {
	action := actions.HelpAction{}
	action.Run(nil)
}

func signalHandler(signal os.Signal) {
	log.Printf("[main] received signal %s", signal.String())
}

func main() {
	var err error

	var execArgs []string = os.Args
	var action string

	if os.Getuid() != 0 {
		fmt.Println("need to be root")
		return
	}

	if len(execArgs) < 2 {
		usage()
		os.Exit(1)
	}

	/* Initialize qemuctl */
	err = runtime.SetupRuntimeData()
	if err != nil {
		fmt.Printf("[\033[31merror\033[0m] %s\n", err.Error())
	}

	action = execArgs[1]
	execArgs = execArgs[2:]

	runtime.SetupSignalHandler(signalHandler)

	fmt.Println()
	appAction := actions.GetActionInterface(action)
	err = appAction.Run(execArgs)

	if err != nil {
		fmt.Printf("[\033[31merror\033[0m] %s\n", err.Error())
	}

	os.Exit(0)
}
