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

package certutil

import (
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"net"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/pkiutil"
)

const (
	caPEM = `
-----BEGIN CERTIFICATE-----
MIICyDCCAbCgAwIBAgIBADANBgkqhkiG9w0BAQsFADAVMRMwEQYDVQQDEwprdWJl
cm5ldGVzMB4XDTE5MTIwNDAyMjU1OFoXDTI5MTIwMTAyMjU1OFowFTETMBEGA1UE
AxMKa3ViZXJuZXRlczCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBANZf
n+jOJ2UdlHreWHl08TrCJVCY9sF64wflpLCG6A90T/3i+/63ksi/AlIGGPXr3jej
F2qy0y33XKy6FQu6nzP3BzYHi6Df5078ykxJNclCmF0a7blrwST+zqs3bYy+Xi5p
xqExWpOaqGNWVjE+wReUKu0dRxK2DOYERyyJRRjc0ZRixXT/ouGifjAg/pd6q8TJ
/owle4N/OVH8tkftf2Kvn7T3Xzh3xtaxs9Iza4MAIInJpQyfrrVrxpcgV5EdypG+
Wat4MXzswFrEbaR/KOTic3XweFk6XnWkuGEWM3TAcPGiEq1z5YPrNn8cgBUa+0vU
elGj/oRBNmBiozcEq4UCAwEAAaMjMCEwDgYDVR0PAQH/BAQDAgKkMA8GA1UdEwEB
/wQFMAMBAf8wDQYJKoZIhvcNAQELBQADggEBAEuHDKNMZeuXsEYjNCUVuttnCRz3
lwrZZEsUlIYHQEchjJECCSkkmc3dfxDmlzgBldMLewneiB9iGlFyw1ihWj78tVL6
mJ0oeBlhbb9PXtjBaaG9Ewezrh/bT5ciR6O88+BusVwdgebpHw4SpEi9LN+j0PIX
0mAqbMFPpHQyVbRt1hfKNyGqSSrjti2UPDDEYKHmiqXqM4YGfPjfNJImZn5I+fXp
yHempNdFVxADD+8W76uu93bbufXm+QnBmIhWtpTqqGgEvkaWgvV2wUtSMWRSSXHY
QzwWV3TxhgPGPBzRDZ+hb9zQW+lvR+tSEDxPUNoGIPIXU4AL4N9jGup++EI=
-----END CERTIFICATE-----`

	caKeyPEM = `
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA1l+f6M4nZR2Uet5YeXTxOsIlUJj2wXrjB+WksIboD3RP/eL7
/reSyL8CUgYY9eveN6MXarLTLfdcrLoVC7qfM/cHNgeLoN/nTvzKTEk1yUKYXRrt
uWvBJP7OqzdtjL5eLmnGoTFak5qoY1ZWMT7BF5Qq7R1HErYM5gRHLIlFGNzRlGLF
dP+i4aJ+MCD+l3qrxMn+jCV7g385Ufy2R+1/Yq+ftPdfOHfG1rGz0jNrgwAgicml
DJ+utWvGlyBXkR3Kkb5Zq3gxfOzAWsRtpH8o5OJzdfB4WTpedaS4YRYzdMBw8aIS
rXPlg+s2fxyAFRr7S9R6UaP+hEE2YGKjNwSrhQIDAQABAoIBAAwm4Yat4PfPZHJO
lk9UPLRq+viFoz82exYgg2RqUU9G9Z3btxMqTszIXxZNOC8AjtkyiopG1se9ROiZ
p8XBb3Lfpu3+IYEeEBufIsyOPdlJyB9G/oDLReiV9RspijE3PVl+L39Fr++8DZ2L
8FjcSM/QW1qTlUrPPQ3w4iP5KAyPpUTXzKkeJZx47wyVdDRDv+UDTRcV1d7I17JN
n5yeL3vwhxHuKAJF12kOeTW4hKMyVqckjd8CMgiP495QIVQcrrDl7TQJmgsKGAQd
IUjgKDLdKy0AvMDst/Lnyst2k+CsUJF3eddcSBRyUMs1jozd1rLPkZkY1uXqWCJI
ai9rMkECgYEA2fKt0Fc3Oow/DKzRGg5GgPIeqH7zqZZDjk84vhG37yZGnQPMmnpw
jZg/Y1QS2TVP/74ToGgCrPu5lKb+3MGy60giQBdXbutbTFU+1AY9uBybPBI36Tfi
CIcFmfKXx6j32KrhFEDQXYuWQBFfRJfb+YGuXFMViU2gNP/WOhQvE3ECgYEA+80v
KzdclRUO2ujIFghYeMa8AVWxYAuBGJE/vMt7XK7MuO1PS4AWrmXNU3FChPRh7QLk
vPAiRoObbNeHspxJHramxiGc/7dj5mXHYam7rnSHBi0U2JyF9nLyPqnv9gMJNzKP
bSx5rCqb2rDVTGlxNAyPPtGHOQDLLL9C/U29J1UCgYAdAJcynziBOQJ23FRjBD1L
kWyU/XfNPGq2+EHTwSXZ1B0Xbdb/Q4XQwc7Fl/1+HAMORCv2b4DTphe2+VX26Gu3
tXyhTLncz8LxcHKQ4le6NUxO/RmllkMk4VrUdpzN++UnVu3mtQ1FNXsEAYvM4+xo
0mHydTfrcoH8K4NFbUQqcQKBgQD4vURvSI0oqFi4X2Pof+4FwSxPlTtXSYYJotJ3
yfrfH74UoDjIuIuvU9l1KFkxxchGvakAC6eQSMnsxyzBgCmrMXumFeZlpeAF5V0E
WQuR1oLb0wTYxiZ/wiUTSgRF3dHouQV+L4UyUhUL/8t1ZGPzqsSGpa0S3nnWhknC
uFy20QKBgQClMenL1HjNG5F5ZWSsSsG4iZVrGq/OxUk2MUM0mviTKz4F2dHn7F7x
TjGRBWJRgs+LfMVQF3HZ86AgWq1teRBZzb60ClcjUmqdFAdEL9deGHpSTrYTd37E
rgIozdRJx1v+1zU3Ht6Fxd2KRVo6LhH7xV6H5epctRYGthRG8m94tw==
-----END RSA PRIVATE KEY-----
`

	certPEM = `
-----BEGIN CERTIFICATE-----
MIIC+zCCAeOgAwIBAgIICdwVmK/rtykwDQYJKoZIhvcNAQELBQAwFTETMBEGA1UE
AxMKa3ViZXJuZXRlczAeFw0xOTEyMDQwMjI1NThaFw0yMDEyMDMwMjI2NTZaMCwx
FzAVBgNVBAoTDnN5c3RlbTptYXN0ZXJzMREwDwYDVQQDEwhvaWRjLWRleDCCASIw
DQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAKwtgunogXnCuJX2CjkwcB+NGd3l
CWzGVu3eVICbGKGLZf7kxFDcm+aX/Kt6yfhtP2zRrmqll0ytoQrGfgIo1L7S6uDd
l1p2Ue2kq7CPEVq+BD8rdVLS/4Mum6cr7+3407q6+h7slFNudxXsU4FSDLQMboEb
9c8uGS6vLX5z4dSWFSWnhNedY1Da0s5jgv+5ujyzoytnEaEJNaO5rGCufH5A8yzn
jjapCe8w/EQ35UQDvkC6L4mcPT6nxUfKnl36Ri3BJAxamdEsr+ppNcNRaGktioh3
uEIjjTqza97gI6FBZfn6jrRV/xhr0L3iuTTHc+gW0m5B4QCJT8lnEda4DOECAwEA
AaM4MDYwDgYDVR0PAQH/BAQDAgWgMBMGA1UdJQQMMAoGCCsGAQUFBwMBMA8GA1Ud
EQQIMAaHBApWAmMwDQYJKoZIhvcNAQELBQADggEBADBU5yAoSnh6SozMVtpnrb8g
i+56K0NMPFz5IG5uoQd7Ghj0d+E0t8qMrZMi62zzS1r5QxkVy3Rb54iDQZ3YDwK3
CtRFN42H46QTsKbX0JJ/m53LHoNjFYAO1QW+XmYnV0C+ogtU51GOZyVfkwHc5QoI
u6WuquN7WBEhB4dOqjVvjMsL5C9U9jhQU3FjM0g10gE8Z+DhPPIBsWh0zvRP4kLK
OBRY4KzDxcgoL7nHbrvZ5TRmXeBo1T8UPpHcJYM12gkRBg8r2V8xgMq2Csk966Hs
upAYY7lz4OQioShWYBcNIdCpO71GfVFi5lX1KYqCzeuKE7qz8SkDijM2fb6fBqc=
-----END CERTIFICATE-----`

	fakeCAPEM = `
-----BEGIN CERTIFICATE-----
MIICyDCCAbCgAwIBAgIBADANBgkqhkiG9w0BAQsFADAVMRMwEQYDVQQDEwprdWJl
cm5ldGVzMB4XDTE5MTIwMzAyNTY1NVoXDTI5MTEzMDAyNTY1NVowFTETMBEGA1UE
AxMKa3ViZXJuZXRlczCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBALvn
oWVWWJV3JFGQw/vvIFhP69EmbRWojnOa8qNj2M0x82mqz1b1+jlAPT9HlFU56OUZ
m8hINybqgAqYqZe1X+9h96YeWGRMWy7rQXDbSZNuqukOZ+/yDMGzhIWy9d9eXsGk
QDDFhKrp7OvNSEhco5IOw/FJzNiOU6zd4+6jt+8mzN6Gi6OyWl5q0QnWI/DrcnHB
+TJZnwcfmCrbztrFiC3sF6JRtHq9ujA8qTg99Gdu6JuVm7m/xDn3uLot6WW0wGE+
3o/mGBKVAlOP+zk0c0Y+hiWDgd2ileYWDUFwoDp97ZkHMJe5SO0hRJcNTnliQttU
yFGmIzN4VXcLpXjn0zsCAwEAAaMjMCEwDgYDVR0PAQH/BAQDAgKkMA8GA1UdEwEB
/wQFMAMBAf8wDQYJKoZIhvcNAQELBQADggEBAK0pF2nUyUMgZpf83HSA4Kr8R58B
BiJCPJnWB6cbJIBcHcexi8XzNYiWkGZ4mUR2lyIiqZ4k5myGfmhqXMCRdKpq6MU2
hDpGBpVNKeF+y0qRS7z6cjyvewp3zYznHrw6RxdVZXoIZb57zXjVyUNpMmD3XR4k
CSfyJ3P2qVsxcObcM8GqAtdQxpr7VOAgMuNYN2cm1H3uNllWhI+2QVHF+R1ljL/7
h8VyVk8qkrs0jhz2Ej71G8c3IYeMtcjEZ9qcSOH8XakpYdQTCX4q3KKMSosqTwLs
Gr2nQJCKDe6ftV7GAl8qp4QhV731tpXx4RrW47/5omi97BZ0LsBv9FlLRSU=
-----END CERTIFICATE-----`

	fakeCAKeyPEM = `
-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAu+ehZVZYlXckUZDD++8gWE/r0SZtFaiOc5ryo2PYzTHzaarP
VvX6OUA9P0eUVTno5RmbyEg3JuqACpipl7Vf72H3ph5YZExbLutBcNtJk26q6Q5n
7/IMwbOEhbL1315ewaRAMMWEquns681ISFyjkg7D8UnM2I5TrN3j7qO37ybM3oaL
o7JaXmrRCdYj8OtyccH5MlmfBx+YKtvO2sWILewXolG0er26MDypOD30Z27om5Wb
ub/EOfe4ui3pZbTAYT7ej+YYEpUCU4/7OTRzRj6GJYOB3aKV5hYNQXCgOn3tmQcw
l7lI7SFElw1OeWJC21TIUaYjM3hVdwuleOfTOwIDAQABAoIBAFI/SwfeSZvysHT7
Vq2Zt6CwKto7ZZgLVX8InZgjBiya5p6j42l+9W3Fzok6PZUoaeaN1QBPi8R+9Fiv
BdyfyUQwr4OI2MveGDNrShOqCIR99lVYtunyGt9WQnV7JeAFoJhF2sr+Sdm91rRI
AJGb6wTtbZrZ4M4RTlLmNPSpuML2iVNOQN3NJNAW5SxrZXnQuGFVagnPZpiXJlPv
QGF6saNw6lDUmnnMZmnKSt02ktDdam/xBX0D0X4yp1UAiy75rnmfbj4tJDS8o/2e
NWlrogBcZBaP/n0lHj1Pri9IYjIjZvvwgiPeULMFN4w812oB4S5sGjtqpmGnEGSN
imx0aTECgYEA+BhcNIUvoL/sB65wsRJG5WCili083s7sF/QeK7x03I3vKHIZqYDr
+0FckBLlp+iDdIiyVKmwCz2BERfiMxG+87h664iRkphxAp2RCroRzzvUEdsKssAQ
qSIMuYn3jisBqqp6H5jwTweCEZHUZYnPTRXUP6ylqsKctozUOuPMXxcCgYEAweRQ
rVy/flVfc4ew5RtOzjIAUXtc90OgXjiC8uBjq0RLs5zRgoVi3pgydn9gnazpvwZq
pSjJpSgMTNwkw+o35ac7NragBE8i3jBzv4qshmmP8yYmTeH6RbNd4Sgpr3YLydqe
bqSA71GbCRlsWiifupnsMIyKp6Ef/2gsZSgU430CgYEApDILJD9ZbDxZDCRpNOfx
v/Ga6WV7OcMdAiVwqmWJuka9l7kcPtCyXZG+nyPClsQN7FxkGiBMAMRt3VA/Rqli
BY982tGB9tGpSZ/a1IydKNhh3Idppy/yVt3QKiOjkZXo/njhZnQj50oCzXoEZkc0
ycG+vX2YD1HJwg+mjmshYXUCgYBS9OSvx+cGnnBgdcXxwGVPQ4VvV2DHSl/q8DLW
x7rdJDNffdEGDxvmMSgmGwmzbK/100D9uR3NU/0vRWFVkXipAYwMNMbyEQnSFtjv
Mt3uBGxalA//cpgqCjw4gX6UW+VfT/JJVIj12+yBUCdTy93LcN/lRbxtTDrshB26
ihOl4QKBgFs6VLvi38XHoZ94SRACdMfqqetUnaNWOpBA5/pcrlO37cPuIjeoLyOO
diIum+sOpXZCCCEUExQyTgvuM1PmbNeErwZpRRXXjn8e2CePA8ZFq6yxHagvbvkU
hpFIj0VGTCEWiL0DrB1BacBkjcWwvVeLVwCLr4A1YxoKvx7p418Y
-----END RSA PRIVATE KEY-----`

	invalidPEMWithBlock = `
-----BEGIN CERTIFICATE-----
-----END CERTIFICATE-----`

	invalidPEMWithoutBlock = ``
)

