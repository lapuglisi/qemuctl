package qemuctl_actions

import (
	"fmt"
	"os"
	"strings"

	runtime "github.com/lapuglisi/qemuctl/runtime"
)

type ListAction struct {
	showHeadings bool
	namesOnly    bool
}

func (action *ListAction) Run(arguments []string) (err error) {
	action.showHeadings = true
	action.namesOnly = false

	for _, _value := range arguments {
		switch _value {
		case "--no-headings":
			{
				action.showHeadings = false
				break
			}
		case "--names-only":
			{
				action.namesOnly = true
				break
			}
		default:
			{
				fmt.Printf("\033[33mwarning\033[0m: invalid option '%s'\n", _value)
			}
		}
	}

	qemuctlDir := runtime.GetMachinesBaseDir()

	/* Iterate through subdirs */
	dirEntries, err := os.ReadDir(qemuctlDir)
	if err != nil {
		return err
	}

	if action.showHeadings {
		headings := ""
		if action.namesOnly {
			headings = fmt.Sprintf("%-32s", "MACHINE")
		} else {
			headings = fmt.Sprintf("%-32s %-16s %-16s %-12s", "MACHINE", "STATUS", "SSH", "QEMU PID")
		}
		fmt.Println(headings)
		fmt.Printf("%s\n", strings.Repeat("-", len(headings)))
	}

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

			if action.namesOnly {
				fmt.Println(machine.Name)
			} else {
				fmt.Printf("%-32s %-16s %-16s %-12s\n",
					machine.Name, machine.Status, sshString, qemuPid)
			}
		}
	}

	fmt.Println("")
	return nil
}

func (action *ListAction) getMachine(machineName string) (machine *runtime.Machine) {
	return runtime.NewMachine(machineName)
}
