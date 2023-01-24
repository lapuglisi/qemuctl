package qemuctl_helpers

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v2"
)

// ConfigurationData holds the power of the serominers
type portForwards struct {
	GuestPort int `yaml:"guestPort"`
	HostPort  int `yaml:"hostPort"`
}

type ConfigurationData struct {
	Machine struct {
		EnableKVM   bool   `yaml:"enableKVM"`
		CPU         string `yaml:"cpuType"`
		MachineName string `yaml:"name"`
		MachineType string `yaml:"type"`
		AccelType   string `yaml:"accel"`
		TPM         struct {
			Enabled     bool `yaml:"enabled"`
			Passthrough struct {
				Enabled    bool   `yaml:"enabled"`
				ID         string `yaml:"id"`
				Path       string `yaml:"path"`
				CancelPath string `yaml:"cancelPath"`
			} `yaml:"passthrough"`
			Emulator struct {
				Enabled    bool   `yaml:"enabled"`
				ID         string `yaml:"id"`
				CharDevice string `yaml:"charDevice"`
			} `yaml:"emulator"`
		} `yaml:"tpm"`
	} `yaml:"machine"`
	RunAsDaemon bool   `yaml:"runAsDaemon"`
	Memory      string `yaml:"memory"`
	CPUs        int64  `yaml:"cpus"`
	Net         struct {
		DeviceType string `yaml:"deviceType"`
		User       struct {
			Enabled      bool           `yaml:"enabled"`
			ID           string         `yaml:"id"`
			IPSubnet     string         `yaml:"ipSubnet"`
			PortForwards []portForwards `yaml:"portForwards"`
		} `yaml:"user"`
		Bridge struct {
			Enabled    bool   `yaml:"enabled"`
			ID         string `yaml:"id"`
			Interface  string `yaml:"interface"`
			MacAddress string `yaml:"mac"`
			Helper     string `yaml:"helper"`
		}
		Tap struct {
			Enabled      bool   `yaml:"enabled"`
			ID           string `yaml:"id"`
			Bridge       string `yaml:"bridge"`
			TapInterface string `yaml:"tap"`
			MacAddress   string `yaml:"mac"`
			Scripts      struct {
				Enabled    bool   `yaml:"enabled"`
				UpScript   string `yaml:"upScript"`
				DownScript string `yaml:"upScript"`
			} `yaml:"scripts"`
		} `yaml:"tap"`
	} `yaml:"net"`
	SSH struct {
		LocalPort int `yaml:"localPort"`
	} `yaml:"ssh"`
	Disks struct {
		BlockDevice string `yaml:"blockDevice"`
		HardDisk    string `yaml:"hardDisk"`
		ISOCDrom    string `yaml:"cdrom"`
	} `yaml:"disks"`
	Display struct {
		EnableGraphics bool   `yaml:"enableGraphics"`
		VGAType        string `yaml:"vgaType"`
		DisplaySpec    string `yaml:"displaySpec"`
		VNC            struct {
			Enabled bool   `yaml:"enabled"`
			Listen  string `yaml:"listen"`
		} `yaml:"vnc"`
		Spice struct {
			Enabled          bool   `yaml:"enabled"`
			EnableIPv4       bool   `yaml:"ipv4"`
			EnableIPv6       bool   `yaml:"ipv6"`
			Port             int    `yaml:"port"`
			Address          string `yaml:"address"`
			TLSPort          int    `yaml:"tlsPort"`
			DisableTicketing bool   `yaml:"disableTicketing"`
			Password         string `yaml:"password"`
			EnableAgentMouse bool   `yaml:"enableAgentMouse"`
			OpenGL           bool   `yaml:"openGL"`
		} `yaml:"spice"`
	} `yaml:"display"`
	Audio struct {
		Enabled bool   `yaml:"enabled"`
		Driver  string `yaml:"driver"`
		Model   string `yaml:"model"`
	} `yaml:"enableSound"`
	Boot struct {
		KernelPath     string `yaml:"kernelPath"`
		RamdiskPath    string `yaml:"ramdiskPath"`
		KernelArgs     string `yaml:"kernelArgs"`
		BiosFile       string `yaml:"biosFile"`
		EnableBootMenu bool   `yaml:"enableBootMenu"`
		BootOrder      string `yaml:"bootOrder"`
	} `yaml:"boot"`
	QemuBinary string `yaml:"qemuBinary"`
}

// RuntimeConfiguration FTW
type RuntimeConfigurationMachine struct {
	Enabled bool   `yaml:"enabled"`
	Name    string `yaml:"name"`
}

type RuntimeConfiguration struct {
	Machines []RuntimeConfigurationMachine `yaml:"machines"`
}

// ConfigurationHandler is one hell of a seroclockers
type ConfigurationHandler struct {
	filePath string
}

func init() {
}

/* ConfigurationData implementation */
func NewConfigData() (configData *ConfigurationData) {
	configData = &ConfigurationData{}

	configData.Machine.MachineType = "q35"
	configData.Machine.AccelType = "hvm"
	configData.Machine.CPU = "host"
	configData.Machine.EnableKVM = true

	configData.Machine.TPM.Passthrough.Enabled = false
	configData.Machine.TPM.Emulator.Enabled = false

	configData.Net.DeviceType = "e1000"

	configData.Net.User.ID = "mynet0"

	configData.Net.Bridge.ID = "mybr0"

	configData.RunAsDaemon = false

	/* Display spec */
	configData.Display.EnableGraphics = true
	configData.Display.VGAType = "none"
	configData.Display.DisplaySpec = "none"

	configData.Display.VNC.Enabled = false

	configData.Display.Spice.Enabled = false
	configData.Display.Spice.EnableIPv4 = true
	configData.Display.Spice.EnableIPv6 = false

	return configData
}

/* ConfigurationHandler implementation */
func NewConfigHandler(configFile string) (configHandler *ConfigurationHandler) {
	return &ConfigurationHandler{
		filePath: configFile,
	}
}

func (ch *ConfigurationHandler) ParseConfigFile() (configData *ConfigurationData, err error) {
	var configBytes []byte = nil
	var bufReader *bufio.Reader = nil

	// Open file
	fileHandle, osErr := os.OpenFile(ch.filePath, os.O_RDONLY, 0644)
	if osErr != nil {
		err = fmt.Errorf("could not open file '%s': %s", ch.filePath, osErr.Error())
		return nil, err
	}
	defer fileHandle.Close()

	// Read lines
	bufReader = bufio.NewReader(fileHandle)

	configData = NewConfigData()
	osErr = nil

	configBytes, err = io.ReadAll(bufReader)
	if err != nil {
		return nil, err
	}

	/* Now YAML the whole thing */
	err = yaml.Unmarshal(configBytes, &configData)
	if err != nil {
		return nil, err
	}

	return configData, nil
}
