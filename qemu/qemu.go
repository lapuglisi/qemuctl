package qemuctl_qemu

import (
	"fmt"
	"log"

	//"os"
	"os/exec"
	"regexp"
	"strings"

	config "github.com/lapuglisi/qemuctl/helpers"
	runtime "github.com/lapuglisi/qemuctl/runtime"
)

const (
	QemuDefaultSystemBin string = "qemu-system-x86_64"
)

type QemuCommand struct {
	QemuPath      string
	Configuration *config.ConfigurationData
	Monitor       *QemuMonitor
}

func NewQemuCommand(configData *config.ConfigurationData, qemuMonitor *QemuMonitor) (qemu *QemuCommand) {
	var qemuPath string
	var qemuBinary string = configData.QemuBinary

	if len(qemuBinary) == 0 {
		qemuBinary = QemuDefaultSystemBin
	}

	qemuPath, err := exec.LookPath(qemuBinary)
	if err != nil {
		qemuPath = qemuBinary
	}

	return &QemuCommand{
		QemuPath:      qemuPath,
		Configuration: configData,
		Monitor:       qemuMonitor,
	}
}

func (qemu *QemuCommand) getBoolString(qemuFlag bool, trueValue string, falseValue string) string {
	if qemuFlag {
		return trueValue
	}

	return falseValue
}

func (qemu *QemuCommand) getKeyValuePair(include bool, key string, value string) string {
	if include {
		return fmt.Sprintf("%s=%s", key, value)
	}

	return ""
}

func (qemu *QemuCommand) appendQemuArg(argSlice []string, argKey string, argValue string) []string {
	return append(argSlice, []string{argKey, argValue}...)
}

