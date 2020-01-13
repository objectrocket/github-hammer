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
	"context"
	"fmt"

	"github.com/objectrocket/github-hammer/internal/ghammer"
	"github.com/spf13/cobra"
)

var (
	scannerCmd = &cobra.Command{
		Use:           "scanner",
		Short:         "enable vulnerability scanning",
		Long:          `enable vulnerability scanning`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runScanner()
		},
	}
)

func init() {
	rootCmd.AddCommand(scannerCmd)
	scannerCmd.Flags().IntVar(&listLimit, "limit", 5000, "Limit action to this many repositories")
	includeArchived = false
}

// runScanner executes the command to enable vulnerability reports for all repositories
func runScanner() (err error) {
	ctx := context.Background()
	client := ghammer.GetV3Client()

	repoList, err := ghammer.GetRepoList(ghammer.RepoListOptions{Limit: listLimit})
	if err != nil {
		return err
	}

	for _, repo := range repoList {
		_, err = client.Repositories.EnableVulnerabilityAlerts(ctx, repo.OwnerLogin, repo.Name)
		if err != nil {
			return err
		}
		fmt.Printf("Vulnerability alerts are enabled for: %s\n", repo.Name)

	}

	return nil

}
