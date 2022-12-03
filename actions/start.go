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
	machineName string
	configFile  string
	qemuBinary  string
}

func (action *StartAction) Run(arguments []string) (err error) {
	/* Check for machine name */
	if len(arguments) < 1 {
		return fmt.Errorf("machine name is mandatory")
	}
	action.machineName = arguments[0]

	fmt.Printf("[start] starting machine '%s'... ", action.machineName)

	/* Do proper handling */
	err = action.handleStart()
	if err != nil {
		fmt.Println("\033[33;1merror!\033[0m")
		return err
	}

	fmt.Println("\033[32;1mok!\033[0m")
	return nil
}

func (action *StartAction) handleStart() (err error) {
	var machine *runtime.Machine

	log.Printf("[start] starting machine '%s'", action.machineName)
	machine = runtime.NewMachine(action.machineName)

	if !machine.Exists() {
		return fmt.Errorf("machine '%s' dos not exist", action.machineName)
	}

	if machine.IsStarted() {
		return fmt.Errorf("[start] machine '%s' is already started", action.machineName)
	}

	if machine.IsDegraded() {
		return fmt.Errorf("[start] cannot start a degraded machine")
	}

	/* in this release, starting a machine means creating it again */
	log.Printf("[start] relaunching machine '%s' (%s)", machine.Name, machine.ConfigFile)

	log.Printf("[start] parsing config file '%s'", machine.ConfigFile)
	configHandle := helpers.NewConfigHandler(machine.ConfigFile)
	configData, err := configHandle.ParseConfigFile()
	if err != nil {
		return err
	}

	log.Printf("[start] creating qemuMonitor instance")
	qemuMonitor := qemuctl_qemu.NewQemuMonitor(machine)

	log.Printf("[start] launching qemu command")
	qemu := qemuctl_qemu.NewQemuCommand(configData, qemuMonitor)

	err = qemu.Launch()
	if err == nil {
		procPid := 0
		pidString, err := qemuMonitor.GetPidFileData()
		if err != nil {
			log.Printf("[start] could not get process pid: %s", err.Error())
		} else {
			procPid, err = strconv.Atoi(pidString)
			if err != nil {
				log.Printf("[start] could not convert pid string to int %s", err.Error())
			} else {
				log.Printf("[start] got machine pid: %d", procPid)
			}
		}
		machine.QemuPid = procPid
		machine.SSHLocalPort = configData.SSH.LocalPort
		machine.UpdateStatus(runtime.MachineStatusStarted)
	} else {
		machine.UpdateStatus(runtime.MachineStatusDegraded)
	}

	return err
}
