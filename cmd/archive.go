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
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/objectrocket/github-hammer/internal/ghammer"
	"github.com/spf13/cobra"
)

var (
	archiveCmd = &cobra.Command{
		Use:   "archive",
		Short: "archive repositories",
		Long: `Set the status of a list of repositories to archived via the Github API, either
		a file with one repo per line, or a list of repositories as arguments to the program can
		be specified.`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runArchive(args)
		},
	}
)

func init() {
	rootCmd.AddCommand(archiveCmd)
	archiveCmd.Flags().StringVar(&archiveFile, "file", "", "path to file containing list of repos to archive (one repo per line)")
	includeArchived = false
	listLimit = 10
}

func runArchive(args []string) (err error) {
	var reposToArchive []string
	fmt.Printf("Len of args: %v\n", len(args))
	if archiveFile != "" {
		archiveList, err := ioutil.ReadFile(archiveFile)
		if err != nil {
			return err
		}
		reposToArchive = strings.Split(string(archiveList), "\n")
	} else if len(args) < 1 {
		return errors.New("Must either supply a file, or list of repos to archive")
	} else {
		reposToArchive = args
	}

	// not terribly efficient to get a list of all repositories for this, huge room for optimization
	repoList, err := ghammer.GetRepoList(ghammer.RepoListOptions{Limit: listLimit, IncludeArchived: includeArchived})
	if err != nil {
		return err
	}

	client := ghammer.GetV3Client()
	for i := 0; i < len(repoList); i++ {
		for _, item := range reposToArchive {
			if item == "" {
				continue
			}
			if repoList[i].Name == item {
				// would archive this here, so say it for now
				fmt.Printf("archiving: %s\n", repoList[i].Name)
				*repoList[i].Repo.Archived = true
				_, _, err := client.Repositories.Edit(context.Background(), org, repoList[i].Name, repoList[i].Repo)
				if err != nil {
					return err
				}
				continue
			}
		}
	}
	return nil
}
