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

package ghammer

import (
	"context"
	"encoding/base64"
	"fmt"
	"regexp"

	githubv3 "github.com/google/go-github/v28/github"
	"github.com/shurcooL/githubv4"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

// GetV3Client returns a GitHub v3 API client
func GetV3Client() (client *githubv3.Client) {
	ctx := context.Background()
	tokenSrc := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: viper.GetString("token")},
	)
	tokenClient := oauth2.NewClient(ctx, tokenSrc)
	client = githubv3.NewClient(tokenClient)
	return
}

// GetV4Client returns a GitHub v3 API client
func GetV4Client() (client *githubv4.Client) {
	ctx := context.Background()
	tokenSrc := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: viper.GetString("token")},
	)
	tokenClient := oauth2.NewClient(ctx, tokenSrc)
	client = githubv4.NewClient(tokenClient)

	return
}

// GetRepoList returns a slice of RepoInfo objects
func GetRepoList(listOpts RepoListOptions) (repoList []RepoInfo, err error) {
	clientV3 := GetV3Client()
	opt := &githubv3.RepositoryListByOrgOptions{ListOptions: githubv3.ListOptions{PerPage: 25}}
	ctx := context.Background()

	for {
		var repos []*githubv3.Repository
		var response *githubv3.Response
		repos, response, err = clientV3.Repositories.ListByOrg(ctx, viper.GetString("organization"), opt)
		if err != nil {
			return nil, err
		}

		for _, repo := range repos {
			// skip archived repositories
			if !listOpts.IncludeArchived && *repo.Archived {
				continue
			}

			repoList = append(repoList, RepoInfo{
				Name:       *repo.Name,
				OwnerLogin: repo.Owner.GetLogin(),
				Repo:       repo,
			})

			if len(repoList) >= listOpts.Limit {
				return
			}
		}

		if response.NextPage == 0 {
			break
		}
		opt.Page = response.NextPage
	}
	return
}

// GetCodeOwners returns a slice of strings, that are the possible code owners for a repository
func GetCodeOwners(repoName string, ownerLogin string) (codeOwners []string, err error) {
	var codeOwnerFiles = [...]string{"CODEOWNERS", "docs/CODEOWNERS", ".github/CODEOWNERS"}
	err = nil
	codeOwners = make([]string, 0)
	ctx := context.Background()
	client := GetV3Client()

	for _, file := range codeOwnerFiles {
		fileContents, _, response, err := client.Repositories.GetContents(ctx, ownerLogin, repoName, file, nil)
		if err != nil {
			if response.Response.StatusCode == 404 {
				// no code owners file at this location
				continue
			} else {
				// unknown error from github :(
				return nil, err
			}
		}

		// file matched
		decodedString, err := base64.StdEncoding.DecodeString(*fileContents.Content)
		if err != nil {
			panic(fmt.Sprintf("Error decoding codeowners from repo: %s\n", repoName))
		}

		// pull out the code owners from any valid lines, can be more than one per code owner file
		re := regexp.MustCompile(`\*\s+(?P<owner>\S+)`)
		match := re.FindAllStringSubmatch(string(decodedString), -1)
		for _, singleMatch := range match {
			codeOwners = append(codeOwners, fmt.Sprintf("%s (from /%s)", singleMatch[1], file))
		}
	}
	return
}
