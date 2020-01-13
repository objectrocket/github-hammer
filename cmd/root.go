// Copyright 2020 ObjectRocket, RackSpace
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	archiveFile     string
	includeArchived bool
	listLimit       int
	org             string
	token           string
)

var (
	rootCmd = &cobra.Command{
		Use:   "github-hammer",
		Short: "interact with github",
		Long: `Github Hammer is not meant to hammer github, but rather a hammer for
making changes to a large number of github repositories. It's also meant as a
reporting tool for gathering information from repositories.`}
)

func init() {
	// Must get the organization and github token
	viper.SetEnvPrefix("GITHUB")
	viper.AutomaticEnv()
	// organization to operate on
	rootCmd.PersistentFlags().StringVar(&org, "organization", "", "Organization to use when performing operations (or via env var: GITHUB_ORGANIZATION)")
	viper.BindPFlag("organization", rootCmd.PersistentFlags().Lookup("organization"))
	// token for github api
	rootCmd.PersistentFlags().StringVar(&token, "token", "", "Github API token for interaction with github. (or via env var: GITHUB_TOKEN)")
	viper.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token"))
}

//Execute begins the main application execution and exists appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// CheckRequiredFlags ensures required flags are set
func CheckRequiredFlags() error {
	requiredStringFlags := []string{"organization", "token"}
	for _, flag := range requiredStringFlags {
		if viper.GetString(flag) == "" {
			return errors.New("Required flag `" + flag + "` is not set.")
		}
	}
	return nil
}
