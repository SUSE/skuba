/*
 * Copyright (c) 2020 SUSE LLC.
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

package oidc

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/kubernetes/cmd/kubeadm/app/constants"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/pkg/skuba"
)

var (
	clusterCACert = []byte(`-----BEGIN CERTIFICATE-----
MIICyDCCAbCgAwIBAgIBADANBgkqhkiG9w0BAQsFADAVMRMwEQYDVQQDEwprdWJl
cm5ldGVzMB4XDTE5MDcxMDAzMDkyMFoXDTI5MDcwNzAzMDkyMFowFTETMBEGA1UE
AxMKa3ViZXJuZXRlczCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBALtk
cvsxbH5Cy2Wg+qZiWlXVmMsX9E+j1swFJmXtsvvB6eEr2Z9Xp+qYseQl/LPvJLNB
EV2/0ZK/MnhVuONZB+0zWmige4Yn0G3cqDDVmeCGEoTl6tyiT5dlC/dy+aAHVgj8
CZj+k+v3TJ29D8BquQh8LTvGYmTScqgDuJJfeVUQgkUco6DlojPLbNKqIvfjiqyQ
UCrVX/XdqyChbdepS+kafq/Ox6RlDBaoxQZtFBktoNNm4n2WREI9D7IBXhf03fvy
BN2awqeeLOmy43LEX6rEzrUVvzullh0NvK6dVfFc6hD4VxQOJMbMSSG1plO41ZWx
J6mWW2kDh0eIwpijcZECAwEAAaMjMCEwDgYDVR0PAQH/BAQDAgKkMA8GA1UdEwEB
/wQFMAMBAf8wDQYJKoZIhvcNAQELBQADggEBAEaOvTRa+0fAJpRzSPU8UfttLmTI
RemCU8lWjm23uOUaWCsjjHdB5AcW7v19FYYMHKiI2PzJIjDml3PTbSVG9PqzVSta
8YzMLf/dO3qR62JwFGyjtr14NV2tox4odS4ozCIDXKb3Uuk6QtgfLdvSqe5En0qw
QXvEbN7cwnaYVW8XHJ/AMCOCHHEKTkNIHPed17gDdMxtkfKB3cPKxBWEZglN0ML/
tNe10q61TNG2Zr66LHaH+1rXHyM65UQ3ZhPH4woxDtLl2xmVFN6X9kMnxrMe4NBO
xqBcRBbKj2oX77qoEGYn2RnnuD+aFEKKVwj9yKwzwBClDXJ/asUFeEd3JQA=
-----END CERTIFICATE-----
`)

	clusterCAKey = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAu2Ry+zFsfkLLZaD6pmJaVdWYyxf0T6PWzAUmZe2y+8Hp4SvZ
n1en6pix5CX8s+8ks0ERXb/Rkr8yeFW441kH7TNaaKB7hifQbdyoMNWZ4IYShOXq
3KJPl2UL93L5oAdWCPwJmP6T6/dMnb0PwGq5CHwtO8ZiZNJyqAO4kl95VRCCRRyj
oOWiM8ts0qoi9+OKrJBQKtVf9d2rIKFt16lL6Rp+r87HpGUMFqjFBm0UGS2g02bi
fZZEQj0PsgFeF/Td+/IE3ZrCp54s6bLjcsRfqsTOtRW/O6WWHQ28rp1V8VzqEPhX
FA4kxsxJIbWmU7jVlbEnqZZbaQOHR4jCmKNxkQIDAQABAoIBAASLIfKPNzc0fWQA
bDNejRwpqqO59/V6Xnnh4hV0lvNdt0L7YjJn2K1qeF20whTGxpgGCtrADN/G/P1H
7sysWvHYXefNhFEzY+1V/Un509pKxeYtAO3Kja15CLo+Mmk75E0hd0YbXhDJlTT5
4YjjzYq0IBCj9bzGEXubY31BDVsPmcB2SkNv+3gSmsxJB0H+N8zSniE+Nzm2I4UT
9ZF7ZWcsqMyAgmb6SXKL+ZKq1hzremNB2//ze4YxovvA1j7m/LHMhfzphk/KNiln
/JviBx0h/AdMGNq6phzWCfkElRncmp9VJXGmMyOJZwhtocjBMUArlZheIQsTKuLT
2Jfro0UCgYEAyXIr6K/Zlx4ZkIoGhNsEGS8prmuv2Rm+qxDOMwdHe7oc/cVkk7pF
0gdz6GXEFNL7LYb3pgrN2Iyy8UDU5kqVjI8+Q15Szkm+hlwroTkPZO3En0TTa6nA
SvuFvcfrLJZgfx2PYDD2pbk4DGRRbImm3RPlCe/FzEqI67ZiPLGd4uMCgYEA7iP4
gfQRiTkgZrFjFkutscPmE1mX1nVv1siJ0lhc1l/0+Bv8gQGxmulwc83djbuMnwBa
FvDaOiPfUPeQ/9GvStqK3m0niS3wvF0sEcDGDrCUYNpopUeLQoOfEJ0kNOIz7AIb
qzMO/eMe2/D/+4odP1YArMwQc9voVAk4BBZZn/sCgYAf1DVR368erHfXmadsCHr+
S7BnScaQI/w3ZUFfWLAXwZFhG3ZHzS/I/rp2ptWzgJ6FVAT/LEfYBmsjqV26QWCd
w+hPlrC4vXgoHJocMDrZdKQTkjlhkoG6l0FLejITanM2woHr7fCawMg1VQh0gM9n
sQQIbOUw4Tn/HGMrOdk7hwKBgQCdWWkvPLyFhhiRoi3Noo2PQth3+p/oFUqjiXf/
Y2FcSKUNdzh9aUgYCpzB98mnh7/fo5TjSZt4BRHeZuYJElyXwhU16LvR2WgSniGr
TUvQkv5HjKjOZJpwhZWJnbs5sikKjU4I7cC/It3WB8SsSNMQcVwa0O8iDrDRLhI0
KSxpFwKBgGN4FgXwBD3x7WUcckhzd5fu06wqiZCH0T0wMDZDVlh5b4a8d8zbK4n2
VdeDsT7jcvwDlI0PDCcqYwfCtp91OJlgeAPFEepGcdWzp8C/5aVqd7LqX4KQV2Vy
IkKGUpeNe2KbjCBkQX2prFLVmaFLGS7VtyDSvSKcGLSpWJNkvpBg
-----END RSA PRIVATE KEY-----
`)

	customOIDCCACert = []byte(`-----BEGIN CERTIFICATE-----
MIIEiTCCAvGgAwIBAgIQNkWxh+312VtqIZ3+rK4nfDANBgkqhkiG9w0BAQsFADBd
MR4wHAYDVQQKExVta2NlcnQgZGV2ZWxvcG1lbnQgQ0ExGTAXBgNVBAsMEGplbnRp
bmdAb3BlbnN1c2UxIDAeBgNVBAMMF21rY2VydCBqZW50aW5nQG9wZW5zdXNlMB4X
DTIwMDcxMzA2NDU0MloXDTMwMDcxMzA2NDU0MlowXTEeMBwGA1UEChMVbWtjZXJ0
IGRldmVsb3BtZW50IENBMRkwFwYDVQQLDBBqZW50aW5nQG9wZW5zdXNlMSAwHgYD
VQQDDBdta2NlcnQgamVudGluZ0BvcGVuc3VzZTCCAaIwDQYJKoZIhvcNAQEBBQAD
ggGPADCCAYoCggGBAOabSPrAcHL9mwB1VfawcgT3AEq3fcI36z66GNXhVUOSV+K7
GwK52zhL0xG2rd0AGhlUWc4W0l+bcSh06ZS5Omx7UfQDR8vL+9BMZ6cU67SWRShl
0eTeBxWlAgYgAFxgiM4xaGkhUabJtFG7LhKNyJb3bojWNSeJpZHFLVaBFVMQSP7z
fY2YIrMBRKlRq2X/Zg0Sg0KptbnQZhJ/zErnwIRoHvmH7FHpuECSRw6j47zWYNTB
6BjxMq/26t6eGTrz1SDO/+OfKELjfjQz8G/G711jz2mjEWamPO/4Z/Xv/BW0FL/M
4GfPwmwZrsYoSENGMH6qBpOwOy0UkBPczVYtakIQ1YrB3DDXVQj6fb42I0cGYPUC
6Qh1qH3KCxxVVrUo5/Cijzi8T0CNNTkrSR0yhOaosHX7z1qYyki5D6PhLPwqJMfH
BX3OsM5ohWuAEg+DcAXSfXr6CKkjmeOUq03SarV1mamuH+vZS674Fe+8SDzSE3D9
s4JrjsNZYpE13nOnNwIDAQABo0UwQzAOBgNVHQ8BAf8EBAMCAgQwEgYDVR0TAQH/
BAgwBgEB/wIBADAdBgNVHQ4EFgQUdLMNqbgu0gXysahLkFNOIOJ1YOMwDQYJKoZI
hvcNAQELBQADggGBAL14qNckB/zUrhybTJWqBk3ewR+6LKDQ3u/HGiCyn/w49iYO
Y4XW/Md+mxdNuTb+hYxHEDuYCxtQjI3GNuUXOXTNIY1A/DwulwVzcSWxRWUYb4s7
RVkMJBYFCafKlIUnX2mOdT7p+YRFQTl/4JHs827/GZ89PjthtvgGTVX1NuCW0Zuo
o9QPn8hICWyPWf1nd2+1ELQJ+8JGkGHYb43bUStD5zVvBzMTuOHCW4mCAn/uteGi
TB1bTAySf+4bc/Q9xKYU55ndfYv1xujh7KhJQAoRByZUkL2w5zK7bqbH//Dfcsbo
o/rDf6dx1PX17M97YCfyyaKEQkxwA44qJPD4y7fEW7Qk16zAVJEfz09En1r9Y8oK
xm1E+5JtJ/8pWBcdcj3jTogfJb9FRGXVZtpERoiD6ORtDOsugyTnMIkGUfrXxoif
yLqrkxzX+phhM/OFdjNKgIp5+U1s/YGyXsc/tNjnlyUzIhTNvsxLXeNNFZHa8gno
iVqPhDCmj01HGPK4rw==
-----END CERTIFICATE-----
`)

	customOIDCCAKey = []byte(`-----BEGIN PRIVATE KEY-----
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
ov5LcUMdfVabOYbjZTtxht9UzUI=
-----END PRIVATE KEY-----
`)

	oidcDexCert = []byte(`-----BEGIN CERTIFICATE-----
MIIECDCCAnCgAwIBAgIRALvXj2+uo5BbnFbcVGoKRg4wDQYJKoZIhvcNAQELBQAw
XTEeMBwGA1UEChMVbWtjZXJ0IGRldmVsb3BtZW50IENBMRkwFwYDVQQLDBBqZW50
aW5nQG9wZW5zdXNlMSAwHgYDVQQDDBdta2NlcnQgamVudGluZ0BvcGVuc3VzZTAe
Fw0yMDA3MTMwNjUwMDZaFw0yMTA3MTMwNjUwMDZaMCwxFzAVBgNVBAoTDnN5c3Rl
bTptYXN0ZXJzMREwDwYDVQQDEwhvaWRjLWRleDCCASIwDQYJKoZIhvcNAQEBBQAD
ggEPADCCAQoCggEBALYZc3k4pRBjtRuXE3LcAqCJ1JSi0eOnJlzLZJR6lezYhf7l
SxirVhqAl02L5uKRnspyuOWEbR8q42UnW3sXRh6CkpbFVN88gUq56E10H30G6dty
v+re1HRT3WYc8bKLxMghLEyHG6kfU9zKNtDATyY62RaSO7dLpapoV49c65xlaNix
rwdifkdplNZ0CpZE0Cset5HLQm16TVxQjmB4sDmtnxdxEGpoXFixAcb85LD+Aruh
a2NkBobtjNiFSk2LSM0KGD9c4BU0/3GHn5u8+PzTKUevcukLCTfIWEGoA7k3qdgq
8UXdnQetM/OYdAhLYyV5Pv6cTKj24RC5F233xmECAwEAAaN0MHIwDgYDVR0PAQH/
BAQDAgWgMBMGA1UdJQQMMAoGCCsGAQUFBwMBMAwGA1UdEwEB/wQCMAAwHwYDVR0j
BBgwFoAUdLMNqbgu0gXysahLkFNOIOJ1YOMwHAYDVR0RBBUwE4ILMTAuODQuNzIu
MTGHBApUSAswDQYJKoZIhvcNAQELBQADggGBACtAyLM625R/mrRZkq/yfYlN+LHd
FzNLnLbTnwHiEm74xziGIP/O6gA+uMWuSt7wI6jRkrHRj5Zq2rgdhO/RYsbfiJxX
L63e34f9UdzeWXCanaTYgECjGe5ILpxtUnvHJqwzi6rHGLrh/EsvRJpIvvwFTBCT
EY1tCp6T1ir9OQyQWHe0QqnXuK4oidxF4DsZxOQ4jKOnC+HjsiYhcwQ53GKMrren
Ls7EEIFvt2cK6qlbDus7amORXzDoX6IiKUQHqEB7Rse8osT3C8rbZtXACsk5lqgF
o+XnGnxXwEUbry0WRWAmGFgBH3Zc9D0CktYNtEqACryulaDfQk25S4MZ3vUh7Nnw
UGQoIOo2qHhp32W58+8d5qcGzxBEjvWA2HJ4AYEDRGzzTPiBviyUQ9Gy0J0j9plO
SZ1jNfggpulwhafPscvcebuHfI76IMtCITRwMzKDyCd5BHWxZYmdvX6hgei8fwJB
sJzGAD6jXz4h41szjIYjhTAanjYn+ZYU4KjTrQ==
-----END CERTIFICATE-----
`)

	oidcDexkey = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAthlzeTilEGO1G5cTctwCoInUlKLR46cmXMtklHqV7NiF/uVL
GKtWGoCXTYvm4pGeynK45YRtHyrjZSdbexdGHoKSlsVU3zyBSrnoTXQffQbp23K/
6t7UdFPdZhzxsovEyCEsTIcbqR9T3Mo20MBPJjrZFpI7t0ulqmhXj1zrnGVo2LGv
B2J+R2mU1nQKlkTQKx63kctCbXpNXFCOYHiwOa2fF3EQamhcWLEBxvzksP4Cu6Fr
Y2QGhu2M2IVKTYtIzQoYP1zgFTT/cYefm7z4/NMpR69y6QsJN8hYQagDuTep2Crx
Rd2dB60z85h0CEtjJXk+/pxMqPbhELkXbffGYQIDAQABAoIBAA5P+8p4UEj0fUSY
4Ddz4WT60FGKZpLNDW/XsKUW6Xe/IPEC1p8uwEq+9qVqrI/8QA3LbIrlmKoNdef6
au9GygUV4C2nft844zSbXg3QZbUu+Ox9nWX5c5tdCBbBiaGt6J6ONOwi5mKpiq7c
2egYZWAs2ekzPyxN7sxw/QjQldgp4kor9bF9WZ9R9tLNR/Z/yliFIeYSj8g2grPt
mXMpZs3CVUUEJp69/iHrJ6SCkCQMKSpWYHyCA3FOY14S19j+gdnNCxZypYRcrAye
55q0M6zI6t3O/80VsygES1+Tq9zgNrMHqzBYwWmONn0NptD1xV8fzuusnoseVE4U
CVb4VUECgYEA5wGsW/mrNs00PXPuLaOvFL3I+RZK7DyqyGRC/QBxMh9ZEUXCR148
zlHxPIp2FhNruWvP9cYzlL6UUVYk9+8cYsXFaO/C3bYFwT35sOyJUJ3O2Tvv9yBY
n4HTV7P69U7CL7MK5boPnfUsYeKT1WlsUx46XymXORGDpjBqcAHS5fkCgYEAyc0r
AMd9LOed2fS7gAVu4Ao2EJ/1IwMF8bBSe1ISOrPxgJTt+JCv6OF148SxHk1xBvBR
vdUgJUei5xjngdbvk8LWgelIzX3W6EHSAkkHtHLd9lpHeskrROXYNP4LHWXVRLTW
Lm8hjXHYPTRY3bVBpiyo+stzZc6Snmc5Kchm3akCgYEAyiuAuQ4Mde2pZo7rSC4U
sEZYeQa1k4KUxMRajCmy53bf8Gno2aTz+m1kfuN+7VsZ0DE205Ye1nLkQzrtJ7+w
TBFh77DGDlubNcATom+gzVkPCreWD+XTKeXpHLx7Se0frbc4Nk1cFZXYveIaF5Ao
KaYu19ICcwONAAknXdd6x6ECgYA7y4JRgcrScnLwcTbZsUJwOjZY6Ly/OhcZzVAz
YFcsc8M8gWSeAWlOTPgcnFyLCRFTqAPghvU2dqqLZXK7o09r7hCXf+NlmEMEoPQ0
XyVcT6j7ZTbG6DLdAGn3EcuDU3hFGnxYV++ONMyJHiiy0RF5xsPvRDeWVAZXz2g9
vDbWuQKBgCKJCJMTyZY/60JUKFpoqp5gxAO0GnTwI/oNzqIypVWROjSWdCrn/vHD
pIFOK2jbc3YXMCHSx75kwv9YbD8GspaZ8Hglu1OjhpK+TSX80ilBwg3DJt16j8RU
UHbQgtuV9b4Jqz2s5cfXy45aRKXroTVcdJyboVzH25/MBSaspGTG
-----END RSA PRIVATE KEY-----
`)

	oidcGangwayCert = []byte(`-----BEGIN CERTIFICATE-----
MIIECzCCAnOgAwIBAgIQTD0RIffuVKlGQNrnFa6LsjANBgkqhkiG9w0BAQsFADBd
MR4wHAYDVQQKExVta2NlcnQgZGV2ZWxvcG1lbnQgQ0ExGTAXBgNVBAsMEGplbnRp
bmdAb3BlbnN1c2UxIDAeBgNVBAMMF21rY2VydCBqZW50aW5nQG9wZW5zdXNlMB4X
DTIwMDcxMzA2NTAzNVoXDTIxMDcxMzA2NTAzNVowMDEXMBUGA1UEChMOc3lzdGVt
Om1hc3RlcnMxFTATBgNVBAMTDG9pZGMtZ2FuZ3dheTCCASIwDQYJKoZIhvcNAQEB
BQADggEPADCCAQoCggEBAJUapRtcvxpetC44ThJulucwprmQ7Y6p2vl01nTE+9bf
5tbaNwe+gCS7+J//s+MhUSrfOneiaVyh0C8Gcmugd4wl8vYIevaCDMdRL4YUimjT
Bpjz4BAEgREoPQBaWqu5EAYw/r5o7G0C0uzXdljHZhduDILOBoZDpwu6mDHCWNHS
nc6vXuRAXuXros2of85RjQNellm07p7qwWuH9i1hcYAG5RGY+jYZvf9MRk8CHeFq
YKNQkVhOrp0/D5AYFA90P6uPWOJaLuX/6I/4F/0DqS1RQlMNCXZeDZ4yMfBRiBhn
akIhMtXCYwm5hcx8G8uKPc6wBzPr+TWzYB9BwzIZEYcCAwEAAaN0MHIwDgYDVR0P
AQH/BAQDAgWgMBMGA1UdJQQMMAoGCCsGAQUFBwMBMAwGA1UdEwEB/wQCMAAwHwYD
VR0jBBgwFoAUdLMNqbgu0gXysahLkFNOIOJ1YOMwHAYDVR0RBBUwE4ILMTAuODQu
NzIuMTGHBApUSAswDQYJKoZIhvcNAQELBQADggGBAKWUlK5PX3GU6U214p3LEmLL
UVmO7Vj2rod28/KR7hoPKuLNjqYsNwVhGnU2QpzqjqNndNH+/lNA0OOwxrCas3uI
bYPJ51axlwxQ+/9R4qDwEYc2tK/FtJTOLplCAuOsM2McyHhLoJUEq/lh3U4Dn5Y5
hxZfayyCIErev5ddve/krTPGmeJuzlrbQYGhyZWWzaaHcttdJe9gpuTn2VdnxXsH
Bg7YSHj5B/zNPcBjZKOb6AIRH/kgvL8B7yuqndrj3WHeUBGBcN0vHIJoCpS6PR2L
fpNKplMEifF130QZzp/bo/ggjQ1n2NfBXo7IOLrW0rMh0TCo0nw30Tz0yZMX2Y82
dkHdXnlYOyzWZ//RXG3NLuGcyM4X+ztKkwn53SyoJWZs3xKYhRRmOabjMpvkDQAF
cJTvh3e2ZXUt6R1B0FPiK55tk/eiKupqbwBol4gUZ3iPwy1gy2RVZJVq1L8n5cDX
PMwrzUB8riDPt6khiAbFW2JYl7u3XvNastRY72Ledg==
-----END CERTIFICATE-----
`)

	oidcGangwayKey = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEogIBAAKCAQEAlRqlG1y/Gl60LjhOEm6W5zCmuZDtjqna+XTWdMT71t/m1to3
B76AJLv4n/+z4yFRKt86d6JpXKHQLwZya6B3jCXy9gh69oIMx1EvhhSKaNMGmPPg
EASBESg9AFpaq7kQBjD+vmjsbQLS7Nd2WMdmF24Mgs4GhkOnC7qYMcJY0dKdzq9e
5EBe5euizah/zlGNA16WWbTunurBa4f2LWFxgAblEZj6Nhm9/0xGTwId4Wpgo1CR
WE6unT8PkBgUD3Q/q49Y4lou5f/oj/gX/QOpLVFCUw0Jdl4NnjIx8FGIGGdqQiEy
1cJjCbmFzHwby4o9zrAHM+v5NbNgH0HDMhkRhwIDAQABAoIBACY3DwwlSEGc9jdG
bRQiTgDxIjo7X/eJpzr6bCC/gACBoGt+wE320GcZ2k4/vj4/gssnLZgOA00fl/kF
gPv+w8Ui0NOoNsPJbzcCotiNmS/mrjEziEDytpXMJ15YyhZrNyUvF4t73uTNrXvQ
eV+ls4/bv88iuN9epYjHCUzvj84u9eVzZPy3hWpdTcazwH442H1odNg2/4IOCzkY
6FVmb1JYjT/yaHjonrFsH2QUUld3lpqxTdcR0NszxRTCySGtSDmF7PhozfWGBNSv
LlfekKsF9cDsg79DaChmwFce4nPQl/SIPh+fl1SbmNCDQ515mSwObA5BQzsqTOll
++kePAkCgYEAxaQ/SIjDshTImWD/PaN1U4MJwpazFpT0MkVXeEUq3atp7GeU5xJe
A3gwYHrJ123e0P4S5Rtvb8Kh5SGe4b4Hyde8F1N9IqmiIij8hYm/8jEAWE0yk6qk
DlMtqoayZ2UZe0YQssXMFDtbzExnyAbiCwggBZAhsYNwul6BVafG1nsCgYEAwSFw
+tiTbXZthDwgWORSEi0WOrQHIv4RHqOqQmb5fxs5zjX6QNeLBmlBSU1+0YrRpw0T
9z9UwNucO1NU4Ruh7TJM/gFyhiwuT0jZ9zTeuI/KuN989nZQjgqcg6pQwC9VX+L9
g/Nnuw8A5s9BFmmit48Tr3N06KnPD1iNQn+qaWUCgYAYxWtBFhMhAMXbo3KaMSCF
ZQkWIHk1vVmV62b5JgInYlKWVK0vAPhTiv7VOM6Pd6/TleScXoHrCgPsifg15vFm
9OkYK1ilvYkaqvRrcEZkfovChXpvU5XYTciNdPBrURqOfsuc/HmFl6L7yh+/zE0M
gOoyiEwQyZ6ZXTrsl2iufQKBgDK1zSyQYWWEiw0FnJi6mrIbFJMlYhpWC7i30KTO
1QQC6hKzKZqM/fwY9wOATaRHhvUOAggRoPdisosBPnA9CS923bB0QNXqE97Nii3W
vARJ/Ti9tdohBtXFA4Ou3LUZuJkMyPQ0nTAIqHvyP2zbH9aCwvB2qGPO8oddAPpM
+znhAoGAc8GVicNBiDKJji09YhN29PVj26brHTTdfcOA3lCzBfkko5ucwZ9osJFU
+fzmDo3PNfDNlqwYl0oYeWGMrudcVfsb9Dm3nXjrj4cXLAhTvOVaOdGSpFQzYGQJ
MMIAV64pSSMHH6L7H6X0hNxlgG/0J9n7p0r1ES01Nh5jvKIwYCw=
-----END RSA PRIVATE KEY-----
`)
)

func setupCertAndKey(t *testing.T, cert, key *[]byte, certFileName, keyFileName string) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Errorf("unable to get current directory: %v", err)
		return
	}

	// write temporary pki folder
	if err := os.Mkdir(filepath.Join(pwd, skuba.PkiDir()), 0700); err != nil {
		if !os.IsExist(err) {
			t.Errorf("unable to create directory %s: %v", skuba.PkiDir(), err)
			return
		}
	}

	if cert != nil {
		if err := ioutil.WriteFile(filepath.Join(filepath.Join(pwd, skuba.PkiDir()), certFileName), *cert, 0644); err != nil {
			t.Errorf("unable to write CA cert: %v", err)
			return
		}
	}

	if key != nil {
		if err := ioutil.WriteFile(filepath.Join(filepath.Join(pwd, skuba.PkiDir()), keyFileName), *key, 0600); err != nil {
			t.Errorf("unable to write CA key: %v", err)
			return
		}
	}
}

func teardownCertAndKey(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Errorf("unable to get current directory: %v", err)
		return
	}

	// removes rendered pki folder
	dir := filepath.Join(pwd, skuba.PkiDir())
	if f, err := os.Stat(dir); !os.IsNotExist(err) && f.IsDir() {
		if err := os.RemoveAll(dir); err != nil {
			t.Errorf("unable to remove rendered addon folder: %v", err)
			return
		}
	}
}

func checkSecretCertAndKey(t *testing.T, client clientset.Interface, secretName string, caCert []byte, cert, key *[]byte) bool {
	// check the Secret resource is the same as local server cert file
	secret, err := client.CoreV1().Secrets(metav1.NamespaceSystem).Get(context.TODO(), secretName, metav1.GetOptions{})
	exist, _ := kubernetes.DoesResourceExistWithError(err)
	if !exist {
		t.Errorf("expected secret exited but not found: %v", err)
		return false
	}

	// check CA
	if !reflect.DeepEqual(secret.Data["ca.crt"], caCert) {
		t.Error("the ca cert mismatch")
		return false
	}
	roots := x509.NewCertPool()
	if ok := roots.AppendCertsFromPEM(secret.Data["ca.crt"]); !ok {
		t.Error("failed to parse root certificate")
		return false
	}

	// check cert
	if len(secret.Data[corev1.TLSCertKey]) == 0 {
		t.Error("the tls cert should exist")
		return false
	}
	if cert != nil && !reflect.DeepEqual(secret.Data[corev1.TLSCertKey], *cert) {
		t.Error("the tls cert mismatch")
		return false
	}
	block, _ := pem.Decode(secret.Data[corev1.TLSCertKey])
	if block == nil {
		t.Error("failed to parse certificate PEM")
		return false
	}
	parseCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Errorf("failed to parse certificate: %v", err)
		return false
	}
	if _, err := parseCert.Verify(x509.VerifyOptions{Roots: roots}); err != nil {
		t.Errorf("failed to verify certificate: %v", err)
		return false
	}

	// check key
	if len(secret.Data[corev1.TLSPrivateKeyKey]) == 0 {
		t.Error("the tls key should exist")
		return false
	}
	if cert != nil && !reflect.DeepEqual(secret.Data[corev1.TLSPrivateKeyKey], *key) {
		t.Error("the tls key mismatch")
		return false
	}
	return true
}

func TestIsCACertAndKeyExist(t *testing.T) {
	gotCert, gotKey := IsCACertAndKeyExist()
	if gotCert == true {
		t.Error("expected cert not exist but exited")
		return
	}
	if gotKey == true {
		t.Error("expected key not exist but exited")
		return
	}

	setupCertAndKey(t, &clusterCACert, &clusterCAKey, CACertFileName, caKeyFileName)
	defer teardownCertAndKey(t)

	gotCert, gotKey = IsCACertAndKeyExist()
	if gotCert == false {
		t.Error("expected cert exist but not exited")
		return
	}
	if gotKey == false {
		t.Error("expected key exist but not existed")
		return
	}
}

func TestTryToUseLocalServerCert(t *testing.T) {
	tests := []struct {
		name                string
		cert, key           []byte
		localServerFileName string
		certSecretName      string
	}{
		{
			name:                "oidc-dex",
			cert:                oidcDexCert,
			key:                 oidcDexkey,
			localServerFileName: DexServerCertAndKeyBaseFileName,
			certSecretName:      DexCertSecretName,
		},
		{
			name:                "oidc-gangway",
			cert:                oidcGangwayCert,
			key:                 oidcGangwayKey,
			localServerFileName: GangwayServerCertAndKeyBaseFileName,
			certSecretName:      GangwayCertSecretName,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			client := fake.NewSimpleClientset()

			if err := TryToUseLocalServerCert(client, tt.localServerFileName, tt.certSecretName); err == nil {
				t.Error("expected got err but not error reported")
				return
			}

			setupCertAndKey(t, &tt.cert, &tt.key, fmt.Sprintf("%s.crt", tt.localServerFileName), fmt.Sprintf("%s.key", tt.localServerFileName))
			if err := TryToUseLocalServerCert(client, tt.localServerFileName, tt.certSecretName); err == nil {
				t.Error("expected got err but not error reported")
				return
			}

			setupCertAndKey(t, &customOIDCCACert, &customOIDCCAKey, CACertFileName, caKeyFileName)
			if err := TryToUseLocalServerCert(client, tt.localServerFileName, tt.certSecretName); err != nil {
				t.Errorf("expected no error but an error reported: %v", err)
				return
			}
			teardownCertAndKey(t)

			// check the Secret resource is the same as local server cert file
			if !checkSecretCertAndKey(t, client, tt.certSecretName, customOIDCCACert, &tt.cert, &tt.key) {
				return
			}
		})
	}
}

func TestSignServerWithLocalCACertAndKey(t *testing.T) {
	tests := []struct {
		name             string
		certCN           string
		controlPlaneHost string
		certSecretName   string
	}{
		{
			name:             "oidc-dex",
			certCN:           DexCertCN,
			controlPlaneHost: "10.10.10.10",
			certSecretName:   DexCertSecretName,
		},
		{
			name:             "oidc-gangway",
			certCN:           GangwayCertCN,
			controlPlaneHost: "unit-test.suse.com",
			certSecretName:   GangwayCertSecretName,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			client := fake.NewSimpleClientset()

			if err := SignServerWithLocalCACertAndKey(client, tt.certCN, tt.controlPlaneHost, tt.certSecretName); err == nil {
				t.Error("expected got err but not error reported")
				return
			}

			setupCertAndKey(t, &clusterCACert, &clusterCAKey, constants.CACertName, constants.CAKeyName)
			if err := SignServerWithLocalCACertAndKey(client, tt.certCN, tt.controlPlaneHost, tt.certSecretName); err != nil {
				t.Errorf("expected no error but an error reported: %v", err)
				return
			}
			teardownCertAndKey(t)

			// check the server certificate is signed by cluster CA cert/key
			if !checkSecretCertAndKey(t, client, tt.certSecretName, clusterCACert, nil, nil) {
				return
			}

			setupCertAndKey(t, &customOIDCCACert, nil, CACertFileName, caKeyFileName)
			if err := SignServerWithLocalCACertAndKey(client, tt.certCN, tt.controlPlaneHost, tt.certSecretName); err == nil {
				t.Error("expected got err but not error reported")
				return
			}
			teardownCertAndKey(t)

			setupCertAndKey(t, &customOIDCCACert, &customOIDCCAKey, CACertFileName, caKeyFileName)
			if err := SignServerWithLocalCACertAndKey(client, tt.certCN, tt.controlPlaneHost, tt.certSecretName); err != nil {
				t.Errorf("expected no error but an error reported: %v", err)
				return
			}
			teardownCertAndKey(t)

			// check the server certificate is signed by cluster CA cert/key
			if !checkSecretCertAndKey(t, client, tt.certSecretName, customOIDCCACert, nil, nil) {
				return
			}
		})
	}
}
