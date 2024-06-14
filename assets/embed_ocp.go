//go:build ocp
package embedded

import (
	"embed"
	"io/fs"
)

//go:embed controllers core crd version release components/csi-snapshot-controller  components/openshift-dns components/openshift-router components/ovn  components/service-ca
//go:embed components/lvms
var content embed.FS

func Asset(name string) ([]byte, error) {
	return content.ReadFile(name)
}

func AssetStreamed(name string) (fs.File, error) {
	return content.Open(name)
}

func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}
