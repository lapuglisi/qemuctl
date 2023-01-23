package qemuctl_actions

import (
	"fmt"
	"log"
	"os"

	qemuctl_helpers "github.com/lapuglisi/qemuctl/helpers"
	qemuctl_runtime "github.com/lapuglisi/qemuctl/runtime"
)

type ServiceAction struct {
	forceStart bool
}

const (
	ServiceActionEnabledFile string = "qemuctl-enabled"
)

func (action *ServiceAction) Run(arguments []string) (err error) {
	// get runtime directory
	var startAction StartAction
	var autoStartDir string = fmt.Sprintf("%s/%s", qemuctl_runtime.GetSystemConfDir(), qemuctl_runtime.RuntimeAutoStartDirName)

	log.Printf("[qemuctl::actions::service] checking directory '%s'...\n", autoStartDir)

	dirEntries, err := os.ReadDir(autoStartDir)
	for _, entry := range dirEntries {
		currentConf := fmt.Sprintf("%s/%s", autoStartDir, entry.Name())
		log.Printf("[qemuctl::actions::service] found config file '%s'...\n", currentConf)

		configHandler := qemuctl_helpers.NewConfigHandler(currentConf)
		configData, err := configHandler.ParseConfigFile()
		if err != nil {
			log.Printf("[qemuctl::actions::service] error while parsing file '%s': %s\n",
				currentConf, err.Error())
		}

		log.Printf("[qemuctl::actions::service] starting machine '%s'...\n", configData.Machine.MachineName)

		startAction = StartAction{}
		startAction.Run([]string{configData.Machine.MachineName})
	}

	return err
}
