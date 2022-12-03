package qemuctl_runtime

import (
	"fmt"
	"log"
	"os"
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
