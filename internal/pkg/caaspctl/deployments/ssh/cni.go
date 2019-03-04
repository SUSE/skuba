package ssh

import (
	"path"
	"io/ioutil"

	"github.com/pkg/errors"

	"suse.com/caaspctl/pkg/caaspctl"
	"suse.com/caaspctl/internal/pkg/caaspctl/deployments"
)

func init() {
	stateMap["cni.deploy"] = cniDeploy()
}

func cniDeploy() deployments.Runner {
	runner := struct{ deployments.State }{}
	runner.DoRun = func(t deployments.Target, data interface{}) error {
		cniFiles, err := ioutil.ReadDir(caaspctl.CniDir())
		if err != nil {
			return errors.Wrap(err, "could not read local cni directory")
		}
		for _, f := range cniFiles {
			t.UploadFile(path.Join(caaspctl.CniDir(), f.Name()), path.Join("/tmp/cni.d", f.Name()))
		}
		if target := sshTarget(t); target != nil {
			target.ssh("kubectl --kubeconfig=/etc/kubernetes/admin.conf apply -f /tmp/cni.d")
			target.ssh("rm -rf /tmp/cni.d")
		}
		return nil
	}
	return runner
}
