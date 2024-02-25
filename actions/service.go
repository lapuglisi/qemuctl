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
	var createAction CreateAction
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

				log.Printf("[qemuctl::actions::service::start] creating machine from config '%s'...\n", currentConf)

				createAction = CreateAction{}
				createAction.Run([]string{"-config", currentConf})
			}
		}

	case "stop":
		{
			log.Printf("[qemuctl::actions::service::stop] stopping running machines...\n")

			dirEntries, _ := os.ReadDir(autoStartDir)
			for _, entry := range dirEntries {
				currentConf := fmt.Sprintf("%s/%s", autoStartDir, entry.Name())

				log.Printf("[qemuctl::actions::service::stop] found config file '%s'...\n", currentConf)
				log.Printf("[qemuctl::actions::service::stop] destroying machine from config '%s'...\n", currentConf)

				configData := qemuctl_helpers.NewConfigHandler(currentConf)
				data, err := configData.ParseConfigFile()
				if err != nil {
					log.Printf("[qemuctl::actions::service::stop] error while loading config file '%s': %s.\n",
						currentConf, err.Error())
				} else {
					destroyAction := DestroyAction{}
					destroyAction.Run([]string{
						"-machine", data.Machine.MachineName,
						"-force", "true"})
				}
			}

		}
	default:
		{
			log.Printf("[qemuctl::actions::service] unknown service action '%s'.\n", serviceAction)
		}
	}

	return err
}
