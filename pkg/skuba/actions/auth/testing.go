/*
 * Copyright (c) 2019 SUSE LLC. All rights reserved.
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
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

const (
	mockDefaultUsername = "hello@suse.com"
	mockDefaultPassword = "bar"

	mockIDToken      = "eyJhbGciOiJSUzI1NiIsImtpZCI6ImFlNjg5YTI1OWJkYjRjMWZiZDZmZGFjMzg0OTk5YTJhNWNlNmRmOGEifQ.eyJpc3MiOiJodHRwczovLzEwLjg2LjAuMTE2OjMyMDAwIiwic3ViIjoiQ2loamJqMW9aV3hzYjNkdmNteGtMRzkxUFhWelpYSnpMR1JqUFdWNFlXMXdiR1VzWkdNOWIzSm5FZ1JzWkdGdyIsImF1ZCI6WyJvaWRjIiwib2lkYy1jbGkiXSwiZXhwIjoxNTY1MTY1NjE0LCJpYXQiOjE1NjUwNzkyMTQsImF6cCI6Im9pZGMtY2xpIiwiYXRfaGFzaCI6ImJ2cml3TEthMzZuTjExUG5Jc3RRSmciLCJlbWFpbCI6ImhlbGxvQHN1c2UuY29tIiwiZW1haWxfdmVyaWZpZWQiOnRydWUsImdyb3VwcyI6WyJkZXZlbG9wZXJzIl0sIm5hbWUiOiJIZWxsbyBXb3JsZCJ9.AYN9cbk2hS6S8ZbQSZ4yoGksPJJ9qzbK8iXCoB6XXmhc5AUlwxnXQ-vzcp1u6h8AtY3iJX0s5ZwH3BthKEBlj6Aad6v5qp62Ws0Wb1-RY6TcCNQv4AdpBuFlJtJIxp7wI33bR0gpLOMsjYJRgKuLvQ1Dn7tipT62CPhqwA91lT613_yByLC8ek1Qy3RSwJIA_hkJT0H-yMHM2JC5WuB3P0MEURfl2QIXaWDjoV5RcL0dh_dkwy2v6zxgCPu0gFvL2BOrcHPjv6k6kphMnQ8uCbQaEfNxuMYr7zDRWBcNSpfjhbbYRAjNBHbpMorM3mT83GB76cxdUWCW2q69nM1B_w"
	mockRefreshToken = "ChludG1ncnh1aHQ1a3F0dG83enRvYmtlc2hiEhlsNWNvM3V6cmVjb2FxYW1maHZqa2F5azJh"
)

var (
	localhostCert = []byte(`-----BEGIN CERTIFICATE-----
MIICEzCCAXygAwIBAgIQMIMChMLGrR+QvmQvpwAU6zANBgkqhkiG9w0BAQsFADAS
MRAwDgYDVQQKEwdBY21lIENvMCAXDTcwMDEwMTAwMDAwMFoYDzIwODQwMTI5MTYw
MDAwWjASMRAwDgYDVQQKEwdBY21lIENvMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCB
iQKBgQDuLnQAI3mDgey3VBzWnB2L39JUU4txjeVE6myuDqkM/uGlfjb9SjY1bIw4
iA5sBBZzHi3z0h1YV8QPuxEbi4nW91IJm2gsvvZhIrCHS3l6afab4pZBl2+XsDul
rKBxKKtD1rGxlG4LjncdabFn9gvLZad2bSysqz/qTAUStTvqJQIDAQABo2gwZjAO
BgNVHQ8BAf8EBAMCAqQwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDwYDVR0TAQH/BAUw
AwEB/zAuBgNVHREEJzAlggtleGFtcGxlLmNvbYcEfwAAAYcQAAAAAAAAAAAAAAAA
AAAAATANBgkqhkiG9w0BAQsFAAOBgQCEcetwO59EWk7WiJsG4x8SY+UIAA+flUI9
tyC4lNhbcF2Idq9greZwbYCqTTTr2XiRNSMLCOjKyI7ukPoPjo16ocHj+P3vZGfs
h1fIw3cSS2OolhloGw/XM6RWPWtPAlGykKLciQrBru5NAPvCMsb/I1DAceTiotQM
fblo6RBxUQ==
-----END CERTIFICATE-----
`)
)

func newID() string {
	var encoding = base32.NewEncoding("abcdefghijklmnopqrstuvwxyz234567")

	buff := make([]byte, 16) // 128 bit random ID.
	if _, err := io.ReadFull(rand.Reader, buff); err != nil {
		panic(err)
	}
	// Avoid the identifier to begin with number and trim padding
	return string(buff[0]%26+'a') + strings.TrimRight(encoding.EncodeToString(buff[1:]), "=")
}

func openIDHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		url := fmt.Sprintf("%s://%s", defaultScheme, r.Host)

		_ = json.NewEncoder(w).Encode(&map[string]interface{}{
			"issuer":                                url,
			"authorization_endpoint":                fmt.Sprintf("%s/auth", url),
			"token_endpoint":                        fmt.Sprintf("%s/token", url),
			"jwks_uri":                              fmt.Sprintf("%s/keys", url),
			"response_types_supported":              []string{"code"},
			"subject_types_supported":               []string{"public"},
			"id_token_signing_alg_values_supported": []string{"RS256"},
			"scopes_supported":                      []string{"openid", "email", "groups", "profile", "offline_access"},
			"token_endpoint_auth_methods_supported": []string{"client_secret_basic"},
			"claims_supported":                      []string{"aud", "email", "email_verified", "exp", "iat", "iss", "locale", "name", "sub"},
		})
	}
}

func openIDHandlerInvalidScopes() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		url := fmt.Sprintf("%s://%s", defaultScheme, r.Host)

		_ = json.NewEncoder(w).Encode(&map[string]interface{}{
			"issuer":           url,
			"scopes_supported": []int{1, 2, 3},
		})
	}
}

func openIDHandlerNoScopes() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		url := fmt.Sprintf("%s://%s", defaultScheme, r.Host)

		_ = json.NewEncoder(w).Encode(&map[string]interface{}{
			"issuer": url,
		})
	}
}

func authHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		url := fmt.Sprintf("%s://%s", defaultScheme, r.Host)

		http.Redirect(w, r, fmt.Sprintf("%s/auth/local", url)+"?req="+newID(), http.StatusFound)
	}
}

func authLocalHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		url := fmt.Sprintf("%s://%s", defaultScheme, r.Host)
		authReqID := r.FormValue("req")

		switch r.Method {
		case http.MethodGet:
			htmlOutput := fmt.Sprintf(`
				<div class="theme-panel">
				<h2 class="theme-heading">Log in to Your Account</h2>
				<form method="post" action="/auth/local?req=%s">
					<div class="theme-form-row">
					<div class="theme-form-label">
						<label for="userid">Email Address</label>
					</div>
					<input tabindex="1" required id="login" name="login" type="text" class="theme-form-input" placeholder="email address"  autofocus />
					</div>
					<div class="theme-form-row">
					<div class="theme-form-label">
						<label for="password">Password</label>
						</div>
					<input tabindex="2" required id="password" name="password" type="password" class="theme-form-input" placeholder="password" />
					</div>
					<button tabindex="3" id="submit-login" type="submit" class="dex-btn theme-btn--primary">Login</button>
				</form>
				</div>
				`, authReqID)
			_, _ = w.Write([]byte(htmlOutput))
		case http.MethodPost:
			username := r.FormValue("login")
			password := r.FormValue("password")
			if username == mockDefaultUsername && password == mockDefaultPassword {
				http.Redirect(w, r, fmt.Sprintf("%s/approval", url)+"?req="+authReqID, http.StatusSeeOther)
			}
		}
	}
}

func tokenHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := struct {
			AccessToken  string `json:"access_token"`
			TokenType    string `json:"token_type"`
			ExpiresIn    int    `json:"expires_in"`
			RefreshToken string `json:"refresh_token,omitempty"`
			IDToken      string `json:"id_token"`
		}{
			newID(),
			"bearer",
			86399,
			mockRefreshToken,
			mockIDToken,
		}
		data, _ := json.Marshal(resp)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Length", strconv.Itoa(len(data)))
		_, _ = w.Write(data)
	}
}

func approvalHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		htmlOutput := fmt.Sprintf(`
			<div class="theme-panel">
			<h2 class="theme-heading">Login Successful</h2>
			<p>Please copy this code, switch to your application and paste it there:</p>
			<input type="text" class="theme-form-input" value="%s" />
			</div>
		`, newID())
		_, _ = w.Write([]byte(htmlOutput))
	}
}

func approvalInvalidBodyHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		htmlOutput := fmt.Sprintf(`
			<div class="theme-panel">
			<h2 class="theme-heading">Login Successful</h2>
			</div>
		`)
		_, _ = w.Write([]byte(htmlOutput))
	}
}
