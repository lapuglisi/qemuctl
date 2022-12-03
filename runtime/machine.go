package qemuctl_runtime

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"syscall"
)

func init() {

}

// Machine constants
const (
	MachineBaseDirectoryName string = "machines"
	MachineDataFileName      string = "machine-data.json"
	MachineStatusStarted     string = "started"
	MachineStatusStopped     string = "stopped"
	MachineStatusDegraded    string = "degraded"
	MachineStatusUnknown     string = "unknown"
	MachineConfigFileName    string = "config.yaml"
)

type MachineData struct {
	QemuPid      int    `json:"qemuProcessPID"`
	State        string `json:"machineState"`
	SSHLocalPort int    `json:"sshLocalPort"`
}

type Machine struct {
	Name             string
	Status           string
	QemuPid          int
	SSHLocalPort     int
	RuntimeDirectory string
	ConfigFile       string
	initialized      bool
}

func NewMachine(machineName string) (machine *Machine) {
	var runtimeDirectory string = fmt.Sprintf("%s/%s/%s",
		GetUserDataDir(), MachineBaseDirectoryName, machineName)
	var dataFile string = fmt.Sprintf("%s/%s", runtimeDirectory, MachineDataFileName)
	configFile := fmt.Sprintf("%s/%s", runtimeDirectory, MachineConfigFileName)

	var fileData []byte
	var machineData MachineData = MachineData{
		State:        MachineStatusUnknown,
		SSHLocalPort: 0,
	}

	fileData, err := os.ReadFile(dataFile)
	if err != nil {
		log.Printf("error: could not open data file: %s\n", err.Error())
	} else {
		err = json.Unmarshal(fileData, &machineData)
		if err != nil {
			log.Printf("[machine] could not obtain machine data: %s", err.Error())
			return nil
		}
	}

	machine = &Machine{
		Name:             machineName,
		Status:           machineData.State,
		QemuPid:          machineData.QemuPid,
		SSHLocalPort:     machineData.SSHLocalPort,
		RuntimeDirectory: runtimeDirectory,
		ConfigFile:       configFile,
		initialized:      true,
	}

	/* Make sure to check if qemu's process is actually running */
	if machine.IsStarted() {
		log.Printf("[machine] checking for pid file")
		log.Printf("[machine] checking for qemu process #%d", machineData.QemuPid)

		if machine.QemuPid > 0 {
			procHandle, err := os.FindProcess(machine.QemuPid)
			if err == nil {
				err = procHandle.Signal(syscall.SIGCONT)
			}
			if err != nil {
				log.Printf("[machine] looks like the process %d is not there (%v). updating machine status", machineData.QemuPid, err)
				machine.QemuPid = 0
				machine.Status = MachineStatusDegraded
				machine.SSHLocalPort = 0
				machine.UpdateStatus(MachineStatusDegraded)
			}
		} else {
			log.Printf("[machine] invalid PID #%d for machine '%s'", machineData.QemuPid, machineName)
			log.Printf("[machine] PID %d is not valid, machine is therefore degraded; updating machine status", machineData.QemuPid)
			machine.QemuPid = 0
			machine.Status = MachineStatusDegraded
			machine.SSHLocalPort = 0
			machine.UpdateStatus(MachineStatusDegraded)
		}
	}

	return machine
}

func (m *Machine) Exists() bool {
	fileInfo, err := os.Stat(m.RuntimeDirectory)
	if os.IsNotExist(err) {
		return false
	}

	return fileInfo.IsDir()
}

func (m *Machine) Destroy() bool {
	log.Printf("qemuctl: destroying machine %s\n", m.Name)

	err := os.RemoveAll(m.RuntimeDirectory)

	return err == nil
}

func (m *Machine) IsStarted() bool {
	return (strings.Compare(MachineStatusStarted, m.Status) == 0)
}

func (m *Machine) IsStopped() bool {
	return (strings.Compare(MachineStatusStopped, m.Status) == 0)
}

func (m *Machine) IsDegraded() bool {
	return (strings.Compare(MachineStatusDegraded, m.Status) == 0)
}

func (m *Machine) IsUnknown() bool {
	return (strings.Compare(MachineStatusUnknown, m.Status) == 0)
}

func (m *Machine) UpdateStatus(status string) (err error) {
	var statusFile string = fmt.Sprintf("%s/%s", m.RuntimeDirectory, MachineDataFileName)
	var fileHandle *os.File
	var machineData MachineData

	log.Printf("[UpdateStatus] opening file '%s'\n", statusFile)
	fileHandle, err = os.OpenFile(statusFile, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0755)
	if err != nil {
		return err
	}

	/* populate new MachineData */
	machineData = MachineData{
		QemuPid:      m.QemuPid,
		SSHLocalPort: m.SSHLocalPort,
		State:        status,
	}

	switch status {
	case MachineStatusDegraded, MachineStatusStarted, MachineStatusStopped, MachineStatusUnknown:
		{
			log.Printf("[UpdateStatus] updating file '%s' with [%v].\n", statusFile, machineData)
			jsonBytes, err := json.Marshal(machineData)

			if err != nil {
				log.Printf("[UpdateStatus] error while generating new JSON: '%s'.\n", err.Error())
			} else {
				log.Printf("[UpdateStatus] writing [%s] to file '%s'.\n", string(jsonBytes), statusFile)
				_, err = fileHandle.Write(jsonBytes)

				if err != nil {
					log.Printf("[UpdateStatus] error while updating '%s': %s\n", statusFile, err.Error())
				}
			}
			break
		}
	default:
		{
			err = fmt.Errorf("invalid machine status '%s'", status)
		}
	}
	fileHandle.Close()

	return err
}

func (m *Machine) CreateRuntime() {
	os.Mkdir(m.RuntimeDirectory, 0744)
}

func (m *Machine) UpdateConfigFile(sourcePath string) (err error) {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	targetFile, err := os.Create(m.ConfigFile)
	if err != nil {
		return err
	}
	defer targetFile.Close()

	_, err = io.Copy(targetFile, sourceFile)

	return err
}

func (m *Machine) GetMachineFileData(fileName string) (data []byte, err error) {
	var filePath string = fmt.Sprintf("%s/%s", m.RuntimeDirectory, fileName)

	data, err = os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return data, nil
}
