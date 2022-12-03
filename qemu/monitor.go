package qemuctl_qemu

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"

	runtime "github.com/lapuglisi/qemuctl/runtime"
)

func init() {

}

const (
	QemuMonitorSocketFileName string = "qemu-monitor.sock"
	QemuMonitorDefaultID      string = "qemu-mon-qmp"
)

type QemuMonitor struct {
	Machine *runtime.Machine
}

func NewQemuMonitor(machine *runtime.Machine) *QemuMonitor {
	return &QemuMonitor{
		Machine: machine,
	}
}

func (monitor *QemuMonitor) ReadQmpHeader(unix net.Conn) (qmpHeader *QmpHeader, err error) {
	const BufferSize = 1024

	var dataBytes []byte = make([]byte, 0)
	var buffer []byte = make([]byte, BufferSize)
	var nBytes int = 0

	qmpHeader = &QmpHeader{}

	dataBytes = make([]byte, 0)
	for nBytes, err = unix.Read(buffer); err == nil && nBytes > 0; {
		dataBytes = append(dataBytes, buffer[:nBytes]...)
		if nBytes < BufferSize {
			break
		}
	}

	if err != nil {
		return nil, err
	}

	/* Now that we have dataBytes, Unmarshal it to QmpHeader */
	err = json.Unmarshal(dataBytes, qmpHeader)
	if err != nil {
		return nil, fmt.Errorf("json error: %s", err.Error())
	}

	return qmpHeader, nil
}

func (monitor *QemuMonitor) GetUnixSocketPath() string {
	return fmt.Sprintf("%s/%s", monitor.Machine.RuntimeDirectory, QemuMonitorSocketFileName)
}

func (monitor *QemuMonitor) GetChardevSpec() string {
	return fmt.Sprintf("socket,id=%s,path=%s,server=on,wait=off",
		QemuMonitorDefaultID, monitor.GetUnixSocketPath())
}

func (monitor *QemuMonitor) GetMonitorSpec() string {
	return fmt.Sprintf("chardev:%s", QemuMonitorDefaultID)
}

func (monitor *QemuMonitor) GetPidFilePath() string {
	return fmt.Sprintf("%s/%s",
		monitor.Machine.RuntimeDirectory, runtime.RuntimeQemuPIDFileName)
}

func (monitor *QemuMonitor) GetPidFileData() (pidString string, err error) {
	var filePath string = monitor.GetPidFilePath()
	var fileData []byte

	log.Printf("[monitor] reading PID from file '%s'", filePath)

	fileData, err = os.ReadFile(filePath)
	if err != nil {
		return "0", err
	}

	pidString = strings.TrimSpace(string(fileData))

	return pidString, nil
}

func (monitor *QemuMonitor) GetControlSocket() (unix net.Conn, err error) {
	var qmpCommand QmpBasicCommand

	log.Printf("[InitializeSocket] opening socket '%s'\n", monitor.GetUnixSocketPath())
	{
		unix, err = net.Dial("unix", monitor.GetUnixSocketPath())
		if err != nil {
			return nil, err
		}
	}

	log.Printf("[InitializeSocket] Reading QMP header")
	{
		_, err = monitor.ReadQmpHeader(unix)
		if err != nil {
			return nil, err
		}
	}

	log.Printf("[initialize] enabling QMP capabilities")
	qmpCommand.Command = QmpCapabilitiesCommand
	_, err = qmpCommand.Execute(unix)
	if err != nil {
		return nil, err
	}

	log.Printf("[InitializeSocket] socket initialized")
	return unix, nil
}

func (monitor *QemuMonitor) QueryStatus() (result *QmpQueryStatusResult, err error) {
	var unix net.Conn
	var qmpCommand QmpCommandQueryStatus

	/* Initialize socket */
	log.Printf("[QueryStatus] initializing socket\n")
	unix, err = monitor.GetControlSocket()
	if err != nil {
		return nil, err
	}

	/* Create QueryStatus command and send it */
	log.Printf("[QueryStatus] create query-status command\n")
	result, err = qmpCommand.Execute(unix)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (monitor *QemuMonitor) SendShutdownCommand() (err error) {
	var unix net.Conn
	var shutdownCommand QmpBasicCommand

	log.Printf("[SendShutdownCommand] initializing socket")
	unix, err = monitor.GetControlSocket()
	if err != nil {
		return err
	}
	defer unix.Close()

	log.Printf("[SendShutdownCommand] sending shutdown command")
	shutdownCommand = QmpBasicCommand{
		Command: QmpSystemPowerdownCommand,
	}
	_, err = shutdownCommand.Execute(unix)
	if err != nil {
		return err
	}

	/* Now read incoming events */
	err = nil
	log.Printf("[SendShutdownCommand] reading incoming events")
	qmpEvent := QmpEventResult{}
	for {
		err = qmpEvent.ReadEvent(unix)
		if err != nil {
			break
		}
		log.Printf("[SendShutdownCommand] event received: %v", qmpEvent)
	}

	if err != nil && err == io.EOF {
		log.Printf("[monitor] ReadEvent returned err == EOF; ignoring")
		err = nil
	}

	return err
}