func TestCreateOrUpdateServerCertAndKeyToSecret(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		caPEM          string
		caKeyPEM       string
		certPEM        string
		secretDataKey  string
		certCommonName string
		certSANs       []string
		expectedError  bool
		expectedUpdate bool
	}{
		{
			name:           "server cert not exist in kube-system namespace",
			namespace:      metav1.NamespaceDefault,
			caPEM:          caPEM,
			caKeyPEM:       caKeyPEM,
			certPEM:        certPEM,
			secretDataKey:  corev1.TLSCertKey,
			certCommonName: "test-cn",
			certSANs:       []string{"192.168.1.1"},
		},
		{
			name:           "server cert exist and signed by the same ca",
			namespace:      metav1.NamespaceSystem,
			caPEM:          caPEM,
			caKeyPEM:       caKeyPEM,
			certPEM:        certPEM,
			secretDataKey:  corev1.TLSCertKey,
			certCommonName: "test-cn",
			certSANs:       []string{"192.168.1.1"},
			expectedUpdate: true,
		},
		{
			name:           "server cert exist but not signed by the same ca",
			namespace:      metav1.NamespaceSystem,
			caPEM:          fakeCAPEM,
			caKeyPEM:       fakeCAKeyPEM,
			certPEM:        certPEM,
			secretDataKey:  corev1.TLSCertKey,
			certCommonName: "test-cn",
			certSANs:       []string{"192.168.1.1"},
		},
		{
			name:           "server cert exist but not in correct secret data key",
			namespace:      metav1.NamespaceSystem,
			caPEM:          fakeCAPEM,
			caKeyPEM:       fakeCAKeyPEM,
			certPEM:        certPEM,
			secretDataKey:  corev1.TLSPrivateKeyKey,
			certCommonName: "test-cn",
			certSANs:       []string{"192.168.1.1"},
		},
		{
			name:           "server cert exist but no block",
			namespace:      metav1.NamespaceSystem,
			caPEM:          caPEM,
			caKeyPEM:       caKeyPEM,
			certPEM:        invalidPEMWithBlock,
			secretDataKey:  corev1.TLSCertKey,
			certCommonName: "test-cn",
			certSANs:       []string{"192.168.1.1"},
			expectedError:  true,
		},
		{
			name:           "server cert exist but empty",
			namespace:      metav1.NamespaceSystem,
			caPEM:          caPEM,
			caKeyPEM:       caKeyPEM,
			certPEM:        invalidPEMWithoutBlock,
			secretDataKey:  corev1.TLSCertKey,
			certCommonName: "test-cn",
			certSANs:       []string{"192.168.1.1"},
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			caBlock, _ := pem.Decode([]byte(tt.caPEM))
			caCert, _ := x509.ParseCertificate(caBlock.Bytes)

			caKeyblock, _ := pem.Decode([]byte(tt.caKeyPEM))
			caKey, _ := x509.ParsePKCS1PrivateKey(caKeyblock.Bytes)

			client := fake.NewSimpleClientset(
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cert-secret-name",
						Namespace: tt.namespace,
					},
					Type: corev1.SecretTypeTLS,
					Data: map[string][]byte{
						tt.secretDataKey: []byte(tt.certPEM),
					},
				},
			)

			gotUpdate, err := CreateOrUpdateServerCertAndKeyToSecret(
				client, caCert, caKey,
				tt.certCommonName, tt.certSANs,
				"cert-secret-name",
			)
			if tt.expectedError {
				if err == nil {
					t.Error("error expected but no error reported")
				}
				return
			}

			if err != nil {
				t.Errorf("error not expected but an error was reported: %v", err)
				return
			}
			if tt.expectedUpdate != gotUpdate {
				t.Errorf("expect %t, got %t", tt.expectedUpdate, gotUpdate)
				return
			}
		})
	}
}
func TestNewServerCertAndKey(t *testing.T) {
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
			}

			if err != nil {
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

func TestCreateOrUpdateCertToSecret(t *testing.T) {
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
			if tt.expectedError {
				if err == nil {
					t.Errorf("error expected on %s, but no error reported", tt.name)
				}
				return
			}

			if err != nil {
				t.Errorf("error not expected on %s, but an error was reported (%v)", tt.name, err)
				return
			}

			// test create/update certificate to secret
			err = CreateOrUpdateCertAndKeyToSecret(fake.NewSimpleClientset(), caCert, cert, key, tt.secretName)
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
