package qemuctl_actions

import (
	"flag"
	"fmt"
	"log"
	"os"

	runtime "github.com/lapuglisi/qemuctl/runtime"
)

type EnableAction struct {
	machineName string
	doLink      bool
}

func (action *EnableAction) Run(arguments []string) (err error) {
	var flagSet *flag.FlagSet = flag.NewFlagSet("qemuctl enable", flag.ExitOnError)

	flagSet.BoolVar(&action.doLink, "link", false, "link config file instead of copying")

	err = flagSet.Parse(arguments)
	if err != nil {
		return err
	}

	action.machineName = flagSet.Args()[0]

	fmt.Printf("[qemuctl] enabling machine '%s'...", action.machineName)

	/* Do proper handling */
	err = action.handleEnable()
	if err != nil {
		fmt.Println(" \033[31merror!\033[0m")
		return err
	}

	fmt.Println(" \033[32mok!\033[0m")
	return nil
}

func (action *EnableAction) handleEnable() (err error) {
	var autoStartDir string = fmt.Sprintf("%s/%s", runtime.GetSystemConfDir(), runtime.RuntimeAutoStartDirName)
	var machinesDir string = runtime.GetMachinesBaseDir()
	var machineFound bool = false

	// iterate through 'machines' directory and find related machine config
	log.Printf("[qemuctl::actions::enable] checking machines in directory '%s'...\n", machinesDir)

	machines, err := os.ReadDir(machinesDir)
	if err != nil {
		return err
	}

	for _, machine := range machines {
		log.Printf("[qemuctl::actions::enable] handling entry '%s'.\n", machine.Name())

		if machine.IsDir() {
			log.Printf("[qemuctl::actions::enable] entry is a directory. Machine name is '%s'...\n", machine.Name())

			currentMachine := runtime.NewMachine(machine.Name())
			if currentMachine.Name == action.machineName {
				log.Printf("[qemuctl::actions::enable] macthed machine with dir '%s'...\n", machine.Name())

				currentConf := currentMachine.ConfigFile
				log.Printf("[qemuctl::actions::enable] enabling machine conf '%s'...\n", currentConf)

				targetConf := fmt.Sprintf("%s/%s.conf", autoStartDir, currentMachine.Name)

				if action.doLink {
					log.Printf("[qemuctl::actions::enable] linking '%s' to '%s'...",
						currentMachine.ConfigFile, targetConf)
				} else {
					log.Printf("[qemuctl::actions::enable] copying '%s' to '%s'...",
						currentMachine.ConfigFile, targetConf)

					err = runtime.CopyFile(currentConf, targetConf)
				}
				machineFound = true
				break
			}
		} else {
			log.Printf("[qemuctl::actions::enable] current entry '%s' is not a directory. Skipping.\n", machine.Name())
		}
	}

	if !machineFound {
		err = fmt.Errorf("machine '%s' not found", action.machineName)
	}

	return err
}
