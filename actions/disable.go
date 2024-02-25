package qemuctl_actions

import (
	"flag"
	"fmt"
	"log"
	"os"

	helpers "github.com/lapuglisi/qemuctl/helpers"
	runtime "github.com/lapuglisi/qemuctl/runtime"
)

type DisableAction struct {
	machineName string
	doForce     bool
}

func (action *DisableAction) Run(arguments []string) (err error) {
	var flagSet *flag.FlagSet = flag.NewFlagSet("qemuctl disable", flag.ExitOnError)

	err = flagSet.Parse(arguments)
	if err != nil {
		return err
	}

	action.machineName = arguments[0]

	fmt.Printf("[qemuctl] disabling machine '%s'...", action.machineName)

	/* Do proper handling */
	err = action.handleDisable()
	if err != nil {
		fmt.Println(" \033[31merror!\033[0m")
		return err
	}

	fmt.Println(" \033[32mok!\033[0m")
	return nil
}

func (action *DisableAction) handleDisable() (err error) {
	var autoStartDir string = fmt.Sprintf("%s/%s", runtime.GetSystemConfDir(), runtime.RuntimeAutoStartDirName)
	var machineFound bool = false

	log.Printf("[qemuctl::actions::disable] reading directory '%s'...\n", autoStartDir)

	dirEntries, err := os.ReadDir(autoStartDir)
	for _, entry := range dirEntries {
		if entry.IsDir() {
			log.Printf("[qemuctl::actions::disable] current entry '%s' is a directory. Skipping.\n", entry.Name())
		} else {
			currentConf := fmt.Sprintf("%s/%s", autoStartDir, entry.Name())
			log.Printf("[qemuctl::actions::disable] checking file '%s'.\n", currentConf)

			configHandler := helpers.NewConfigHandler(currentConf)
			configData, err := configHandler.ParseConfigFile()
			if err != nil {
				log.Printf("[qemuctl::actions::disable] error while parsing file '%s': %s.\n", currentConf, err.Error())
				continue
			}

			if configData.Machine.MachineName == action.machineName {
				log.Printf("[qemuctl::actions::disable] found config file '%s' for machine '%s'. Removing it.\n",
					currentConf, action.machineName)
				os.Remove(currentConf)

				machineFound = true
				break
			}
		}
	}

	if !machineFound {
		err = fmt.Errorf("machine '%s' is not enabled", action.machineName)
	}

	return err
}
