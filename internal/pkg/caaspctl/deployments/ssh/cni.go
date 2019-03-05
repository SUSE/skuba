package ssh

import (
	"io/ioutil"
	"path"

	"github.com/pkg/errors"

	"suse.com/caaspctl/pkg/caaspctl"
)

func init() {
	stateMap["cni.deploy"] = cniDeploy()
}

func cniDeploy() Runner {
	return func(t *Target, data interface{}) error {
		cniFiles, err := ioutil.ReadDir(caaspctl.CniDir())
		if err != nil {
			return errors.Wrap(err, "could not read local cni directory")
		}

		defer t.ssh("rm -rf /tmp/cni.d")

		for _, f := range cniFiles {
			if err := t.UploadFile(path.Join(caaspctl.CniDir(), f.Name()), path.Join("/tmp/cni.d", f.Name())); err != nil {
				return err
			}
		}

		_, _, err = t.ssh("kubectl --kubeconfig=/etc/kubernetes/admin.conf apply -f /tmp/cni.d")
		return err
	}
}
