package qemuctl_runtime

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const (
	QemuUser = "qemuctl"
)

func init() {

}

const (
	RuntimeBaseDirName      string = "qemuctl"
	RuntimeQemuPIDFileName  string = "qemu.pid"
	RuntimeAutoStartDirName string = "autostart"
	RuntimeVNCViewerPath    string = "/usr/bin/vncviewer"
	RuntimeSpiceViewerPath  string = "/usr/bin/remote-viewer"
)

func GetRuntimeDir() string {
	var runtimeDir string = "/var/run"
	_, err := os.Stat(runtimeDir)
	if os.IsNotExist(err) {
		fmt.Printf("ERROR: directory '%s' does not exist.\n", runtimeDir)
		return ""
	}

	return fmt.Sprintf("%s/%s", runtimeDir, RuntimeBaseDirName)
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

	_, err = os.Stat(qemuctlDir)
	if os.IsNotExist(err) {
		fmt.Printf("INFO: directory '%s' does not exist. Creating it.\n", qemuctlDir)
		os.Mkdir(qemuctlDir, os.ModeDir|os.ModePerm)
	}

	/* Setup log */
	{
		logFilePath := fmt.Sprintf("%s/qemuctl.log", logDir)
		logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0744)
		if err != nil {
			return err
		}
		log.SetOutput(logFile)
	}

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

	autoStartDir := fmt.Sprintf("%s/%s", etcConfDir, RuntimeAutoStartDirName)
	log.Printf("[qemuctl::runtime] checking for directory '%s'\n", autoStartDir)
	_, err = os.Stat(autoStartDir)
	if os.IsNotExist(err) {
		/* Create directory /etc/qemuctl/autostart file */
		log.Printf("[qemuctl::runtime] creating directory '%s'\n", autoStartDir)

		err = os.Mkdir(autoStartDir, os.ModePerm)
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

func CopyFile(source string, target string) (err error) {

	log.Printf("[qemuctl::runtime::copy] copying file '%s' to '%s'...\n", source, target)

	log.Printf("[qemuctl::runtime::copy] reading file '%s'...\n", source)
	sourceBytes, err := os.ReadFile(source)
	if err != nil {
		log.Printf("[qemuctl::runtime::copy] error while copying file: %s...\n", err.Error())
		return err
	} else {
		log.Printf("[qemuctl::runtime::copy] successfully read file '%s'...\n", source)
	}

	log.Printf("[qemuctl::runtime::copy] writing file '%s'...\n", target)
	err = os.WriteFile(target, sourceBytes, os.ModePerm|fs.FileMode(os.O_CREATE))
	if err != nil {
		log.Printf("[qemuctl::runtime::copy] error while writing file: %s...\n", err.Error())
		return err
	} else {
		log.Printf("[qemuctl::runtime::copy] successfully wrote file '%s'...\n", target)
	}

	return nil
}

func LinkFile(source string, target string) (err error) {

	log.Printf("[qemuctl::runtime::link] linking file '%s' to '%s'...\n", source, target)

	return os.Symlink(source, target)
}

func LaunchVNCViewer(connect string, background bool) (err error) {
	var procAttrs *os.ProcAttr = nil
	var procState *os.ProcessState = nil

	// TODO: use the log feature; DONE
	log.Println("[QemuCommand::Launch] Executing vncviewer with:")
	log.Printf("vncviewer ....... %s\n", RuntimeVNCViewerPath)
	log.Printf("vnc args ........ %s\n", connect)

	/* Actual execution of VNC Viewr */
	err = nil

	log.Printf("[launch] creating vncviewer command struct")
	procAttrs = &os.ProcAttr{
		Dir: os.ExpandEnv("$HOME"),
		Env: os.Environ(),
		Files: []*os.File{
			os.Stdin,
			os.Stdout,
			os.Stderr,
		},
		Sys: nil,
	}
	execArgs := make([]string, 0)
	execArgs = append(execArgs, RuntimeVNCViewerPath)
	execArgs = append(execArgs, connect)

	log.Printf("[launch] starting qemu process")
	vncProcess, err := os.StartProcess(RuntimeVNCViewerPath, execArgs, procAttrs)
	if err != nil {
		log.Printf("[launch] error starting process: %s", err.Error())
		return err
	}

	if !background {
		log.Printf("[launch] waiting for vncviewer process to finish")
		procState, err = vncProcess.Wait()
		if err != nil {
			log.Printf("[launch] waiting for qemu command failed: %s (exit code: %d)",
				err.Error(), procState.ExitCode())

			vncProcess.Kill()
			return err
		}
	} else {
		err = vncProcess.Signal(syscall.SIGCONT)
		if err == nil {
			log.Printf("[launch] vncviewer process running with PID %d", vncProcess.Pid)
		} else {
			log.Printf("[launch] vncviewer process error: %s", err.Error())
		}
	}

	return err
}

func LaunchSpiceViewer(host string, port int, background bool) (err error) {
	var procAttrs *os.ProcAttr = nil
	var procState *os.ProcessState = nil
	var spiceArgs string

	if len(host) == 0 {
		spiceArgs = fmt.Sprintf("spice://127.0.0.1:%d", port)
	} else {
		spiceArgs = fmt.Sprintf("spice://%s:%d", host, port)
	}

	// TODO: use the log feature; DONE
	log.Println("[LaunchSpiceViewer] Executing Spice with:")
	log.Printf("remote-viewer ....... %s\n", RuntimeSpiceViewerPath)
	log.Printf("spice args ......... %s\n", spiceArgs)

	/* Actual execution of VNC Viewr */
	err = nil

	log.Printf("[LaunchSpiceViewer] creating remote-viewer command struct")
	procAttrs = &os.ProcAttr{
		Dir: os.ExpandEnv("$HOME"),
		Env: os.Environ(),
		Files: []*os.File{
			os.Stdin,
			os.Stdout,
			os.Stderr,
		},
		Sys: nil,
	}
	execArgs := make([]string, 0)
	execArgs = append(execArgs, RuntimeSpiceViewerPath)
	execArgs = append(execArgs, spiceArgs)

	log.Printf("[LaunchSpiceViewer] starting remote-viewer process")
	spiceProcess, err := os.StartProcess(RuntimeSpiceViewerPath, execArgs, procAttrs)
	if err != nil {
		log.Printf("[LaunchSpiceViewer] error starting process: %s", err.Error())
		return err
	}

	if background {
		err = spiceProcess.Signal(syscall.SIGCONT)
		if err == nil {
			log.Printf("[launch] spice process running with PID %d", spiceProcess.Pid)
		} else {
			log.Printf("[launch] spice process error: %s", err.Error())
		}
	} else {
		log.Printf("[LaunchSpiceViewer] waiting for remote-viewer process to finish")
		procState, err = spiceProcess.Wait()
		if err != nil {
			log.Printf("[LaunchSpiceViewer] waiting for remote-viewer command failed: %s (exit code: %d)",
				err.Error(), procState.ExitCode())

			spiceProcess.Kill()
		}
	}

	return err
}
