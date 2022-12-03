package qemuctl_actions

import (
	"fmt"
	"os"
	"strings"

	runtime "luizpuglisi.com/qemuctl/runtime"
)

type ListAction struct {
}

func (action *ListAction) Run(arguments []string) (err error) {

	qemuctlDir := runtime.GetMachinesBaseDir()

	/* Iterate through subdirs */
	dirEntries, err := os.ReadDir(qemuctlDir)
	if err != nil {
		return err
	}

	fmt.Printf("%-32s %-16s %-16s %-12s\n", "MACHINE", "STATUS", "SSH", "QEMU PID")
	fmt.Printf("%s\n", strings.Repeat("-", 76))
	for _, _value := range dirEntries {
		if _value.Type().IsDir() {
			machine := action.getMachine(_value.Name())

			/* Format QEMU PID */
			qemuPid := "N/A"
			if machine.QemuPid > 0 {
				qemuPid = fmt.Sprint(machine.QemuPid)
			}

			/* Format SSH string  */
			sshString := "N/A"
			if machine.SSHLocalPort > 0 {
				sshString = fmt.Sprintf("127.0.0.1:%d", machine.SSHLocalPort)
			}

			fmt.Printf("%-32s %-16s %-16s %-12s\n",
				machine.Name, machine.Status, sshString, qemuPid)
		}
	}

	fmt.Println("")
	return nil
}

func (action *ListAction) getMachine(machineName string) (machine *runtime.Machine) {
	return runtime.NewMachine(machineName)
}