func (qemu *QemuCommand) getQemuArgs() (qemuArgs []string, err error) {
	/* Config specific */
	var machineSpec string
	var netSpec string

	var cd *config.ConfigurationData = qemu.Configuration

	log.Printf("[debug::getQemuArgs] configData is [%v]", cd)

	var machine *runtime.Machine = runtime.NewMachine(cd.Machine.MachineName)
	var monitor *QemuMonitor = NewQemuMonitor(machine)

	/* VNC Spec parser */
	var vncRegex regexp.Regexp = *regexp.MustCompile(`[0-9\.]+:\d+`)

	// qemuArgs = append(qemuArgs, qemu.QemuPath)

	/* Do the config stuff */
	if cd.Machine.EnableKVM {
		qemuArgs = append(qemuArgs, "-enable-kvm")
	}

	// -- Machine spec (type and accel)
	{
		machineSpec = fmt.Sprintf("type=%s", cd.Machine.MachineType)
		if len(cd.Machine.AccelType) > 0 {
			machineSpec = fmt.Sprintf("%s,accel=%s", machineSpec, cd.Machine.AccelType)
		}

		qemuArgs = qemu.appendQemuArg(qemuArgs, "-machine", machineSpec)

		/* Add CPU spec */
		qemuArgs = qemu.appendQemuArg(qemuArgs, "-cpu", cd.Machine.CPU)

		/* TPM Specification, if any */
		if cd.Machine.TPM.Enabled {
			if cd.Machine.TPM.Passthrough.Enabled {
				tpmSpec := fmt.Sprintf("passthrough,id=%s%s%s",
					cd.Machine.TPM.Passthrough.ID,
					qemu.getKeyValuePair(len(cd.Machine.TPM.Passthrough.Path) > 0, ",path", cd.Machine.TPM.Passthrough.Path),
					qemu.getKeyValuePair(len(cd.Machine.TPM.Passthrough.CancelPath) > 0, ",cancel-path", cd.Machine.TPM.Passthrough.CancelPath))

				qemuArgs = qemu.appendQemuArg(qemuArgs, "-tpmdev", tpmSpec)
			} else if cd.Machine.TPM.Emulator.Enabled {
				tpmSpec := fmt.Sprintf("emulator,id=%s,chardev=%s",
					cd.Machine.TPM.Emulator.ID,
					cd.Machine.TPM.Emulator.CharDevice)

				qemuArgs = qemu.appendQemuArg(qemuArgs, "-tpmdev", tpmSpec)
			}
		}
	}

	// -- Machine Name
	if len(cd.Machine.MachineName) > 0 {
		qemuArgs = qemu.appendQemuArg(qemuArgs, "-name", cd.Machine.MachineName)
	}

	// -- Memory
	qemuArgs = qemu.appendQemuArg(qemuArgs, "-m", cd.Memory)

	// -- cpus
	qemuArgs = qemu.appendQemuArg(qemuArgs, "-smp", fmt.Sprintf("%d", cd.CPUs))

	// -- CDROM
	if len(cd.Disks.ISOCDrom) > 0 {
		qemuArgs = qemu.appendQemuArg(qemuArgs, "-cdrom", cd.Disks.ISOCDrom)
	}

	/*
	 * Display specification
	 */
	if !cd.Display.EnableGraphics {
		qemuArgs = append(qemuArgs, "-nographic")
	} else {
		// -- VGA
		qemuArgs = qemu.appendQemuArg(qemuArgs, "-vga", cd.Display.VGAType)

		// -- Display
		qemuArgs = qemu.appendQemuArg(qemuArgs, "-display", cd.Display.DisplaySpec)
	}

	// VNC ?
	if cd.Display.VNC.Enabled {
		// Is it in the format "xxx.xxx.xxx.xxx:ddd" ?
		if vncRegex.Match([]byte(cd.Display.VNC.Listen)) {
			qemuArgs = qemu.appendQemuArg(qemuArgs, "-vnc", cd.Display.VNC.Listen)
		} else {
			qemuArgs = qemu.appendQemuArg(qemuArgs, "-vnc", fmt.Sprintf("127.0.0.1:%s", cd.Display.VNC.Listen))
		}
	}

	// Spice is enabled?
	if cd.Display.Spice.Enabled {
		if cd.Display.Spice.Port <= 0 {
			log.Printf("[getQemuArgs] spice is enable but spice.port is not defined")
		} else {
			spiceSpec := fmt.Sprintf("port=%d,tls-port=%d%s,disable-ticketing=%s,agent-mouse=%s,password=%s,gl=%s",
				cd.Display.Spice.Port, cd.Display.Spice.TLSPort,
				qemu.getKeyValuePair(len(cd.Display.Spice.Address) > 0, ",addr", cd.Display.Spice.Address),
				qemu.getBoolString(cd.Display.Spice.DisableTicketing, "on", "off"),
				qemu.getBoolString(cd.Display.Spice.EnableAgentMouse, "on", "off"),
				cd.Display.Spice.Password,
				qemu.getBoolString(cd.Display.Spice.OpenGL, "on", "off"))

			qemuArgs = qemu.appendQemuArg(qemuArgs, "-spice", spiceSpec)
		}
	}

	/**
	 * BIOS and Boot habling
	 */
	if len(cd.Boot.KernelPath) > 0 && len(cd.Boot.RamdiskPath) > 0 {
		// Do not use biosFile or boot related stuff. Boot directly to kernel
		qemuArgs = qemu.appendQemuArg(qemuArgs, "-kernel", cd.Boot.KernelPath)
		qemuArgs = qemu.appendQemuArg(qemuArgs, "-initrd", cd.Boot.RamdiskPath)
	} else {
		if len(cd.Boot.BiosFile) > 0 {
			qemuArgs = qemu.appendQemuArg(qemuArgs, "-bios", cd.Boot.BiosFile)
		}

		// -- Boot menu & Boot order (exclusive)
		if cd.Boot.EnableBootMenu {
			qemuArgs = qemu.appendQemuArg(qemuArgs, "-boot", "menu=on")
		} else if len(cd.Boot.BootOrder) > 0 {
			qemuArgs = qemu.appendQemuArg(qemuArgs, "-boot", "order="+cd.Boot.BootOrder)
		}
	}

	// -- Background?
	if cd.RunAsDaemon {
		qemuArgs = append(qemuArgs, "-daemonize")
	}

	// -- Network spec
	{
		/* Configure user network device */
		netSpec = fmt.Sprintf("%s,netdev=%s", cd.Net.DeviceType, cd.Net.User.ID)
		qemuArgs = qemu.appendQemuArg(qemuArgs, "-device", netSpec)

		/* Configure User NIC */
		netSpec = fmt.Sprintf("user,id=%s", cd.Net.User.ID)

		if len(cd.Net.User.IPSubnet) > 0 {
			netSpec = fmt.Sprintf("%s,net=%s", netSpec, cd.Net.User.IPSubnet)
		}

		if cd.SSH.LocalPort > 0 {
			netSpec = fmt.Sprintf("%s,hostfwd=tcp::%d-:22", netSpec, cd.SSH.LocalPort)
		}

		/* Port fowards come here */
		for _, _value := range cd.Net.User.PortForwards {
			netSpec = fmt.Sprintf("%s,hostfwd=tcp::%d-:%d", netSpec, _value.HostPort, _value.GuestPort)
		}

		qemuArgs = qemu.appendQemuArg(qemuArgs, "-netdev", netSpec)

		/*
		 * Configure bridge, if any
		 */
		if len(cd.Net.Bridge.Interface) > 0 {
			//-- Device specification
			netSpec = fmt.Sprintf("%s,netdev=%s", cd.Net.DeviceType, cd.Net.Bridge.ID)
			if len(cd.Net.Bridge.MacAddress) > 0 {
				netSpec = fmt.Sprintf("%s,mac=", cd.Net.Bridge.MacAddress)
			}
			qemuArgs = qemu.appendQemuArg(qemuArgs, "-device", netSpec)

			// Bridge definition
			netSpec = fmt.Sprintf("bridge,id=%s,br=%s", cd.Net.Bridge.ID, cd.Net.Bridge.Interface)
			if len(cd.Net.Bridge.Helper) > 0 {
				netSpec = fmt.Sprintf("%s,helper=%s", netSpec, cd.Net.Bridge.Helper)
			}
			qemuArgs = qemu.appendQemuArg(qemuArgs, "-netdev", netSpec)
		}
	}

	/*
	 * Disk specification
	 */
	if len(cd.Disks.BlockDevice) > 0 { // TODO: Use stat to check whether it is a valid block device
		driveName := "xvda"
		// Appends drive/device specification
		qemuArgs = qemu.appendQemuArg(qemuArgs, "-device", fmt.Sprintf("virtio-blk-pci,drive=%s", driveName))

		// Appends block device configuration
		qemuArgs = qemu.appendQemuArg(qemuArgs,
			"-blockdev",
			fmt.Sprintf("node-name=%s,driver=raw,file.driver=host_device,file.filename=%s", driveName, cd.Disks.BlockDevice))
	} else {
		// -- Otherwise, we finally add hard disk info
		qemuArgs = append(qemuArgs, cd.Disks.HardDisk)
	}

	/* Add a monitor specfication to be able to operate on the machine */
	qemuArgs = qemu.appendQemuArg(qemuArgs, "-chardev", monitor.GetChardevSpec())
	qemuArgs = qemu.appendQemuArg(qemuArgs, "-qmp", monitor.GetMonitorSpec())

	/* Add PIDfile spec */
	qemuArgs = qemu.appendQemuArg(qemuArgs, "-pidfile", monitor.GetPidFilePath())

	return qemuArgs, nil
}

