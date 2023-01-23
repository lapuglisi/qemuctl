package qemuctl_runtime

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func init() {

}

const (
	RuntimeBaseDirName     string = "qemuctl"
	RuntimeQemuPIDFileName string = "qemu.pid"
	RuntimeConfFileName    string = "qemuctl.yaml"
)

func GetRuntimeDir() string {
	var osRunPath string = "/var/run"
	_, err := os.Stat(osRunPath)
	if os.IsNotExist(err) {
		osRunPath = "/run"
	}

	return fmt.Sprintf("%s/%s", osRunPath, RuntimeBaseDirName)
}

func GetMachinesBaseDir() string {
	return fmt.Sprintf("%s/%s", GetRuntimeDir(), MachineBaseDirectoryName)
}

func GetSystemConfDir() string {
	return fmt.Sprintf("/etc/%s", RuntimeBaseDirName)
}

func SetupRuntimeData() (err error) {
	var qemuctlDir string = GetRuntimeDir()
	var etcConfDir string = GetSystemConfDir()
	var logDir string = "/var/log"

	/* Setup log */
	logFilePath := fmt.Sprintf("%s/qemuctl.log", logDir)
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0744)
	if err != nil {
		return err
	}

	log.SetOutput(logFile)
	/**************************/

	err = nil

	/* Create directory {RUN}/qemuctl if it does not exits */
	log.Printf("[qemuctl::runtime] checking for directory '%s'...\n", qemuctlDir)
	_, err = os.Stat(qemuctlDir)
	if os.IsNotExist(err) {
		/* Create qemuctl directory */
		log.Printf("[qemuctl::runtime] creating directory '%s'\n", qemuctlDir)

		err = os.Mkdir(qemuctlDir, os.ModeDir|os.ModePerm)
		if err != nil {
			return err
		}
	}

	/* create directory /etc/qemuctl if it does not exist */
	_, err = os.Stat(etcConfDir)
	if os.IsNotExist(err) {
		/* Create qemuctl directory */
		log.Printf("[qemuctl::runtime] creating directory '%s'\n", etcConfDir)

		err = os.Mkdir(etcConfDir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	runtimeConf := fmt.Sprintf("%s/%s", etcConfDir, RuntimeConfFileName)
	log.Printf("[qemuctl::runtime] checking for file '%s'\n", runtimeConf)
	_, err = os.Stat(runtimeConf)
	if os.IsNotExist(err) {
		/* Create empty /etc/qemuctl/qemuctl.yaml file */
		log.Printf("[qemuctl::runtime] creating empty file '%s'\n", runtimeConf)

		_, err = os.Create(runtimeConf)
		if err != nil {
			return err
		}
	}

	/* Setup Machines Runtime */
	machinesDir := fmt.Sprintf("%s/machines", qemuctlDir)
	if _, err = os.Stat(machinesDir); os.IsNotExist(err) {
		os.Mkdir(machinesDir, 0744)
	}

	log.Println("[qemuctl::runtime] setup runtime done")

	return nil
}

func FileExists(filePath string) (yes bool) {
	info, err := os.Stat(filePath)

	return (err == nil && !info.IsDir())
}

func SetupSignalHandler(signalHandler func(signal os.Signal)) {

	log.Println("[signals] setting up signal handler")
	go func(handler func(os.Signal)) {
		var osSignals chan (os.Signal) = make(chan os.Signal, 1)
		var doLoop bool = true

		log.Println("[signals] inside signal subroutine")

		log.Println("[signals] installing signal notify")
		signal.Notify(osSignals, syscall.SIGABRT, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

		for doLoop {
			log.Println("[signals] waiting for some signal")
			osSignal := <-osSignals

			log.Printf("[signals] got signal %s", osSignal.String())
			signalHandler(osSignal)

			switch osSignal {
			case syscall.SIGABRT, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM:
				{
					doLoop = false
				}
			default:
				{
					// does nothing
				}
			}
		}
	}(signalHandler)
}
