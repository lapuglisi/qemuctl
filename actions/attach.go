package qemuctl_actions

import (
	"flag"
	"fmt"
	"log"
	"regexp"

	helpers "github.com/lapuglisi/qemuctl/helpers"
	runtime "github.com/lapuglisi/qemuctl/runtime"
)

type AttachAction struct {
	machineName string
	background  bool
}

func (action *AttachAction) Run(arguments []string) (err error) {
	var flagSet *flag.FlagSet = flag.NewFlagSet("qemuctl attach", flag.ExitOnError)
	flagSet.BoolVar(&action.background, "bg", false, "run in background")

	err = flagSet.Parse(arguments)
	if err != nil {
		return err
	}

	/* Do flags validation */
	action.machineName = flagSet.Args()[0]
	if len(action.machineName) == 0 {
		flagSet.Usage()
		return fmt.Errorf("machine name is mandatory")
	}

	/* Do proper handling */
	err = action.handleAttach()
	if err != nil {
		return err
	}

	return nil
}

func (action *AttachAction) handleAttach() (err error) {
	var configData *helpers.ConfigurationData = nil
	var machine *runtime.Machine
	var vncRegex regexp.Regexp = *regexp.MustCompile(`[0-9\.]+:\d+`)

	err = nil

	machine = runtime.NewMachine(action.machineName)

	log.Printf("[create] using config file: %s", machine.ConfigFile)

	configHandle := helpers.NewConfigHandler(machine.ConfigFile)
	configData, err = configHandle.ParseConfigFile()
	if err != nil {
		return err
	}

	if configData.Display.VNC.Enabled {
		vncListen := ""
		log.Printf("[qemuctl::attach] '%s': Attaching to VNC display (%s).... ",
			machine.Name, configData.Display.VNC.Listen)

		if vncRegex.Match([]byte(configData.Display.VNC.Listen)) {
			vncListen = configData.Display.VNC.Listen
		} else {
			vncListen = fmt.Sprintf("127.0.0.1:%s", configData.Display.VNC.Listen)
		}

		runtime.LaunchVNCViewer(vncListen, action.background)
	} else if configData.Display.Spice.Enabled {
		spiceConnect := ""
		unixSocket, err := machine.GetSpiceSocketPath()
		if err == nil {
			spiceConnect = fmt.Sprintf("spice+unix:///%s", unixSocket)
			log.Printf("[qemuctl::attach] '%s': Attaching to Spice socket %s.... ",
				machine.Name, unixSocket)

			runtime.LaunchSpiceViewer(spiceConnect, action.background)
		} else {
			spiceConnect = fmt.Sprintf("spice://%s:%d",
				runtime.GetValueOrDefault(configData.Display.Spice.Address, "127.0.0.1"),
				configData.Display.Spice.Port)

			log.Printf("[qemuctl::attach] '%s': Attaching to Spice display %s.... ",
				machine.Name, spiceConnect)

			runtime.LaunchSpiceViewer(spiceConnect, action.background)
		}
	}

	return nil
}
