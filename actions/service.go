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
	var stopAction StopAction
	var autoStartDir string = fmt.Sprintf("%s/%s", qemuctl_runtime.GetSystemConfDir(), qemuctl_runtime.RuntimeAutoStartDirName)
	var serviceAction string = "start"

	if len(arguments) > 0 {
		serviceAction = arguments[0]
	}

	switch serviceAction {
	case "start":
		{
			log.Printf("[qemuctl::actions::service::start] checking directory '%s'...\n", autoStartDir)

			dirEntries, _ := os.ReadDir(autoStartDir)
			for _, entry := range dirEntries {
				currentConf := fmt.Sprintf("%s/%s", autoStartDir, entry.Name())
				log.Printf("[qemuctl::actions::service::start] found config file '%s'...\n", currentConf)

				configHandler := qemuctl_helpers.NewConfigHandler(currentConf)
				configData, err := configHandler.ParseConfigFile()
				if err != nil {
					log.Printf("[qemuctl::actions::service::start] error while parsing file '%s': %s\n",
						currentConf, err.Error())
				}

				log.Printf("[qemuctl::actions::service] starting machine '%s'...\n", configData.Machine.MachineName)

				startAction = StartAction{}
				startAction.Run([]string{configData.Machine.MachineName})
			}
		}

	case "stop":
		{
			log.Printf("[qemuctl::actions::service::stop] stopping running machines...\n")
			listAction := ListAction{}
			machines := listAction.getMachines(qemuctl_runtime.MachineStatusRunning)
			for _, machine := range machines {
				stopAction = StopAction{}
				stopAction.Run([]string{machine})
			}

		}
	default:
		{
			log.Printf("[qemuctl::actions::service] unknown service action '%s'.\n", serviceAction)
		}
	}

	return err
}
