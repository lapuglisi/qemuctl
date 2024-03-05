package qemuctl_qemu

import (
	"crypto/rand"
	"fmt"
	"log"
	"os"

	//"os"
	"os/exec"
	"os/user"
	"regexp"
	"strings"

	config "github.com/lapuglisi/qemuctl/helpers"
	runtime "github.com/lapuglisi/qemuctl/runtime"
)

const (
	QemuDefaultSystemBin       string = "qemu-system-x86_64"
	QemuDefaultMacAddressBytes string = "52:54:00"
)

type QemuCommand struct {
	QemuPath      string
	Configuration *config.ConfigurationData
	Monitor       *QemuMonitor
}

var nodeDevices []string = []string{"xvda", "xvdb", "xvdc"}

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

func (qemu *QemuCommand) generateQemuMacAdress() (mac string, err error) {
	/* QEMU usually sets a mac address in the form '52:54:00:XX:XX:XX', so i'm sticking to it */
	var macBytes []byte = make([]byte, 3)

	/* read 3 random bytes into 'macBytes' */
	_, err = rand.Read(macBytes)
	if err != nil {
		return "", err
	}

	mac = fmt.Sprintf("%s:%02X:%02X:%02X", QemuDefaultMacAddressBytes, macBytes[0], macBytes[1], macBytes[2])

	return mac, nil
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

	var machine *runtime.Machine = qemu.Monitor.Machine
	var monitor *QemuMonitor = qemu.Monitor

	/* VNC Spec parser */
	var vncRegex regexp.Regexp = *regexp.MustCompile(`[0-9\.]+:\d+`)

	/* Do the config stuff */
	if cd.Machine.EnableKVM {
		qemuArgs = append(qemuArgs, "-enable-kvm")
	}

	// QEMU should never run as root
	runAsUser := cd.RunAs
	if len(runAsUser) == 0 {
		return nil, fmt.Errorf("machine config must specify 'runAs'")
	}

	user, err := user.Lookup(runAsUser)
	if err != nil {
		fmt.Printf("error: cannot find user '%s'\n", runAsUser)
		return nil, err
	}

	qemuArgs = qemu.appendQemuArg(qemuArgs, "-runas", user.Username)

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
	 * PCI Passthrough spec
	 */
	if cd.PCI.Passthrough {
		for _, device := range cd.PCI.Devices {
			qemuArgs = qemu.appendQemuArg(qemuArgs, "-device", fmt.Sprintf("vfio-pci,host=%s", device))
		}
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
		spiceSpec := ""
		if cd.Display.Spice.OpenGL {
			unixSocket, err := machine.GetSpiceSocketPath()
			if err != nil {
				return nil, err
			}
			spiceSpec = fmt.Sprintf("disable-ticketing=%s,agent-mouse=%s,gl=on,unix=on,addr=%s",
				qemu.getBoolString(cd.Display.Spice.DisableTicketing, "on", "off"),
				qemu.getBoolString(cd.Display.Spice.EnableAgentMouse, "on", "off"),
				unixSocket)
		} else {
			spiceSpec = fmt.Sprintf("port=%d%s,ipv4=%s,ipv6=%s,tls-port=%d,disable-ticketing=%s,agent-mouse=%s,gl=%s,unix=on",
				cd.Display.Spice.Port,
				qemu.getKeyValuePair(len(cd.Display.Spice.Address) > 0, ",addr", cd.Display.Spice.Address),
				qemu.getBoolString(cd.Display.Spice.EnableIPv4, "on", "off"),
				qemu.getBoolString(cd.Display.Spice.EnableIPv6, "on", "off"),
				cd.Display.Spice.TLSPort,
				qemu.getBoolString(cd.Display.Spice.DisableTicketing, "on", "off"),
				qemu.getBoolString(cd.Display.Spice.EnableAgentMouse, "on", "off"),
				qemu.getBoolString(cd.Display.Spice.OpenGL, "on", "off"))
		}

		qemuArgs = qemu.appendQemuArg(qemuArgs, "-spice", spiceSpec)
	}

	/**
	 * BIOS and Boot handling
	 */
	if len(cd.Boot.KernelPath) > 0 {
		// Do not use biosFile or boot related stuff. Boot directly to kernel

		log.Printf("[qemuArgs] using kernel image '%s'", cd.Boot.KernelPath)
		qemuArgs = qemu.appendQemuArg(qemuArgs, "-kernel", cd.Boot.KernelPath)

		if len(cd.Boot.RamdiskPath) > 0 {
			log.Printf("[qemuArgs] using ramdisk image '%s'", cd.Boot.RamdiskPath)
			qemuArgs = qemu.appendQemuArg(qemuArgs, "-initrd", cd.Boot.RamdiskPath)
		}

		if len(cd.Boot.KernelArgs) > 0 {
			log.Printf("[qemuArgs] using kernel args '%s'", cd.Boot.KernelArgs)
			qemuArgs = qemu.appendQemuArg(qemuArgs, "-append", fmt.Sprintf("'%s'", cd.Boot.KernelArgs))
		}
	} else {
		if len(cd.Boot.BiosFile) > 0 {
			// TODO: should be using -drive if=pflash,format=raw,file=/copy/of/OVMF.fd | read-write
			/*
			 1. Copy cd.Boot.BiosFile to machine's directory
			 2. Use it with -drive if=pflash,format=raw,file=/copy/of/OVMF.fd
			*/
			if err = machine.MakeBiosFileCopy(cd.Boot.BiosFile); err == nil {
				qemuArgs = qemu.appendQemuArg(qemuArgs, "-drive",
					fmt.Sprintf("if=pflash,format=raw,file=%s", machine.BiosFile))
			}
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

	// - enable soundhw='hda' ?
	if cd.Audio.Enabled {
		audioSpec := fmt.Sprintf("%s,model=%s", cd.Audio.Driver, cd.Audio.Model)
		qemuArgs = qemu.appendQemuArg(qemuArgs, "-audiodev", audioSpec)
	}

	// -- Network spec
	{
		if cd.Net.User.Enabled {
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
		}

		/*
		 * Configure bridge, if any
		 */
		if cd.Net.Bridge.Enabled {
			//-- Device specification
			for _, bridge := range cd.Net.Bridge.Interfaces {
				netSpec = fmt.Sprintf("virtio-net-pci,netdev=%s", bridge.ID)
				if len(bridge.MacAddress) > 0 {
					netSpec = fmt.Sprintf("%s,mac=%s", netSpec, bridge.MacAddress)
				} else {
					macAddr, err := qemu.generateQemuMacAdress()
					if err == nil {
						netSpec = fmt.Sprintf("%s,mac=%s", netSpec, macAddr)
					} else {
						log.Printf("[qemuctl::qemu] error while generating mac address: %s\n", err.Error())
					}
				}
				qemuArgs = qemu.appendQemuArg(qemuArgs, "-device", netSpec)

				// Bridge definition
				netSpec = fmt.Sprintf("bridge,id=%s", bridge.ID)

				if len(bridge.Interface) > 0 {
					netSpec = fmt.Sprintf("%s,br=%s", netSpec, bridge.Interface)
				}
				if len(bridge.Helper) > 0 {
					netSpec = fmt.Sprintf("%s,helper=%s", netSpec, bridge.Helper)
				}

				qemuArgs = qemu.appendQemuArg(qemuArgs, "-netdev", netSpec)
			}
		}

		if cd.Net.Tap.Enabled {
			/*
				netSpec = fmt.Sprintf("virtio-net-pci,netdev=%s", cd.Net.Tap.ID)
				if len(cd.Net.Tap.MacAddress) > 0 {
					netSpec = fmt.Sprintf("%s,mac=%s", netSpec, cd.Net.Bridge.MacAddress)
				} else {
					macAddr, err := qemu.generateQemuMacAdress()
					if err == nil {
						netSpec = fmt.Sprintf("%s,mac=%s", netSpec, macAddr)
					} else {
						log.Printf("[qemuctl::qemu] error while generating mac address: %s\n", err.Error())
					}
				}

				qemuArgs = qemu.appendQemuArg(qemuArgs, "-device", netSpec)

				netSpec = fmt.Sprintf("tap,id=%s", cd.Net.Tap.ID)
				if len(cd.Net.Tap.TapInterface) > 0 {
					netSpec = fmt.Sprintf("%s,ifname=%s", netSpec, cd.Net.Tap.TapInterface)
				}

				if len(cd.Net.Tap.Bridge) > 0 {
					netSpec = fmt.Sprintf("%s,br=%s", netSpec, cd.Net.Tap.Bridge)
				}

				// TODO: make it wiser
				if cd.Net.Tap.Scripts.Enabled {
					netSpec = fmt.Sprintf("%s,script=%s,downscript=%s",
						netSpec, cd.Net.Tap.Scripts.UpScript, cd.Net.Tap.Scripts.DownScript)
				} else {
					netSpec = fmt.Sprintf("%s,script=no,downscript=no", netSpec)
				}
				qemuArgs = qemu.appendQemuArg(qemuArgs, "-netdev", netSpec)
			*/
		}
	}

	/*
	 * Disk specification
	 */
	// Check for 9P spec
	if len(cd.Disks.P9.Source) > 0 {
		if len(cd.Disks.P9.SecurityModel) == 0 {
			cd.Disks.P9.SecurityModel = "none"
		}

		p9Spec := fmt.Sprintf("local,path=%s,mount_tag=%s,security_model=%s",
			cd.Disks.P9.Source, cd.Disks.P9.Tag, cd.Disks.P9.SecurityModel)

		qemuArgs = qemu.appendQemuArg(qemuArgs, "-virtfs", p9Spec)
	}
	for _, blockDevice := range cd.Disks.BlockDevices {
		// TODO: Use stat to check whether it is a valid block device
		driveName := nodeDevices[0]
		if len(driveName) > 0 {
			nodeDevices = nodeDevices[1:]
		}

		// Appends drive/device specification
		qemuArgs = qemu.appendQemuArg(qemuArgs,
			"-blockdev",
			fmt.Sprintf("node-name=%s,driver=raw,file.driver=host_device,file.filename=%s", driveName, blockDevice))
	}

	// -- Disk images list
	for _, image := range cd.Disks.Images {
		driveMedia := "disk"
		driveIf := "ide"
		if len(image.Interface) > 0 {
			driveIf = image.Interface
		}

		if len(image.Media) > 0 {
			driveMedia = image.Media
		}
		qemuArgs = qemu.appendQemuArg(qemuArgs,
			"-drive",
			fmt.Sprintf("format=%s,file=%s,if=%s,media=%s", image.Format, image.File, driveIf, driveMedia))
	}
	// qemuArgs = append(qemuArgs, cd.Disks.HardDisk)

	if len(cd.Disks.Default) > 0 {
		qemuArgs = append(qemuArgs, cd.Disks.Default)
	}

	/* Add RTC (guest clock) spec */
	qemuArgs = qemu.appendQemuArg(qemuArgs, "-rtc", "base=utc,clock=host")

	/* Add a monitor specfication to be able to operate on the machine */
	qemuArgs = qemu.appendQemuArg(qemuArgs, "-chardev", monitor.GetChardevSpec())
	qemuArgs = qemu.appendQemuArg(qemuArgs, "-qmp", monitor.GetMonitorSpec())

	/* Add PIDfile spec */
	qemuArgs = qemu.appendQemuArg(qemuArgs, "-pidfile", monitor.GetPidFilePath())

	return qemuArgs, nil
}

func (qemu *QemuCommand) Launch() (processPid int, err error) {
	var procAttrs *os.ProcAttr = nil
	var procState *os.ProcessState = nil
	var procPid int
	var qemuArgs []string

	qemuArgs, err = qemu.getQemuArgs()
	if err != nil {
		return 0, err
	}

	// TODO: use the log feature; DONE
	log.Println("[QemuCommand::Launch] Executing QEMU with:")
	log.Printf("qemu_path ....... %s\n", qemu.QemuPath)
	log.Printf("qemu_args ....... %s\n", strings.Join(qemuArgs, " "))

	/* Actual execution of QEMU */
	err = nil

	log.Printf("[launch] creating qemu command struct")
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
	execArgs = append(execArgs, qemu.QemuPath)
	execArgs = append(execArgs, qemuArgs...)

	log.Printf("[launch] starting qemu process")
	qemuProcess, err := os.StartProcess(qemu.QemuPath, execArgs, procAttrs)
	if err != nil {
		log.Printf("[launch] error starting process: %s", err.Error())
		return 0, err
	}

	log.Printf("[launch] waiting for qemu process to finish")

	if !qemu.Configuration.RunAsDaemon {
		procState, err = qemuProcess.Wait()
		if err != nil {
			log.Printf("[launch] waiting for qemu command failed: %s (exit code: %d)",
				err.Error(), procState.ExitCode())

			qemuProcess.Kill()
			return 0, err
		}

		log.Printf("[launch] qemu process state: %s", procState.String())
		procPid = procState.Pid()
	} else {
		procPid, err = qemu.Monitor.WaitForPid()
	}

	return procPid, err
}
