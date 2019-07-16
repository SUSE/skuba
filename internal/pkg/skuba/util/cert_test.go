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

package util

import (
	"crypto"
	"crypto/x509"
	"net"
	"testing"

	"k8s.io/client-go/kubernetes/fake"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/pkiutil"
)

func Test_NewServerCertAndKey(t *testing.T) {
	// generate test root CA
	caCert, caKey, err := pkiutil.NewCertificateAuthority(&certutil.Config{
		CommonName:   "unit-test",
		Organization: []string{"suse.com"},
		AltNames: certutil.AltNames{
			DNSNames: []string{"unit.test"},
		},
	})
	if err != nil {
		t.Errorf("generate root CA failed: %v", err)
	}

	tests := []struct {
		name          string
		caCert        *x509.Certificate
		caKey         crypto.Signer
		commonName    string
		certSANs      []string
		expectedError bool
	}{
		{
			name:       "control plane is IP address",
			caCert:     caCert,
			caKey:      caKey,
			commonName: "cert-unit-test",
			certSANs: []string{
				"10.20.30.40",
				"20.30.40.50",
			},
		},
		{
			name:       "control plane is FQDN",
			caCert:     caCert,
			caKey:      caKey,
			commonName: "cert-unit-test",
			certSANs: []string{
				"dex.unittest",
				"dex.unit.test",
			},
		},
		{
			name:       "control plane is both IP address and FQDN",
			caCert:     caCert,
			caKey:      caKey,
			commonName: "cert-unit-test",
			certSANs: []string{
				"10.20.30.40",
				"dex.unit.test",
			},
		},
		{
			name:       "invalid input ca cert",
			caCert:     nil,
			caKey:      caKey,
			commonName: "cert-unit-test",
			certSANs: []string{
				"10.20.30.40",
				"dex.unit.test",
			},
			expectedError: true,
		},
		{
			name:       "invalid input ca key",
			caCert:     caCert,
			caKey:      nil,
			commonName: "cert-unit-test",
			certSANs: []string{
				"10.20.30.40",
				"dex.unit.test",
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			cert, _, err := NewServerCertAndKey(tt.caCert, tt.caKey, tt.commonName, tt.certSANs)
			if tt.expectedError {
				if err == nil {
					t.Errorf("error expected on %s, but no error reported", tt.name)
				}
				return
			} else if !tt.expectedError && err != nil {
				t.Errorf("error not expected on %s, but an error was reported (%v)", tt.name, err)
				return
			}

			for _, dns := range cert.DNSNames {
				if net.ParseIP(dns) != nil {
					t.Error("invalid DNS name")
				}
				if !strSliceContains(tt.certSANs, dns) {
					t.Errorf("generate SAN %s is not expected", dns)
				}
			}

			// check generated SANs IP address
			for _, ip := range cert.IPAddresses {
				if !strSliceContains(tt.certSANs, ip.String()) {
					t.Errorf("generate SAN %v is not expected", ip)
				}
			}
		})
	}
}

func Test_CreateOrUpdateCertToSecret(t *testing.T) {
	// generate test root CA
	caCert, caKey, err := pkiutil.NewCertificateAuthority(&certutil.Config{
		CommonName:   "unit-test",
		Organization: []string{"suse.com"},
		AltNames: certutil.AltNames{
			DNSNames: []string{"unit.test"},
		},
	})
	if err != nil {
		t.Errorf("generate root CA failed: %v", err)
	}

	tests := []struct {
		name          string
		caCert        *x509.Certificate
		caKey         crypto.Signer
		commonName    string
		certSANs      []string
		secretName    string
		expectedError bool
	}{
		{
			name:       "create/update secret",
			caCert:     caCert,
			caKey:      caKey,
			commonName: "cert-unit-test",
			certSANs:   []string{"1.1.1.1", "dex.unit.test"},
			secretName: "cert-unit-test-secret-name",
		},
		{
			name:          "invalid input ca cert",
			caCert:        nil,
			caKey:         caKey,
			commonName:    "cert-unit-test",
			certSANs:      []string{"1.1.1.1", "dex.unit.test"},
			secretName:    "cert-unit-test-secret-name",
			expectedError: true,
		},
		{
			name:          "invalid input ca key",
			caCert:        caCert,
			caKey:         nil,
			commonName:    "cert-unit-test",
			certSANs:      []string{"1.1.1.1", "dex.unit.test"},
			secretName:    "cert-unit-test-secret-name",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// generate server certificate
			cert, key, err := NewServerCertAndKey(tt.caCert, tt.caKey, tt.commonName, tt.certSANs)
			if tt.expectedError && err == nil {
				t.Errorf("error expected on %s, but no error reported", tt.name)
				return
			} else if !tt.expectedError && err != nil {
				t.Errorf("error not expected on %s, but an error was reported (%v)", tt.name, err)
				return
			}

			// test create/update certificate to secret
			err = CreateOrUpdateCertToSecret(fake.NewSimpleClientset(), caCert, cert, key, tt.secretName)
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
