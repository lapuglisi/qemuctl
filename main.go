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

	if os.Geteuid() != 0 {
		fmt.Println()
		fmt.Printf("this program must be run as root, but you are user %d.\n", os.Getuid())
		fmt.Println()
		os.Exit(-1)
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
