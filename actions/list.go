package qemuctl_actions

import (
	"fmt"
	"os"
	"strings"

	helpers "github.com/lapuglisi/qemuctl/helpers"
	runtime "github.com/lapuglisi/qemuctl/runtime"
)

type ListAction struct {
	showHeadings bool
	namesOnly    bool
	showFull     bool
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
		case "--full":
			{
				action.showFull = true
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
			headings = fmt.Sprintf("%-32s %-16s %-12s", "MACHINE", "STATUS", "QEMU PID")
			if action.showFull {
				headings = fmt.Sprintf("%s %-16s %-16s", headings, "VNC", "SPICE")
			}
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

			if action.namesOnly {
				fmt.Println(machine.Name)
			} else {
				fmt.Printf("%-32s %-16s %-12s", machine.Name, machine.Status, qemuPid)
				if action.showFull {
					ch := helpers.NewConfigHandler(machine.ConfigFile)
					cd, err := ch.ParseConfigFile()
					if err == nil {
						fmt.Print(" ")
						if cd.Display.VNC.Enabled {
							fmt.Printf("%-16s", cd.Display.VNC.Listen)
						} else {
							fmt.Printf("%s", strings.Repeat(" ", 16))
						}

						if cd.Display.Spice.Enabled {
							fmt.Printf("%-16s", fmt.Sprintf("%s:%d", cd.Display.Spice.Address, cd.Display.Spice.Port))
						} else {
							fmt.Printf("%s", strings.Repeat(" ", 16))
						}
					}
				}
				fmt.Println()
			}
		}
	}

	fmt.Println("")
	return nil
}

func (action *ListAction) getMachine(machineName string) (machine *runtime.Machine) {
	return runtime.NewMachine(machineName)
}

func (action *ListAction) getMachines(machineStatus string) (machines []string) {
	machines = make([]string, 0)
	qemuctlDir := runtime.GetMachinesBaseDir()

	/* Iterate through subdirs */
	dirEntries, err := os.ReadDir(qemuctlDir)
	if err != nil {
		return nil
	}

	for _, _value := range dirEntries {
		if _value.Type().IsDir() {
			machine := action.getMachine(_value.Name())

			if machine.Status == machineStatus {
				machines = append(machines, machine.Name)
			}
		}
	}

	return machines
}
