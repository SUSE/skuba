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
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/coreos/go-oidc"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
	"golang.org/x/oauth2"
)

const (
	// "out-of-browser" URL "urn:ietf:wg:oauth:2.0:oob"
	// which triggers dex to display the OAuth2 code in the browser
	redirectURL = "urn:ietf:wg:oauth:2.0:oob"
)

// request represents an OAuth2 auth request flow
type request struct {
	IssuerURL          string
	Username           string
	Password           string
	RootCAData         []byte
	InsecureSkipVerify bool
	Debug              bool

	clientID     string
	clientSecret string

	provider *oidc.Provider
	verifier *oidc.IDTokenVerifier
	scopes   []string
}

// response is the final auth response
type response struct {
	IDToken      string
	AccessToken  string
	TokenType    string
	Expiry       time.Time
	RefreshToken string
	Scopes       []string
}

func oauth2Config(authReq request) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     authReq.clientID,
		ClientSecret: authReq.clientSecret,
		Endpoint:     authReq.provider.Endpoint(),
		Scopes:       authReq.scopes,
		RedirectURL:  redirectURL,
	}
}

type debugTransport struct {
	t  http.RoundTripper
	ar request
}

func (d debugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	pws := []string{d.ar.clientSecret, d.ar.Password}
	stripPasswords := func(str string) string {
		res := str
		for _, s := range pws {
			res = strings.Replace(res, s, "<REDACTED>", -1)
		}

		return res
	}

	reqDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		return nil, err
	}
	log.Print(stripPasswords(string(reqDump)))

	resp, err := d.t.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	respDump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		resp.Body.Close()
		return nil, err
	}
	log.Print(stripPasswords(string(respDump)))
	return resp, nil
}

func defaultTransport() *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

// httpClientForRootCAs returns a HTTP client which bypass SSL/TLS verify
func httpClientForSkipTLS() (*http.Client, error) {
	transport := defaultTransport()
	transport.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}

	return &http.Client{
		Transport: transport,
	}, nil
}

// httpClientForRootCAs returns a HTTP client which trusts provided root CAs
func httpClientForRootCAs(rootCA []byte) (*http.Client, error) {
	tlsConfig := tls.Config{RootCAs: x509.NewCertPool()}
	if !tlsConfig.RootCAs.AppendCertsFromPEM(rootCA) {
		return nil, errors.New("no valid certificates found in root CA file")
	}

	transport := defaultTransport()
	transport.TLSClientConfig = &tlsConfig

	return &http.Client{
		Transport: transport,
	}, nil
}

// doAuth will perform an OIDC / OAuth2 handshake without requiring a web browser
func doAuth(authReq request) (*response, error) {
	var err error
	var client *http.Client

	if authReq.InsecureSkipVerify {
		client, err = httpClientForSkipTLS()
		if err != nil {
			return nil, err
		}
	} else if len(authReq.RootCAData) > 0 {
		client, err = httpClientForRootCAs(authReq.RootCAData)
		if err != nil {
			return nil, err
		}
	} else {
		client = http.DefaultClient
		client.Transport = defaultTransport()
	}

	if authReq.Debug {
		client.Transport = debugTransport{t: client.Transport, ar: authReq}
	}

	ctx := oidc.ClientContext(context.Background(), client)
	provider, err := oidc.NewProvider(ctx, authReq.IssuerURL)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to query provider %q", authReq.IssuerURL)
	}

	var s struct {
		ScopesSupported []string `json:"scopes_supported"`
	}
	if err := provider.Claims(&s); err != nil {
		return nil, errors.Wrapf(err, "failed to parse provider scopes_supported")
	}

	authReq.provider = provider
	authReq.verifier = provider.Verifier(&oidc.Config{
		ClientID: authReq.clientID,
	})
	authReq.scopes = append(s.ScopesSupported, fmt.Sprintf("audience:server:client_id:%s", authProviderID))

	// Setup complete, start the actual auth
	authCodeURL := oauth2Config(authReq).AuthCodeURL("", oauth2.AccessTypeOffline)
	resp, err := client.Get(authCodeURL)
	if err != nil {
		return nil, errors.Wrapf(err, "failed on get auth code url")
	}
	defer resp.Body.Close()

	z := html.NewTokenizer(resp.Body)

	var actionLink string

Loop:
	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done
			break Loop
		case tt == html.StartTagToken || tt == html.SelfClosingTagToken:
			t := z.Token()

			switch t.Data {
			case "form":
				for _, a := range t.Attr {
					if a.Key == "action" {
						actionLink = a.Val
						break
					}
				}
			}
		}
	}

	formValues := url.Values{}
	formValues.Add("login", authReq.Username)
	formValues.Add("password", authReq.Password)

	oldRedirectChecker := client.CheckRedirect
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if strings.HasPrefix(redirectURL, req.RequestURI) {
			return http.ErrUseLastResponse
		}
		return nil
	}

	loginResp, err := client.PostForm(authReq.IssuerURL+actionLink, formValues)
	if err != nil {
		return nil, errors.Wrapf(err, "failed on post login url")
	}
	defer loginResp.Body.Close()

	approvalLocation, err := loginResp.Location()
	if err != nil {
		return nil, errors.Wrapf(err, "invalid username or password")
	}

	resp, err = client.Get(approvalLocation.String())
	if err != nil {
		return nil, errors.Wrapf(err, "failed on get approval location")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "failed on read approval body")
	}

	r, err := regexp.Compile("(?:(?:.|\n)*)value=\"(.*?)\"(?:(?:.|\n)*)")
	if err != nil {
		return nil, errors.Wrapf(err, "failed on regexp")
	}
	match := r.FindStringSubmatch(string(body))
	// We expect two matches - the entire body, and then just the code group
	if len(match) != 2 {
		return nil, errors.New("failed to extract token from OOB response")
	}
	code := match[1]

	client.CheckRedirect = oldRedirectChecker

	token, err := oauth2Config(authReq).Exchange(ctx, code)
	if err != nil {
		return nil, errors.Wrapf(err, "failed on exchange token")
	}

	result := &response{
		IDToken:      token.Extra("id_token").(string),
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		Expiry:       token.Expiry,
		Scopes:       authReq.scopes,
	}

	return result, nil
}
