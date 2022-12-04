package qemuctl_actions

import (
	"fmt"
	"log"
	"strconv"

	helpers "github.com/lapuglisi/qemuctl/helpers"
	qemuctl_qemu "github.com/lapuglisi/qemuctl/qemu"
	runtime "github.com/lapuglisi/qemuctl/runtime"
)

func init() {
}

type StartAction struct {
	machine *runtime.Machine
}

func (action *StartAction) Run(arguments []string) (err error) {
	/* Check for machine name */
	if len(arguments) < 1 {
		return fmt.Errorf("machine name is mandatory")
	}
	machineName := arguments[0]

	fmt.Printf("[start] starting machine '%s'... ", machineName)

	/* Do proper handling */
	err = action.handleStart(machineName)
	if err != nil {
		fmt.Println("\033[33;1merror!\033[0m")
		return err
	}

	fmt.Println("\033[32;1mok!\033[0m")
	return nil
}

func (action *StartAction) handleStart(machineName string) (err error) {
	log.Printf("[start] starting machine '%s'", machineName)
	action.machine = runtime.NewMachine(machineName)

	if !action.machine.Exists() {
		return fmt.Errorf("machine '%s' dos not exist", action.machine.Name)
	}

	if action.machine.IsStarted() {
		return fmt.Errorf("[start] machine '%s' is already started", action.machine.Name)
	}

	if action.machine.IsDegraded() {
		return fmt.Errorf("[start] cannot start a degraded machine")
	}

	/* in this release, starting a machine means creating it again */
	log.Printf("[start] relaunching machine '%s' (%s)", action.machine.Name, action.machine.ConfigFile)

	log.Printf("[start] parsing config file '%s'", action.machine.ConfigFile)
	configHandle := helpers.NewConfigHandler(action.machine.ConfigFile)
	configData, err := configHandle.ParseConfigFile()
	if err != nil {
		return err
	}

	log.Printf("[start] creating qemuMonitor instance")
	qemuMonitor := qemuctl_qemu.NewQemuMonitor(action.machine)

	log.Printf("[start] launching qemu command")
	qemu := qemuctl_qemu.NewQemuCommand(configData, qemuMonitor)

	err = qemu.Launch()
	if err == nil {
		procPid := 0
		pidString, err := qemuMonitor.GetPidFileData()
		if err == nil {
			procPid, err = strconv.Atoi(pidString)
			if err == nil {
				log.Printf("[start] got machine pid: %d", procPid)

				action.machine.QemuPid = procPid
				action.machine.SSHLocalPort = configData.SSH.LocalPort
				action.machine.UpdateStatus(runtime.MachineStatusStarted)

				fmt.Println("\033[32mok!\033[0m")
			} else {
				log.Printf("[start] could not convert pid string to int %s", err.Error())

				action.machine.QemuPid = 0
				action.machine.SSHLocalPort = 0
				action.machine.UpdateStatus(runtime.MachineStatusStopped)

				fmt.Println("\033[33mstopped!\033[0m")
			}
		} else {
			log.Printf("[start] could not get process pid: %s", err.Error())

			action.machine.QemuPid = 0
			action.machine.SSHLocalPort = 0
			action.machine.UpdateStatus(runtime.MachineStatusStopped)

			fmt.Println("\033[33mstopped!\033[0m")
		}
	} else {
		action.machine.UpdateStatus(runtime.MachineStatusDegraded)
	}

	return err
}
