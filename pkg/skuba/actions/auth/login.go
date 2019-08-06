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

package auth

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"

	"github.com/pkg/errors"
	"k8s.io/client-go/tools/clientcmd/api"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	clientcmdlatest "k8s.io/client-go/tools/clientcmd/api/latest"
)

const (
	clientID     = "oidc-cli"
	clientSecret = "swac7qakes7AvucH8bRucucH"

	authProviderID = "oidc"
)

// LoginConfig represents the login configuration
type LoginConfig struct {
	DexServer          string
	RootCAPath         string
	InsecureSkipVerify bool
	ClusterName        string
	Username           string
	Password           string
	KubeConfigPath     string
	Debug              bool
}

// Login do authentication login process
func Login(cfg LoginConfig) (*api.Config, error) {
	var err error
	var apiserverURL string
	var rootCAData []byte = nil

	if !cfg.InsecureSkipVerify && cfg.RootCAPath != "" {
		rootCAData, err = ioutil.ReadFile(cfg.RootCAPath)
		if err != nil {
			return nil, errors.Wrapf(err, "read root CA failed: %s", err)
		}
	}

	authResp, err := doAuth(request{
		clientID:           clientID,
		clientSecret:       clientSecret,
		IssuerURL:          cfg.DexServer,
		Username:           cfg.Username,
		Password:           cfg.Password,
		RootCAData:         rootCAData,
		InsecureSkipVerify: cfg.InsecureSkipVerify,
		Debug:              cfg.Debug,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "auth failed")
	}

	url, err := url.Parse(cfg.DexServer)
	if err != nil {
		return nil, errors.Wrapf(err, "parse url")
	}
	apiserverURL = fmt.Sprintf("https://%s:6443", url.Hostname()) // Guess kube-apiserver on port 6443

	// fill out clusters
	kubeConfig := clientcmdapi.NewConfig()
	kubeConfig.Clusters[cfg.ClusterName] = &api.Cluster{
		Server:                   apiserverURL,
		InsecureSkipTLSVerify:    cfg.InsecureSkipVerify,
		CertificateAuthorityData: rootCAData,
	}

	// fill out contexts
	kubeConfig.Contexts[cfg.ClusterName] = &api.Context{
		Cluster:  cfg.ClusterName,
		AuthInfo: cfg.Username,
	}
	kubeConfig.CurrentContext = cfg.ClusterName

	// fill out auth infos
	kubeConfig.AuthInfos[cfg.Username] = &api.AuthInfo{
		AuthProvider: &api.AuthProviderConfig{
			Name: authProviderID,
			Config: map[string]string{
				"idp-issuer-url": cfg.DexServer,
				"client-id":      clientID,
				"client-secret":  clientSecret,
				"id-token":       authResp.IDToken,
				"refresh-token":  authResp.RefreshToken,
			},
		},
	}

	return kubeConfig, nil
}

// SaveKubeconfig saves kubeconfig to filename
func SaveKubeconfig(filename string, kubeConfig *clientcmdapi.Config) error {
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return errors.Wrapf(err, "open file")
	}
	defer f.Close()
	w := bufio.NewWriter(f)

	err = clientcmdlatest.Codec.Encode(kubeConfig, w)
	if err != nil {
		return errors.Wrapf(err, "encode kubeconfig")
	}

	w.Flush()
	return nil
}
