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
	"reflect"
	"strings"
	"testing"
)

func Test_processSingleConnector(t *testing.T) {
	tests := []struct {
		name              string
		body              string
		expectedConnector connector
	}{
		{
			name: "single connector",
			body: `
			<!DOCTYPE html>
			<html>
				<head>
				<title>SUSE CaaS Platform</title>
				</head>
				<body class="theme-body">
				<div class="theme-navbar">
					<div class="theme-navbar__logo-wrap">
						<img class="theme-navbar__logo" src="https://10.86.4.17:32000/theme/logo.png">
					</div>
				</div>
				<div class="dex-container">
				<div class="theme-panel">
				<h2 class="theme-heading">Log in to Your Account</h2>
				<form method="post" action="/auth/ldap000?req=yfzclhtr4twoqqvixmdt75goe">
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
				</body>
			</html>
			`,
			expectedConnector: connector{
				id:  "ldap000",
				url: "/auth/ldap000?req=yfzclhtr4twoqqvixmdt75goe",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			gotConnectors := processSingleConnector(strings.NewReader(tt.body))

			if !reflect.DeepEqual(gotConnectors, tt.expectedConnector) {
				t.Errorf("got %v, want %v", gotConnectors, tt.expectedConnector)
				return
			}
		})
	}
}

func Test_processMultipleConnectors(t *testing.T) {
	tests := []struct {
		name               string
		body               string
		expectedConnectors []connector
	}{
		{
			name: "multiple connectors",
			body: `
				<!DOCTYPE html>
				<html>
					<head>
					<title>SUSE CaaS Platform</title>
					</head>
					<body class="theme-body">
					<div class="dex-container">
						<div class="theme-panel">
							<h2 class="theme-heading">Log in to SUSE CaaS Platform</h2>
							<div>
							<div class="theme-form-row">
								<a href="/auth/local?req=ft6cvb6b4om3y7cbe3ncfw6mf" target="_self">
								<button class="dex-btn theme-btn-provider">
									<span class="dex-btn-icon dex-btn-icon--local"></span>
									<span class="dex-btn-text">Log in with Email</span>
								</button>
								</a>
							</div>
							<div class="theme-form-row">
								<a href="/auth/ldap?req=ft6cvb6b4om3y7cbe3ncfw6mf" target="_self">
								<button class="dex-btn theme-btn-provider">
									<span class="dex-btn-icon dex-btn-icon--ldap"></span>
									<span class="dex-btn-text">Log in with openLDAP</span>
								</button>
								</a>
							</div>
						</div>
					</div>
					</body>
				</html>
				`,
			expectedConnectors: []connector{
				{
					id:  "local",
					url: "/auth/local?req=ft6cvb6b4om3y7cbe3ncfw6mf",
				},
				{
					id:  "ldap",
					url: "/auth/ldap?req=ft6cvb6b4om3y7cbe3ncfw6mf",
				},
			},
		},
		{
			name: "no connector found",
			body: `
				<!DOCTYPE html>
				<html>
					<head>
					<title>SUSE CaaS Platform</title>
					</head>
					<body class="theme-body">
					<div class="dex-container">
						<div class="theme-panel">
							<h2 class="theme-heading">Log in to SUSE CaaS Platform</h2>
							<div>
						</div>
					</div>
					</body>
				</html>
				`,
			expectedConnectors: []connector{},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			gotConnectors := processMultipleConnectors(strings.NewReader(tt.body))

			if !reflect.DeepEqual(gotConnectors, tt.expectedConnectors) {
				t.Errorf("got %v, want %v", gotConnectors, tt.expectedConnectors)
				return
			}
		})
	}
}

func Example_printConnectors() {
	c := []connector{
		{id: "local", url: "/auth/local?req=ft6cvb6b4om3y7cbe3ncfw6mf"},
		{id: "ldap", url: "/auth/ldap?req=ft6cvb6b4om3y7cbe3ncfw6mf"},
	}
	printConnectors(c)

	// Output:
	// Available authentication connector IDs are:
	//   local
	//   ldap
}
