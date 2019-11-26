/*
 * Copyright (c) 2019 SUSE LLC.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package cluster

import (
	"os"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	kubectlget "k8s.io/kubernetes/pkg/kubectl/cmd/get"
)

// Status prints the status of the cluster on the standard output by reading the
// admin configuration file from the current folder
func Status(client clientset.Interface) error {
	nodeList, err := client.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "could not retrieve node list")
	}

	outputFormat := "custom-columns=NAME:.metadata.name,OS-IMAGE:.status.nodeInfo.osImage,KERNEL-VERSION:.status.nodeInfo.kernelVersion,KUBELET-VERSION:.status.nodeInfo.kubeletVersion,CONTAINER-RUNTIME:.status.nodeInfo.containerRuntimeVersion,HAS-UPDATES:.metadata.annotations.caasp\\.suse\\.com/has-updates,HAS-DISRUPTIVE-UPDATES:.metadata.annotations.caasp\\.suse\\.com/has-disruptive-updates"

	printFlags := kubectlget.NewGetPrintFlags()
	printFlags.OutputFormat = &outputFormat

	printer, err := printFlags.ToPrinter()
	if err != nil {
		return errors.Wrap(err, "could not create printer")
	}
	if err := printer.PrintObj(nodeList, os.Stdout); err != nil {
		return errors.Wrap(err, "could not print to stdout")
	}
	return nil
}