func (qemu *QemuCommand) Launch() (err error) {
	// var procAttrs *os.ProcAttr = nil
	var qemuArgs []string

	qemuArgs, err = qemu.getQemuArgs()
	if err != nil {
		return err
	}

	// TODO: use the log feature
	log.Println("[QemuCommand::Launch] Executing QEMU with:")
	log.Printf("qemu_path ....... %s\n", qemu.QemuPath)
	log.Printf("qemu_args ....... %s\n", strings.Join(qemuArgs, " "))

	/* Actual execution of QEMU */
	err = nil
	/*
		procAttrs = &os.ProcAttr{
			Dir: os.ExpandEnv("$HOME"),
			Env: os.Environ(),
			Files: []*os.File{
				nil,
				nil,
				os.Stderr,
			},
			Sys: nil,
		}
	*/

	log.Printf("[launch] creating qemu command struct")
	qemuCmd := exec.Command(qemu.QemuPath, qemuArgs...)

	log.Printf("[launch] starting qemu command")
	err = qemuCmd.Start()
	if err != nil {
		log.Printf("[launch] error starting command: %s", err.Error())
		return err
	}

	log.Printf("[launch] waiting for qemu command to finish")
	err = qemuCmd.Wait()
	if err != nil {
		log.Printf("[launch] waiting for qemu command failed: %s", err.Error())
		cmdBytes, err := qemuCmd.CombinedOutput()

		log.Printf("[launch] cmd output: [%s] [%s]", string(cmdBytes), err.Error())
		qemuCmd.Process.Kill()
		return err
	}

	log.Printf("[launch] qemu process state: %s", qemuCmd.ProcessState.String())

	/*
		procHandle, err := os.StartProcess(qemu.QemuPath, qemuArgs, procAttrs)
		if err == nil {
			log.Printf("[qemu.launch] success: %v", procHandle)

			err := procHandle.Release()
			if err != nil {
				log.Printf("[launch] releasing the process failed: %s", err.Error())
			}
		} else {
			log.Printf("[qemu.launch] some error ocurred: %s", err.Error())
		}
	*/

	return err
}
