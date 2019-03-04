package ssh

import (
	"path"
	"io/ioutil"

	"github.com/pkg/errors"

	"suse.com/caaspctl/pkg/caaspctl"
)

func init() {
	stateMap["cni.deploy"] = cniDeploy()
}

func cniDeploy() Runner {
	runner := struct{ State }{}
	runner.DoRun = func(t *Target, data interface{}) error {
		cniFiles, err := ioutil.ReadDir(caaspctl.CniDir())
		if err != nil {
			return errors.Wrap(err, "could not read local cni directory")
		}
		for _, f := range cniFiles {
			t.UploadFile(path.Join(caaspctl.CniDir(), f.Name()), path.Join("/tmp/cni.d", f.Name()))
		}
		t.ssh("kubectl --kubeconfig=/etc/kubernetes/admin.conf apply -f /tmp/cni.d")
		t.ssh("rm -rf /tmp/cni.d")
		return nil
	}
	return runner
}
