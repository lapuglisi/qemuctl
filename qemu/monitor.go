package qemuctl_qemu

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

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

func (monitor *QemuMonitor) GetPidFromPidFile() (procPid int, err error) {
	var filePath string = monitor.GetPidFilePath()
	var fileData []byte

	log.Printf("[monitor] reading PID from file '%s'", filePath)

	fileData, err = os.ReadFile(filePath)
	if err != nil {
		return 0, err
	}

	pidString := strings.TrimSpace(string(fileData))
	procPid, err = strconv.Atoi(pidString)

	if err != nil {
		return 0, err
	}

	return procPid, nil
}

func (monitor *QemuMonitor) WaitForPid() (procPid int, err error) {
	var filePath string = monitor.GetPidFilePath()
	var fileData []byte
	var sleepNanos time.Duration = time.Duration(1000 * 1000 * 1000) // 500 milliseconds

	for {
		log.Printf("[WaitForPid] stating file '%s'", filePath)
		_, err := os.Stat(filePath)
		if err == nil {
			break
		}

		if os.IsNotExist(err) {
			log.Printf("[WaitForPid] '%s' is not there yet. sleeping for %f second(s).",
				filePath, sleepNanos.Seconds())
			time.Sleep(sleepNanos)

			err = nil
		}
	}

	if err != nil {
		return 0, err
	}
	/* now read filePath */
	fileData, err = os.ReadFile(filePath)
	if err != nil {
		log.Printf("[WaitForPid] error while reading file: %s", err.Error())
		return 0, err
	}

	pidString := strings.TrimSpace(string(fileData))
	procPid, err = strconv.Atoi(pidString)
	if err != nil {
		log.Printf("[WaitForPid] error while converting string '%s': %s", pidString, err.Error())
	} else {
		log.Printf("[WaitForPid] got pid %d", procPid)
	}

	return procPid, err
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
	var basicResult *QmpBasicResult

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
	basicResult, err = shutdownCommand.Execute(unix)
	if err != nil {
		return err
	}

	log.Printf("[SendShutdownCommand] got QmpBasicResult [%v]", basicResult)

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
