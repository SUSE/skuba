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
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
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

	singleConnectorMsg    = "required id=\"password\""
	multipleConnectorsMsg = "dex-btn-icon--"
)

// request represents an OAuth2 auth request flow
type request struct {
	IssuerURL          string
	Username           string
	Password           string
	RootCAData         []byte
	InsecureSkipVerify bool
	AuthConnector      string
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
	pws := []string{d.ar.Password}
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

// httpClientForSkipTLS returns a HTTP client which bypass SSL/TLS verify
func httpClientForSkipTLS() *http.Client {
	transport := defaultTransport()
	transport.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}

	return &http.Client{
		Transport: transport,
	}
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
		client = httpClientForSkipTLS()
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

	respDump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return nil, errors.Wrapf(err, "http dump response")
	}

	loginURL := authReq.IssuerURL
	if strings.Contains(string(respDump), singleConnectorMsg) {
		// Handle single connector case
		c := processSingleConnector(resp.Body)
		loginURL = loginURL + c.url
	} else if strings.Contains(string(respDump), multipleConnectorsMsg) {
		// Handle multiple connectors case
		connectors := processMultipleConnectors(resp.Body)

		// Bump out interactive mode to let user choose auth connector
		if authReq.AuthConnector == "" {
			printConnectors(connectors)
			fmt.Print("\nEnter authentication connector ID: ")

			reader := bufio.NewReader(os.Stdin)
			authConnector, err := reader.ReadString('\n')
			if err != nil {
				return nil, errors.Wrapf(err, "read user input")
			}
			authReq.AuthConnector = strings.TrimSpace(authConnector)
		}

		match := false
		for _, c := range connectors {
			if c.id == authReq.AuthConnector {
				match = true
				loginURL = loginURL + c.url
				break
			}
		}
		if !match {
			fmt.Println("\nNo matched authentication connector ID")
			fmt.Printf("Your input is: %s\n", authReq.AuthConnector)
			printConnectors(connectors)
			return nil, errors.New("invalid input auth connector ID")
		}
	} else {
		return nil, errors.New("unknown connector resonse")
	}

	// Do authentication
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

	loginResp, err := client.PostForm(loginURL, formValues)
	if err != nil {
		return nil, errors.Wrapf(err, "failed on post login url")
	}
	defer loginResp.Body.Close()

	approvalLocation, err := loginResp.Location()
	if err != nil {
		return nil, errors.New("invalid username or password")
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

type connector struct {
	id  string
	url string
}

// processSingleConnector handles single connector case
// it returns single connector information
func processSingleConnector(body io.Reader) connector {
	c := connector{}
	z := html.NewTokenizer(body)

	for {
		tt := z.Next()
		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done
			return c
		case tt == html.StartTagToken:
			t := z.Token()
			switch t.Data {
			case "form":
				for _, attr := range t.Attr {
					if attr.Key == "action" {
						u, err := url.Parse(attr.Val)
						path := u.EscapedPath()
						if err == nil {
							c.id = path[strings.LastIndex(path, "/")+1:]
							c.url = attr.Val
						}
					}
				}
			}
		}
	}
}

// processMultipleConnectors handles multiple connectors case
// it returns all connectors information
func processMultipleConnectors(body io.Reader) []connector {
	c := make([]connector, 0)
	z := html.NewTokenizer(body)

	for {
		tt := z.Next()
		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done
			return c
		case tt == html.StartTagToken:
			t := z.Token()
			switch t.Data {
			case "a":
				// find <a href>
				for _, attr := range t.Attr {
					if attr.Key == "href" {
						u, err := url.Parse(attr.Val)
						path := u.EscapedPath()
						if err == nil {
							c = append(c, connector{
								id:  path[strings.LastIndex(path, "/")+1:],
								url: attr.Val,
							})
						}
					}
				}
			}
		}
	}
}

func printConnectors(connectors []connector) {
	fmt.Println("Available authentication connector IDs are:")
	for _, c := range connectors {
		fmt.Printf("  %s\n", c.id)
	}
}
