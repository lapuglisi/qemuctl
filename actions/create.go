package qemuctl_actions

import (
	"flag"
	"fmt"
	"log"

	helpers "github.com/lapuglisi/qemuctl/helpers"
	qemuctl_qemu "github.com/lapuglisi/qemuctl/qemu"
	runtime "github.com/lapuglisi/qemuctl/runtime"
)

type CreateAction struct {
	configFile string
	doForce    bool
}

func (action *CreateAction) Run(arguments []string) (err error) {
	var flagSet *flag.FlagSet = flag.NewFlagSet("qemuctl create", flag.ExitOnError)

	flagSet.StringVar(&action.configFile, "config", "", "YAML configuration file")
	flagSet.BoolVar(&action.doForce, "force", false, "destroys machine if it already exists")

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
	var qemuPid int

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
		existsMessage := fmt.Sprintf("machine '%s' exists", machine.Name)
		if action.doForce {
			fmt.Printf("\n== '--force' specified: destroing machine '%s'\n", machine.Name)
			if !machine.Destroy() {
				fmt.Println("\033[31merror!\033[0m")
				return fmt.Errorf("could not destroy machine '%s'", machine.Name)
			}

			machine.CreateRuntime()
		} else {
			log.Printf("[qemuctl::create] machine '%s' exists and -force was not specified.", machine.Name)
			fmt.Println("\033[31merror!\033[0m")
			return fmt.Errorf(existsMessage)
		}
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

	/* Update machine status to 'created' */
	{
		machine.QemuPid = 0
		machine.SSHLocalPort = 0
		machine.Status = runtime.MachineStatusCreated
		machine.UpdateData()
	}

	log.Printf("[create] launching qemu")
	qemuPid, err = qemu.Launch()
	if err != nil {
		fmt.Println("\033[33merror!\033[0m")
		return err
	} else {
		log.Printf("[create] got machine pid: %d", qemuPid)

		log.Printf("[create] new machine: QemuPid is %d, SSHLocalPort is %d", qemuPid, configData.SSH.LocalPort)
		machine.QemuPid = qemuPid
		machine.SSHLocalPort = configData.SSH.LocalPort
		machine.Status = runtime.MachineStatusRunning
		machine.UpdateData()

		fmt.Println("\033[32mok!\033[0m")
	}

	return nil
}
