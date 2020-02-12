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
		expectedErrorMsg   string
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
			expectedErrorMsg: "auth failed: invalid input auth connector ID",
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
			expectedErrorMsg: "auth failed: no valid certificates found in root CA file",
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
			expectedErrorMsg: "read CA failed: open testdata/nonexist.crt: no such file or directory",
		},
		{
			name:  "auth failed",
			srvCb: startServer,
			cfg: LoginConfig{
				Username:           "admin",
				Password:           "1234",
				InsecureSkipVerify: true,
			},
			expectedErrorMsg: "auth failed: invalid username or password",
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
			expectedErrorMsg: "parse url: parse http://%41:8080/: invalid URL escape \"%41\"",
		},
		{
			name:  "oidc server with incorrect port number",
			srvCb: startServer,
			cfg: LoginConfig{
				DexServer:   "http://127.0.0.1:32001/",
				Username:    mockDefaultUsername,
				Password:    mockDefaultPassword,
				ClusterName: "test-cluster-name",
			},
			expectedErrorMsg: fmt.Sprintf("auth failed: failed to query provider http://127.0.0.1:32001/ (is this the right URL? maybe missing --root-ca or --insecure, or incorrect port number?)"),
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
			expectedErrorMsg: "auth failed: failed to parse provider scopes_supported: json: cannot unmarshal number into Go struct field .scopes_supported of type string",
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
			expectedErrorMsg: "auth failed: failed on get auth code url: Get ?access_type=offline&client_id=oidc-cli&redirect_uri=urn%3Aietf%3Awg%3Aoauth%3A2.0%3Aoob&response_type=code&scope=audience%3Aserver%3Aclient_id%3Aoidc: unsupported protocol scheme \"\"",
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
			expectedErrorMsg: "auth failed: failed to extract token from OOB response",
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

			if tt.expectedErrorMsg != "" {
				if err == nil {
					t.Errorf("error expected on %s, but no error reported", tt.name)
					return
				}
				if err.Error() != tt.expectedErrorMsg {
					t.Errorf("got error msg %s, want %s", err.Error(), tt.expectedErrorMsg)
					return
				}
				return
			}

			if err != nil {
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
