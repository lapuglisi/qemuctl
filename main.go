package main

import (
	"fmt"
	"log"
	"os"

	actions "github.com/lapuglisi/qemuctl/actions"
	runtime "github.com/lapuglisi/qemuctl/runtime"
)

func usage() {
	fmt.Println()
	fmt.Println("usage:")
	fmt.Println("    qemuctl {start|stop|seila} OPTIONS")
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
	switch action {
	case "create":
		{
			action := actions.CreateAction{}
			err = action.Run(execArgs)
			break
		}
	case "destroy":
		{
			action := actions.DestroyAction{}
			err = action.Run(execArgs)
			break
		}
	case "kill":
		{
			action := actions.KillAction{}
			err = action.Run(execArgs)
			break
		}
	case "start":
		{
			action := actions.StartAction{}
			err = action.Run(execArgs)
			break
		}
	case "stop":
		{
			action := actions.StopAction{}
			err = action.Run(execArgs)
			break
		}
	case "status":
		{
			action := actions.StatusAction{}
			err = action.Run(execArgs)
			break
		}
	case "edit":
		{
			action := actions.EditAction{}
			err = action.Run(execArgs)
		}
	case "list":
		{
			action := actions.ListAction{}
			err = action.Run(execArgs)
		}
	case "service":
		{
			action := actions.ServiceAction{}
			err = action.Run(execArgs)
		}

	case "enable":
		{
			action := actions.EnableAction{}
			err = action.Run(execArgs)
		}
	case "disable":
		{
			action := actions.DisableAction{}
			err = action.Run(execArgs)
		}
	default:
		{
			fmt.Printf("[error] Unknown action '%s'\n", action)
		}
	}

	if err != nil {
		fmt.Printf("[\033[31merror\033[0m] %s\n", err.Error())
	}

	os.Exit(0)
}
