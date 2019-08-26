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
	"bytes"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func startServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", openIDHandler())
	mux.HandleFunc("/auth", authSingleConnectorHandler())
	mux.HandleFunc("/auth/local", authLocalHandler())
	mux.HandleFunc("/auth/ldap", authLocalHandler())
	mux.HandleFunc("/token", tokenHandler())
	mux.HandleFunc("/approval", approvalHandler())

	srv := httptest.NewUnstartedServer(mux)
	cert, _ := tls.LoadX509KeyPair("testdata/localhost.crt", "testdata/localhost.key")
	srv.TLS = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	srv.StartTLS()
	return srv
}

func Test_Login(t *testing.T) {
	tests := []struct {
		name               string
		srvCb              func() *httptest.Server
		cfg                LoginConfig
		expectedKubeConfCb func(string, string) *clientcmdapi.Config
		expectedError      bool
	}{
		{
			name:  "secure ssl/tls",
			srvCb: startServer,
			cfg: LoginConfig{
				Username:    mockDefaultUsername,
				Password:    mockDefaultPassword,
				RootCAPath:  "testdata/localhost.crt",
				ClusterName: "test-cluster-name",
			},
			expectedKubeConfCb: func(dexServerURL string, clusterName string) *clientcmdapi.Config {
				url, _ := url.Parse(dexServerURL)

				kubeConfig := clientcmdapi.NewConfig()
				kubeConfig.Clusters[clusterName] = &clientcmdapi.Cluster{
					Server:                   fmt.Sprintf("%s://%s:%s", defaultScheme, url.Hostname(), defaultAPIServerPort),
					CertificateAuthorityData: localhostCert,
				}
				kubeConfig.Contexts[clusterName] = &clientcmdapi.Context{
					Cluster:  clusterName,
					AuthInfo: mockDefaultUsername,
				}
				kubeConfig.CurrentContext = clusterName
				kubeConfig.AuthInfos[mockDefaultUsername] = &clientcmdapi.AuthInfo{
					AuthProvider: &clientcmdapi.AuthProviderConfig{
						Name: authProviderID,
						Config: map[string]string{
							"idp-issuer-url": dexServerURL,
							"client-id":      clientID,
							"client-secret":  clientSecret,
							"id-token":       mockIDToken,
							"refresh-token":  mockRefreshToken,
						},
					},
				}
				return kubeConfig
			},
		},
		{
			name:  "insecure ssl/tls",
			srvCb: startServer,
			cfg: LoginConfig{
				Username:           mockDefaultUsername,
				Password:           mockDefaultPassword,
				InsecureSkipVerify: true,
				ClusterName:        "test-cluster-name",
			},
			expectedKubeConfCb: func(dexServerURL string, clusterName string) *clientcmdapi.Config {
				url, _ := url.Parse(dexServerURL)

				kubeConfig := clientcmdapi.NewConfig()
				kubeConfig.Clusters[clusterName] = &clientcmdapi.Cluster{
					Server:                fmt.Sprintf("%s://%s:%s", defaultScheme, url.Hostname(), defaultAPIServerPort),
					InsecureSkipTLSVerify: true,
				}
				kubeConfig.Contexts[clusterName] = &clientcmdapi.Context{
					Cluster:  clusterName,
					AuthInfo: mockDefaultUsername,
				}
				kubeConfig.CurrentContext = clusterName
				kubeConfig.AuthInfos[mockDefaultUsername] = &clientcmdapi.AuthInfo{
					AuthProvider: &clientcmdapi.AuthProviderConfig{
						Name: authProviderID,
						Config: map[string]string{
							"idp-issuer-url": dexServerURL,
							"client-id":      clientID,
							"client-secret":  clientSecret,
							"id-token":       mockIDToken,
							"refresh-token":  mockRefreshToken,
						},
					},
				}
				return kubeConfig
			},
		},
		{
			name: "multiple connectors",
			srvCb: func() *httptest.Server {
				mux := http.NewServeMux()
				mux.HandleFunc("/.well-known/openid-configuration", openIDHandler())
				mux.HandleFunc("/auth", authMultipleConnectorsHandler())
				mux.HandleFunc("/auth/local", authLocalHandler())
				mux.HandleFunc("/auth/ldap", authLocalHandler())
				mux.HandleFunc("/token", tokenHandler())
				mux.HandleFunc("/approval", approvalHandler())
				return httptest.NewTLSServer(mux)
			},
			cfg: LoginConfig{
				Username:           mockDefaultUsername,
				Password:           mockDefaultPassword,
				ClusterName:        "test-cluster-name",
				InsecureSkipVerify: true,
				AuthConnector:      "ldap",
			},
			expectedKubeConfCb: func(dexServerURL string, clusterName string) *clientcmdapi.Config {
				url, _ := url.Parse(dexServerURL)

				kubeConfig := clientcmdapi.NewConfig()
				kubeConfig.Clusters[clusterName] = &clientcmdapi.Cluster{
					Server:                fmt.Sprintf("%s://%s:%s", defaultScheme, url.Hostname(), defaultAPIServerPort),
					InsecureSkipTLSVerify: true,
				}
				kubeConfig.Contexts[clusterName] = &clientcmdapi.Context{
					Cluster:  clusterName,
					AuthInfo: mockDefaultUsername,
				}
				kubeConfig.CurrentContext = clusterName
				kubeConfig.AuthInfos[mockDefaultUsername] = &clientcmdapi.AuthInfo{
					AuthProvider: &clientcmdapi.AuthProviderConfig{
						Name: authProviderID,
						Config: map[string]string{
							"idp-issuer-url": dexServerURL,
							"client-id":      clientID,
							"client-secret":  clientSecret,
							"id-token":       mockIDToken,
							"refresh-token":  mockRefreshToken,
						},
					},
				}
				return kubeConfig
			},
		},
		{
			name: "no matched connector",
			srvCb: func() *httptest.Server {
				mux := http.NewServeMux()
				mux.HandleFunc("/.well-known/openid-configuration", openIDHandler())
				mux.HandleFunc("/auth", authMultipleConnectorsHandler())
				return httptest.NewTLSServer(mux)
			},
			cfg: LoginConfig{
				Username:           mockDefaultUsername,
				Password:           mockDefaultPassword,
				ClusterName:        "test-cluster-name",
				InsecureSkipVerify: true,
				AuthConnector:      "ldap123",
			},
			expectedError: true,
		},
		{
			name:  "invalid root ca",
			srvCb: startServer,
			cfg: LoginConfig{
				Username:    mockDefaultUsername,
				Password:    mockDefaultPassword,
				RootCAPath:  "testdata/invalid.crt",
				ClusterName: "test-cluster-name",
			},
			expectedError: true,
		},
		{
			name:  "cert file not exist",
			srvCb: startServer,
			cfg: LoginConfig{
				Username:    mockDefaultUsername,
				Password:    mockDefaultPassword,
				RootCAPath:  "testdata/nonexist.crt",
				ClusterName: "test-cluster-name",
			},
			expectedError: true,
		},
		{
			name:  "auth failed",
			srvCb: startServer,
			cfg: LoginConfig{
				Username:           "admin",
				Password:           "1234",
				InsecureSkipVerify: true,
			},
			expectedError: true,
		},
		{
			name:  "invalid url",
			srvCb: startServer,
			cfg: LoginConfig{
				DexServer:          "http://%41:8080/",
				Username:           mockDefaultUsername,
				Password:           mockDefaultPassword,
				InsecureSkipVerify: true,
			},
			expectedError: true,
		},
		{
			name: "oidc server with http",
			srvCb: func() *httptest.Server {
				mux := http.NewServeMux()
				mux.HandleFunc("/.well-known/openid-configuration", openIDHandler())
				mux.HandleFunc("/auth", authSingleConnectorHandler())
				mux.HandleFunc("/auth/local", authLocalHandler())
				mux.HandleFunc("/token", tokenHandler())
				mux.HandleFunc("/approval", approvalHandler())
				return httptest.NewServer(mux)
			},
			cfg: LoginConfig{
				Username:    mockDefaultUsername,
				Password:    mockDefaultPassword,
				ClusterName: "test-cluster-name",
			},
			expectedError: true,
		},
		{
			name: "issuer scopes supported invalid",
			srvCb: func() *httptest.Server {
				mux := http.NewServeMux()
				mux.HandleFunc("/.well-known/openid-configuration", openIDHandlerInvalidScopes())
				return httptest.NewTLSServer(mux)
			},
			cfg: LoginConfig{
				Username:           mockDefaultUsername,
				Password:           mockDefaultPassword,
				InsecureSkipVerify: true,
				ClusterName:        "test-cluster-name",
			},
			expectedError: true,
		},
		{
			name: "issuer no claims",
			srvCb: func() *httptest.Server {
				mux := http.NewServeMux()
				mux.HandleFunc("/.well-known/openid-configuration", openIDHandlerNoScopes())
				return httptest.NewTLSServer(mux)
			},
			cfg: LoginConfig{
				Username:           mockDefaultUsername,
				Password:           mockDefaultPassword,
				InsecureSkipVerify: true,
				ClusterName:        "test-cluster-name",
			},
			expectedError: true,
		},
		{
			name: "approval body content incorrect",
			srvCb: func() *httptest.Server {
				mux := http.NewServeMux()
				mux.HandleFunc("/.well-known/openid-configuration", openIDHandler())
				mux.HandleFunc("/auth", authSingleConnectorHandler())
				mux.HandleFunc("/auth/local", authLocalHandler())
				mux.HandleFunc("/token", tokenHandler())
				mux.HandleFunc("/approval", approvalInvalidBodyHandler())
				return httptest.NewTLSServer(mux)
			},
			cfg: LoginConfig{
				Username:           mockDefaultUsername,
				Password:           mockDefaultPassword,
				InsecureSkipVerify: true,
				ClusterName:        "test-cluster-name",
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			testSrv := tt.srvCb()
			defer testSrv.Close()

			if tt.cfg.DexServer == "" {
				tt.cfg.DexServer = testSrv.URL
			}
			gotKubeConfig, err := Login(tt.cfg)

			if tt.expectedError {
				if err == nil {
					t.Errorf("error expected on %s, but no error reported", tt.name)
				}
				return
			} else if !tt.expectedError && err != nil {
				t.Errorf("error not expected on %s, but an error was reported (%v)", tt.name, err)
				return
			}

			expectKubeConfig := tt.expectedKubeConfCb(tt.cfg.DexServer, tt.cfg.ClusterName)
			if !reflect.DeepEqual(gotKubeConfig, expectKubeConfig) {
				t.Errorf("got %v, want %v", gotKubeConfig, expectKubeConfig)
				return
			}
		})
	}
}

