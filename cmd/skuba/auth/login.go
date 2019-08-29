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
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
	"k8s.io/klog"

	"github.com/SUSE/skuba/pkg/skuba/actions/auth"
)

// NewLoginCmd creates a new `skuba login` cobra command
func NewLoginCmd() *cobra.Command {
	cfg := auth.LoginConfig{}

	cmd := cobra.Command{
		Use:   "login",
		Short: "Login to a cluster",
		Run: func(cmd *cobra.Command, args []string) {
			if cfg.Username == "" {
				reader := bufio.NewReader(os.Stdin)
				fmt.Print("Enter your username: ")
				username, err := reader.ReadString('\n')
				if err != nil {
					klog.Fatalf("error on read username: %v", err)
				}

				username = strings.TrimSpace(username)
				if username == "" {
					fmt.Println("A username must be provided")
					os.Exit(1)
				}
				cfg.Username = username
			}

			if cfg.Password == "" {
				fmt.Print("Enter your password: ")
				bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
				if err != nil {
					klog.Fatalf("error on read password: %v", err)
				}
				fmt.Println("")

				password := strings.TrimSpace(string(bytePassword))
				if password == "" {
					fmt.Println("A password must be provided")
					os.Exit(1)
				}
				cfg.Password = password
			}

			if cfg.RootCAPath != "" {
				fi, err := os.Stat(cfg.RootCAPath)
				if os.IsNotExist(err) {
					fmt.Printf("Root certificate authority chain file \"%s\" not exist\n", cfg.RootCAPath)
					os.Exit(1)
				}
				if fi.IsDir() {
					fmt.Printf("Root certificate authority chain file \"%s\" is a folder\n", cfg.RootCAPath)
					os.Exit(1)
				}
			}

			kubeCfg, err := auth.Login(cfg)
			if err != nil {
				klog.Fatalf("error on login: %v", err)
			}

			if err := auth.SaveKubeconfig(cfg.KubeConfigPath, kubeCfg); err != nil {
				klog.Fatalf("error on save kubeconfig: %v", err)
			}

			fmt.Printf("You have been logged in successfully. kubeconfig path %s\n", cfg.KubeConfigPath)
		},
	}

	cmd.Flags().StringVarP(&cfg.DexServer, "server", "s", "", "OIDC dex server url https://<IP/FQDN>:<Port> (required)")
	cmd.Flags().StringVarP(&cfg.Username, "username", "u", "", "Username")
	cmd.Flags().StringVarP(&cfg.Password, "password", "p", "", "Password")
	cmd.Flags().StringVarP(&cfg.AuthConnector, "auth-connector", "a", "", "Authentication connector ID")
	cmd.Flags().StringVarP(&cfg.RootCAPath, "root-ca", "r", "", "Root certificate authority chain file")
	cmd.Flags().BoolVarP(&cfg.InsecureSkipVerify, "insecure", "k", false, "Insecure SSL connection")
	cmd.Flags().StringVarP(&cfg.ClusterName, "cluster-name", "n", "local", "Kubernetes cluster name")
	cmd.Flags().StringVarP(&cfg.KubeConfigPath, "kubeconfig", "c", "kubeconf.txt", "Path to save kubeconfig file")
	cmd.Flags().BoolVarP(&cfg.Debug, "debug", "d", false, "Debug")

	// Disable sorting of flags
	cmd.Flags().SortFlags = false

	// Hidden flags
	_ = cmd.Flags().MarkHidden("debug")
	_ = cmd.MarkFlagRequired("server")

	return &cmd
}
