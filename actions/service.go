package qemuctl_actions

import (
	"fmt"
	"log"

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
	var runtimeConf string = fmt.Sprintf("%s/%s", qemuctl_runtime.GetSystemConfDir(), qemuctl_runtime.RuntimeConfFileName)
	var startAction StartAction

	action.forceStart = false

	log.Printf("[qemuctl::actions::service] checking for enabled machines in '%s'.\n", runtimeConf)
	configData, err := qemuctl_helpers.GetRuntimeConfiguration(runtimeConf)
	if err != nil {
		return err
	}

	for _, machine := range configData.Machines {
		log.Printf("[qemuctl::actions::service] parsing machine '%s'...\n", machine.Name)
		if machine.Enabled {
			log.Printf("[qemuctl::actions::service] service for machine '%s' is enabled. Starting it.\n", machine.Name)

			startAction = StartAction{}
			startAction.Run([]string{machine.Name})
		} else {
			log.Printf("[qemuctl::actions::service] service for machine '%s' is disabled. Skipping.\n", machine.Name)
		}
	}

	return nil
}
