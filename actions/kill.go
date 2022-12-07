package qemuctl_actions

import (
	"fmt"
	"log"
	"os"

	runtime "github.com/lapuglisi/qemuctl/runtime"
)

type KillAction struct {
	machineName string
}

func (action *KillAction) Run(arguments []string) (err error) {
	var machine *runtime.Machine
	var qemuProcess *os.Process

	if len(arguments) == 0 {
		return fmt.Errorf("qemuctl kill: machine name is mandatory")
	}

	action.machineName = arguments[0]
	machine = runtime.NewMachine(action.machineName)

	fmt.Printf("[qemuctl] killing machine '%s'... ", action.machineName)
	if !machine.Exists() {
		err = fmt.Errorf("machine '%s' does not exist", action.machineName)
		fmt.Printf("\033[31;1merror\033[0m: %s\n", err.Error())
		return err
	}

	log.Printf("qemuctl kill: finding process with PID %d", machine.QemuPid)
	qemuProcess, err = os.FindProcess(machine.QemuPid)
	if err != nil {
		fmt.Printf("\033[31;1merror\033[0m: %s\n", err.Error())
		log.Printf("error while getting qemu process #%d: %s", machine.QemuPid, err.Error())
		return err
	}

	log.Printf("qemuctl kill: killing process with PID %d", machine.QemuPid)
	err = qemuProcess.Kill()
	if err != nil {
		fmt.Printf("\033[31;1merror\033[0m: %s\n", err.Error())
		log.Printf("error while killing process #%d: %s", qemuProcess.Pid, err.Error())
		return err
	}

	log.Printf("qemuctl kill: QEMU process #%d killed", machine.QemuPid)

	machine.QemuPid = 0
	machine.SSHLocalPort = 0
	machine.Status = runtime.MachineStatusStopped
	err = machine.UpdateData()

	if err != nil {
		fmt.Printf("\033[31;1merror\033[0m: %s\n", err.Error())
	} else {
		fmt.Printf("\033[32;1ok!\033[0m\n")
	}

	return err
}
