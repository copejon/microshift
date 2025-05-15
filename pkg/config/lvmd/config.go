package lvmd

import (
	"fmt"
	"os"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/kubernetes/pkg/apis/core"
	"k8s.io/kubernetes/pkg/apis/storage"
	"sigs.k8s.io/yaml"
)

var config *Lvmd

func NewFromConfig(lvmdconfigPath string) error {
	var l = new(Lvmd)
	buf, err := os.ReadFile(lvmdconfigPath)
	if err != nil {
		return fmt.Errorf("error to read lvmd file: %w", err)
	}
	err = yaml.Unmarshal(buf, &l)
	if err != nil {
		return fmt.Errorf("error unmarshalling lvmd file: %w", err)
	}
	if l.SocketName == "" {
		l.SocketName = defaultSockName
	}
	config = l
	return nil
}

func uint64Ptr(val uint64) *uint64 {
	return &val
}

// NewFromVolumeGroups returns a configuration struct for Lvmd with
// default settings based on the current host. If a single volume
// group is found, that value is used. If multiple volume groups are
// available and one is named "rhel", that group is used. Otherwise,
// the configuration returned will report that it is not enabled (see
// IsEnabled()).
func NewFromVolumeGroups() error {
	vgNames, err := getVolumeGroups()
	if err != nil {
		return fmt.Errorf("error discovering LVM volume groups: %w", err)
	}
	l := &Lvmd{SocketName: defaultSockName}
	defaultSet := false
	for _, vg := range vgNames {
		if vg == "microshift" && !defaultSet {
			defaultSet = true
		}
		dc := &DeviceClass{
			Name:    vg,
			Default: defaultSet,
			SpareGB: uint64Ptr(defaultSpareGB),
		}
		l.DeviceClasses = append(l.DeviceClasses, dc)
	}
	if !defaultSet {
		l.DeviceClasses[0].Default = true
	}
	config = l
	return nil
}

func Ready() bool {
	return config != nil && len(config.DeviceClasses) > 0
}

func Write() error {
	buf, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("error marshalling lvmd object: %w", err)
	}
	err = os.WriteFile("/var/lib/microshift/lvms/lvmd.yaml", buf, 0600)
	if err != nil {
		return fmt.Errorf("writing lvmd config: %w", err)
	}
	return nil
}

// GenerateStorageClassList takes a Lvmd object pointer and returns a list of storageClasses representing each device class
// StorageClass names are a concatenation of `topolvm-provision` and `DeviceClass.Name`. If
// lvmd.DeviceClasses[*].
func GenerateStorageClassList() []*storage.StorageClass {
	var storageClasses []*storage.StorageClass
	for _, dc := range config.DeviceClasses {
		sc := deviceClassToStorageClass(dc)
		storageClasses = append(storageClasses, sc)
	}
	return storageClasses
}

func deviceClassToStorageClass(dc *DeviceClass) *storage.StorageClass {
	reclaimPolicy := v1.PersistentVolumeReclaimDelete
	bindingMode := storage.VolumeBindingWaitForFirstConsumer
	allowVolExpansion := true

	return &storage.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: strings.Join([]string{"topolvm-provisioner", dc.Name}, "-"),
			Annotations: map[string]string{
				"storageclass.kubernetes.io/is-default-class": "false",
			},
		},
		Provisioner: "topolvm.io",
		Parameters: map[string]string{
			"csi.storage.k8s.io/fstype": "xfs",
		},
		ReclaimPolicy:        &reclaimPolicy,
		VolumeBindingMode:    &bindingMode,
		AllowVolumeExpansion: &allowVolExpansion,
	}
}
