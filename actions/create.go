package qemuctl_actions

import (
	"flag"
	"fmt"
	"log"
	"strconv"

	helpers "github.com/lapuglisi/qemuctl/helpers"
	qemuctl_qemu "github.com/lapuglisi/qemuctl/qemu"
	runtime "github.com/lapuglisi/qemuctl/runtime"
)

type CreateAction struct {
	machine    *runtime.Machine
	configFile string
}

func (action *CreateAction) Run(arguments []string) (err error) {
	var flagSet *flag.FlagSet = flag.NewFlagSet("qemuctl start", flag.ExitOnError)

	flagSet.StringVar(&action.configFile, "config", "", "YAML configuration file")

	err = flagSet.Parse(arguments)
	if err != nil {
		return err
	}

	/* Do flags validation */
	if len(action.configFile) == 0 {
		flagSet.Usage()
		return fmt.Errorf("--config is mandatory")
	}

	/* Do proper handling */
	err = action.handleCreate()
	if err != nil {
		return err
	}

	return nil
}

func (action *CreateAction) handleCreate() (err error) {
	var configData *helpers.ConfigurationData = nil
	var qemu *qemuctl_qemu.QemuCommand
	var machine *runtime.Machine

	err = nil

	log.Printf("[create] using config file: %s", action.configFile)

	configHandle := helpers.NewConfigHandler(action.configFile)
	configData, err = configHandle.ParseConfigFile()
	if err != nil {
		return err
	}

	machine = runtime.NewMachine(configData.Machine.MachineName)

	fmt.Printf("[qemuctl] Creating machine '%s' (%s).... ",
		machine.Name, action.configFile)

	/* Check machine status */
	if machine.Exists() {
		fmt.Println("\033[31merror!\033[0m")
		return fmt.Errorf("machine '%s' exists", machine.Name)
	} else {
		machine.CreateRuntime()
	}

	/* First, we update the config file for the machine and use it to create it */
	log.Printf("[create] updating '%s' config file", machine.Name)
	err = machine.UpdateConfigFile(action.configFile)
	if err != nil {
		return err
	}

	log.Printf("[create] using machine config file: '%s'", machine.ConfigFile)
	configHandle = helpers.NewConfigHandler(machine.ConfigFile)
	configData, err = configHandle.ParseConfigFile()
	if err != nil {
		return err
	}

	/* Get QemuCommand instance */
	qemuMonitor := qemuctl_qemu.NewQemuMonitor(machine)
	qemu = qemuctl_qemu.NewQemuCommand(configData, qemuMonitor)

	log.Printf("[create] launching qemu")
	err = qemu.Launch()
	if err != nil {
		machine.QemuPid = 0
		machine.SSHLocalPort = 0
		machine.UpdateStatus(runtime.MachineStatusDegraded)
		return err
	} else {
		procPid := 0
		pidData, err := qemuMonitor.GetPidFileData()
		if err == nil {
			procPid, err = strconv.Atoi(pidData)
			if err == nil {
				log.Printf("[start] got machine pid: %d", procPid)

				log.Printf("[create] new machine: QemuPid is %d, SSHLocalPort is %d", procPid, configData.SSH.LocalPort)
				machine.QemuPid = procPid
				machine.SSHLocalPort = configData.SSH.LocalPort
				machine.UpdateStatus(runtime.MachineStatusStarted)

				fmt.Println("\033[32mok!\033[0m")
			} else {
				log.Printf("[start] could not convert pid string to int %s", err.Error())

				machine.QemuPid = 0
				machine.SSHLocalPort = 0
				machine.UpdateStatus(runtime.MachineStatusStopped)
				fmt.Println("\033[33mstopped!\033[0m")
			}
		} else {
			log.Printf("[start] could not convert pid string to int: %s", err.Error())

			machine.QemuPid = 0
			machine.SSHLocalPort = 0
			machine.UpdateStatus(runtime.MachineStatusStopped)

			fmt.Println("\033[33mstopped!\033[0m")
		}
	}

	return nil
}
