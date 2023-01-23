package qemuctl_actions

import (
	"flag"
	"fmt"
	"log"

	helpers "github.com/lapuglisi/qemuctl/helpers"
	runtime "github.com/lapuglisi/qemuctl/runtime"
)

type DisableAction struct {
	machineName string
	doForce     bool
}

func (action *DisableAction) Run(arguments []string) (err error) {
	var flagSet *flag.FlagSet = flag.NewFlagSet("qemuctl start", flag.ExitOnError)

	flagSet.BoolVar(&action.doForce, "force", false, "destroys machine if it already exists")

	err = flagSet.Parse(arguments)
	if err != nil {
		return err
	}

	action.machineName = arguments[0]

	/* Do proper handling */
	err = action.handleDisable()
	if err != nil {
		return err
	}

	return nil
}

func (action *DisableAction) handleDisable() (err error) {
	var configData *helpers.RuntimeConfiguration = nil
	var configPath string = fmt.Sprintf("%s/%s", runtime.GetSystemConfDir(), runtime.RuntimeConfFileName)
	var machineFound bool = false

	configData, err = helpers.GetRuntimeConfiguration(configPath)
	if err != nil {
		return err
	}

	for _, machine := range configData.Machines {
		if machine.Name == action.machineName {
			log.Printf("[qemuctl::actions::disable] machine '%s' is in '%s'.\n", machine.Name, configPath)
			if machine.Enabled {
				log.Printf("[qemuctl::actions::disable] machine '%s' is enabled. Disabling it.\n", machine.Name)
				machine.Enabled = false
			} else {
				log.Printf("[qemuctl::actions::disable] machine '%s' is already disabled. Skipping.\n", machine.Name)
			}

			machineFound = true
			break
		}
	}

	if !machineFound {
		log.Printf("[qemuctl::actions::enable] machine '%s' was not found in '%s'. Enabling it.\n",
			action.machineName, configPath)
		configData.Machines = append(configData.Machines,
			helpers.RuntimeConfigurationMachine{
				Enabled: false,
				Name:    action.machineName,
			})
	}

	err = helpers.SaveRuntimeConfiguration(configData, configPath)
	if err != nil {
		return err
	}

	return nil
}
