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
	RuntimeBaseDirName     string = ".qemuctl"
	RuntimeQemuPIDFileName string = "qemu.pid"
)

func GetUserDataDir() string {
	return fmt.Sprintf("%s/%s", os.ExpandEnv("$HOME"), RuntimeBaseDirName)
}

func GetMachinesBaseDir() string {
	return fmt.Sprintf("%s/%s", GetUserDataDir(), MachineBaseDirectoryName)
}

func SetupRuntimeData() (err error) {
	var qemuctlDir string = GetUserDataDir()

	/* Create directory {userHome}/.qemuctl if it does not exits */
	_, err = os.Stat(qemuctlDir)
	if os.IsNotExist(err) {
		/* Create qemuctl directory */
		log.Printf("creating directory '%s'\n", qemuctlDir)

		err = os.Mkdir(qemuctlDir, os.ModeDir|os.ModePerm)
		if err != nil {
			return err
		}
	}

	/* Setup log */
	logFilePath := fmt.Sprintf("%s/qemuctl.log", qemuctlDir)
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0744)
	if err != nil {
		return err
	}

	log.SetOutput(logFile)
	/**************************/

	/* Setup Machines Runtime */
	machinesDir := fmt.Sprintf("%s/machines", qemuctlDir)
	if _, err = os.Stat(machinesDir); os.IsNotExist(err) {
		os.Mkdir(machinesDir, 0744)
	}

	log.Println("qemuctl: setup runtime done")

	return nil
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
