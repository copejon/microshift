package lvmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

const (
	RuntimeLvmdConfigFile       = "/var/lib/microshift/lvms/lvmd.yaml"
	LvmdConfigFileName          = "lvmd.yaml"
	defaultSockName             = "/run/lvmd/lvmd.socket"
	defaultRHEL4EdgeVolumeGroup = "microshift"
)

// TODO update for accuracy
// lvmd stores the read-in or defaulted values of the lvmd configuration and provides the topolvm-node process information
// about its host's storage environment.
type lvmd struct {
	DeviceClasses []*DeviceClass `json:"device-classes"`
	SocketName    string         `json:"socket-name"`
	Message       string         `json:"-"` //TODO remove this
}

// IsEnabled returns a boolean indicating whether the CSI driver
// should be enabled for this host.
func (l *lvmd) IsEnabled() bool {
	return len(l.DeviceClasses) > 0
}

func getLvmdConfigForVGs(vgNames []string) (*lvmd, error) {
	response := &lvmd{
		SocketName: defaultSockName,
	}
	vgName := ""
	if len(vgNames) == 0 {
		response.Message = "No volume groups found"
		klog.V(2).Info("No volume groups found")
		return response, nil
	} else if len(vgNames) == 1 {
		vgName = vgNames[0]
		klog.V(2).Infof("Using volume group %q", vgName)
		response.Message = "Defaulting to the only available volume group"
	} else {
		for _, name := range vgNames {
			if name == defaultRHEL4EdgeVolumeGroup {
				klog.V(2).Infof("Using default volume group %q", defaultRHEL4EdgeVolumeGroup)
				vgName = name
				response.Message = "Found default volume group \"microshift\""
				break
			}
		}

		// If the default volume group is not found and there are
		// multiple volume groups, disable the CSI driver.
		if vgName == "" {
			klog.V(2).Infof("Multiple volume groups available but no configuration file is present, disabling CSI. %v", vgNames)
			response.Message = "Multiple volume groups are available, but no configuration file was provided."
			return response, nil
		}
	}

	// Fill in the default device class using the selected volume
	// group.
	response.DeviceClasses = []*DeviceClass{
		{
			Name:        "default",
			VolumeGroup: vgName,
			Default:     true,
			SpareGB:     uint64Ptr(defaultSpareGB),
		},
	}
	return response, nil
}

// DefaultLvmdConfig returns a configuration struct for lvmd with
// default settings based on the current host. If a single volume
// group is found, that value is used. If multiple volume groups are
// available and one is named "rhel", that group is used. Otherwise,
// the configuration returned will report that it is not enabled (see
// IsEnabled()).
func DefaultLvmdConfig() (*lvmd, error) {
	vgNames, err := getVolumeGroups()
	if err != nil {
		return nil, fmt.Errorf("failed to discover local volume groups: %w", err)
	}
	return getLvmdConfigForVGs(vgNames)
}

// getVolumeGroups returns a slice of volume group names.
func getVolumeGroups() ([]string, error) {
	cmd := exec.Command("vgs", "--readonly", "--options=name", "--noheadings")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error running vgs: %w", err)
	}
	names := []string{}
	for _, line := range strings.Split(string(output), "\n") {
		newName := strings.Trim(line, " \t\n")
		if newName != "" {
			names = append(names, newName)
		}
	}
	return names, nil
}

func NewLvmdConfigFromFile(p string) (*lvmd, error) {
	l := new(lvmd)
	buf, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("failed to read lvmd file: %w", err)
	}

	err = yaml.Unmarshal(buf, &l)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling lvmd file: %w", err)
	}
	if l.SocketName == "" {
		l.SocketName = defaultSockName
	}
	l.Message = fmt.Sprintf("Read from %s", p)
	return l, nil
}

func SaveLvmdConfigToFile(l *lvmd, p string) error {
	buf, err := yaml.Marshal(l)
	if err != nil {
		return fmt.Errorf("marshalling lvmd config: %w", err)
	}
	err = os.WriteFile(p, buf, 0600)
	if err != nil {
		return fmt.Errorf("writing lvmd config: %w", err)
	}
	return nil
}

func LvmPresentOnMachine() error {
	if _, err := exec.LookPath("lvm"); err != nil {
		return fmt.Errorf("failed to find 'vgs' command line tool: %w", err)
	}
	return nil
}
