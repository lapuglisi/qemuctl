package qemuctl_actions

import (
	"fmt"

	qemuctl_qemu "github.com/lapuglisi/qemuctl/qemu"
	runtime "github.com/lapuglisi/qemuctl/runtime"
)

func init() {

}

type StopAction struct {
	machineName string
}

func (action *StopAction) Run(arguments []string) (err error) {
	var machine *runtime.Machine

	if len(arguments) == 0 {
		return fmt.Errorf("machine name is mandatory")
	}

	if action.machineName = arguments[0]; len(action.machineName) == 0 {
		return fmt.Errorf("machine name is mandatory")
	}

	machine = runtime.NewMachine(action.machineName)

	if !machine.Exists() {
		return fmt.Errorf("machine '%s' does not exist", machine.Name)
	}

	qemuMonitor := qemuctl_qemu.NewQemuMonitor(machine)

	fmt.Printf("[qemuctl] Stopping machine '%s'...", action.machineName)

	err = qemuMonitor.SendShutdownCommand()
	if err != nil {
		fmt.Printf("\033[33m error!\033[0m\n")
		//machine.QemuPid = 0
		machine.SSHLocalPort = 0
		machine.Status = runtime.MachineStatusDegraded
		machine.UpdateData()
		return err
	}

	// Now, update machine status
	machine.QemuPid = 0
	machine.SSHLocalPort = 0
	machine.Status = runtime.MachineStatusStopped
	machine.UpdateData()

	fmt.Printf("\033[32m ok!\033[0m\n")

	return nil
}
