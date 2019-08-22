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
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	clientcmdlatest "k8s.io/client-go/tools/clientcmd/api/latest"
)

const (
	clientID     = "oidc-cli"
	clientSecret = "swac7qakes7AvucH8bRucucH"

	authProviderID = "oidc"
)

const (
	defaultScheme = "https"

	defaultAPIServerPort = "6443"
)

// LoginConfig represents the login configuration
type LoginConfig struct {
	DexServer          string
	Username           string
	Password           string
	RootCAPath         string
	InsecureSkipVerify bool
	AuthConnector      string
	ClusterName        string
	KubeConfigPath     string
	Debug              bool
}

// Login do authentication login process
func Login(cfg LoginConfig) (*clientcmdapi.Config, error) {
	var err error
	var rootCAData []byte

	if !cfg.InsecureSkipVerify && cfg.RootCAPath != "" {
		rootCAData, err = ioutil.ReadFile(cfg.RootCAPath)
		if err != nil {
			return nil, errors.Wrapf(err, "read root CA failed: %s", err)
		}
	}

	url, err := url.Parse(cfg.DexServer)
	if err != nil {
		return nil, errors.Wrapf(err, "parse url")
	}

	authResp, err := doAuth(request{
		clientID:           clientID,
		clientSecret:       clientSecret,
		IssuerURL:          cfg.DexServer,
		Username:           cfg.Username,
		Password:           cfg.Password,
		RootCAData:         rootCAData,
		InsecureSkipVerify: cfg.InsecureSkipVerify,
		AuthConnector:      cfg.AuthConnector,
		Debug:              cfg.Debug,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "auth failed")
	}

	// fill out clusters
	kubeConfig := clientcmdapi.NewConfig()
	kubeConfig.Clusters[cfg.ClusterName] = &clientcmdapi.Cluster{
		Server:                   fmt.Sprintf("%s://%s:%s", defaultScheme, url.Hostname(), defaultAPIServerPort), // Guess kube-apiserver on port 6443
		InsecureSkipTLSVerify:    cfg.InsecureSkipVerify,
		CertificateAuthorityData: rootCAData,
	}

	// fill out contexts
	kubeConfig.Contexts[cfg.ClusterName] = &clientcmdapi.Context{
		Cluster:  cfg.ClusterName,
		AuthInfo: cfg.Username,
	}
	kubeConfig.CurrentContext = cfg.ClusterName

	// fill out auth infos
	kubeConfig.AuthInfos[cfg.Username] = &clientcmdapi.AuthInfo{
		AuthProvider: &clientcmdapi.AuthProviderConfig{
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
