package qemuctl_actions

import (
	"fmt"

	qemuctl_qemu "luizpuglisi.com/qemuctl/qemu"
	runtime "luizpuglisi.com/qemuctl/runtime"
)

type StatusAction struct {
	machineName string
}

func (action *StatusAction) Run(arguments []string) (err error) {
	var machine *runtime.Machine
	var qemuMonitor *qemuctl_qemu.QemuMonitor
	var machineStatus *qemuctl_qemu.QmpQueryStatusResult

	if len(arguments) < 1 {
		return fmt.Errorf("machine name is mandatory")
	}

	if action.machineName = arguments[0]; len(action.machineName) == 0 {
		return fmt.Errorf("machine name is mandatory")
	}

	machine = runtime.NewMachine(action.machineName)
	qemuMonitor = qemuctl_qemu.NewQemuMonitor(machine)

	if !machine.Exists() {
		fmt.Printf("[\033[33mqemuctl\033[0m] machine '%s' does not exist\n",
			action.machineName)
		return fmt.Errorf("invalid machine name")
	}

	machineStatus, err = qemuMonitor.QueryStatus()
	if err != nil {
		return err
	}

	if machineStatus.Return.Running {
		fmt.Printf("[\033[33mqemuctl\033[0m] machine '%s' is \033[32m%s\033[0m\n",
			action.machineName, machineStatus.Return.Status)
	} else {
		fmt.Printf("[\033[33mqemuctl\033[0m] machine '%s' is \033[33m%s\033[0m\n",
			action.machineName, machineStatus.Return.Status)
	}

	return nil
}