func Test_LoginDebug(t *testing.T) {
	testServer := startServer()
	defer testServer.Close()

	// capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)

	_, err := doAuth(request{
		clientID:           clientID,
		clientSecret:       clientSecret,
		IssuerURL:          testServer.URL,
		Username:           mockDefaultUsername,
		Password:           mockDefaultPassword,
		InsecureSkipVerify: true,
		Debug:              true,
	})
	if err != nil {
		t.Errorf("error not expected, but an error was reported (%v)", err)
		return
	}

	if strings.Contains(buf.String(), mockDefaultPassword) {
		t.Error("password is not REDACTED")
	}
	if !strings.Contains(buf.String(), "REDACTED") {
		t.Error("password is not change to REDACTED")
	}
}

func Test_SaveKubeconfig(t *testing.T) {
	tests := []struct {
		name          string
		filename      string
		kubeConfig    *clientcmdapi.Config
		expectedError bool
	}{
		{
			name:       "success output",
			kubeConfig: clientcmdapi.NewConfig(),
		},
		{
			name:          "open file failed",
			filename:      "path/to/kubeconfig",
			kubeConfig:    clientcmdapi.NewConfig(),
			expectedError: true,
		},
		{
			name:          "encode failed",
			kubeConfig:    nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var path string
			if tt.filename != "" {
				path = filepath.Join("testdata", tt.filename+".golden")
			} else {
				path = filepath.Join("testdata", tt.name+".golden")
			}
			err := SaveKubeconfig(path, tt.kubeConfig)
			defer os.Remove(path)

			if tt.expectedError && err == nil {
				t.Errorf("error expected on %s, but no error reported", tt.name)
				return
			} else if !tt.expectedError && err != nil {
				t.Errorf("error not expected on %s, but an error was reported (%v)", tt.name, err)
				return
			}
		})
	}
}
