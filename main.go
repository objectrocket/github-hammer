// Copyright 2019 ObjectRocket, Rackspace Inc,
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

package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"regexp"
	"strings"

	githubv3 "github.com/google/go-github/v28/github"
	githubv4 "github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type repoInfo struct {
	Name       string
	OwnerLogin string
	CodeOwners []string
}

func main() {
	var fullRepoList []*githubv3.Repository

	opt := &githubv3.RepositoryListByOrgOptions{ListOptions: githubv3.ListOptions{PerPage: 25}}

	// setup clients
	ctx := context.Background()
	tokenSrc := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	tokenClient := oauth2.NewClient(ctx, tokenSrc)

	clientV4 := githubv4.NewClient(tokenClient)
	clientV3 := githubv3.NewClient(tokenClient)

	// get a list of all the repositories, deal with pagination
	for {
		repos, response, err := clientV3.Repositories.ListByOrg(ctx, "objectrocket", opt)
		if err != nil {
			panic(err)
		}
		fullRepoList = append(fullRepoList, repos...)

		if response.NextPage == 0 {
			break
		}
		opt.Page = response.NextPage
	}

	// get a list of the repos to archive
	// archiveList, err := ioutil.ReadFile("archive.txt")
	// if err != nil {
	// 	panic(err)
	// }
	// reposToArchive := strings.Split(string(archiveList), "\n")

	// print the security vulnerability repo details
	for _, repo := range fullRepoList {

		// skip archived repositories
		if repo.GetArchived() {
			continue
		}

		// check if the repo is in archive list, and archive it if so.
		// for _, item := range reposToArchive {
		// 	if *repo.Name == item {
		// 		// would archive this here, so say it for now
		// 		fmt.Printf("archiving: %s\n", *repo.Name)
		// 		*repo.Archived = true
		// 		_, _, err := clientV3.Repositories.Edit(context.Background(), "objectrocket", *repo.Name, repo)
		// 		if err != nil {
		// 			fmt.Printf("Error with repo %s: %s", *repo.Name, err.Error())
		// 			continue
		// 		}
		// 		continue
		// 	}
		// }

		// short circut complete list when testing...
		// if *repo.Name != "infra-api" {
		// 	continue
		// }

		// Enable vulnerability alerts (uncomment to enable, already done)
		// _, err := clientV3.Repositories.EnableVulnerabilityAlerts(ctx, repo.Owner.GetLogin(), *repo.Name)
		// if err != nil {
		// 	panic(err)
		// }

		rInfo := repoInfo{OwnerLogin: repo.Owner.GetLogin(), Name: *repo.Name}
		codeOwners, err := getCodeOwners(clientV3, rInfo)
		if err != nil {
			panic(fmt.Sprintf("Error getting code owners for repo %s", *repo.Name))
		}
		rInfo.CodeOwners = codeOwners
		repoDetail(clientV4, rInfo)
	}

}

// getCodeOwner returns a slice of strings, that are the possible code owners for a repository
func getCodeOwners(client *githubv3.Client, rInfo repoInfo) (codeOwners []string, err error) {
	var codeOwnerFiles = [...]string{"CODEOWNERS", "docs/CODEOWNERS", ".github/CODEOWNERS"}
	err = nil
	codeOwners = make([]string, 0)

	for _, file := range codeOwnerFiles {
		ctx := context.Background()
		fileContents, _, response, err := client.Repositories.GetContents(ctx, rInfo.OwnerLogin, rInfo.Name, file, nil)
		if err != nil {
			if response.Response.StatusCode == 404 {
				// no code owners file at this location
				continue
			} else {
				// unknown error from github :(
				panic(err)
			}
		}

		// file matched
		decodedString, err := base64.StdEncoding.DecodeString(*fileContents.Content)
		if err != nil {
			panic(fmt.Sprintf("Error decoding codeowners from repo: %s\n", rInfo.Name))
		}

		// pull out the code owners from any valid lines, can be more than one per code owner file
		re := regexp.MustCompile(`\*\s+(?P<owner>\S+)`)
		// matchGroups := re.SubexpNames()
		match := re.FindAllStringSubmatch(string(decodedString), -1)
		for _, singleMatch := range match {
			codeOwners = append(codeOwners, fmt.Sprintf("%s (from /%s)", singleMatch[1], file))
		}
	}
	return
}

func repoDetail(client *githubv4.Client, rInfo repoInfo) {

	var firstQuery struct {
		Repository struct {
			Description        string
			Name               string
			VulnerabilityAlert struct {
				PageInfo struct {
					StartCursor string
					HasNextPage bool
					EndCursor   string
				}
				Nodes []struct {
					DismissedAt           *githubv4.DateTime
					DismissReason         string
					SecurityVulnerability struct {
						Advisory struct {
							Summary     string
							Description string
							References  []struct {
								URL string
							}
						}
						Severity               githubv4.SecurityAdvisorySeverity
						VulnerableVersionRange string
						Package                struct {
							Ecosystem githubv4.SecurityAdvisoryEcosystem
							Name      string
						}
					}
				}
			} `graphql:"vulnerabilityAlerts(first:10)"`
		} `graphql:"repository(owner:$repoOwner,name:$repoName)"`
	}

	var nextQuery struct {
		Repository struct {
			Description        string
			Name               string
			VulnerabilityAlert struct {
				PageInfo struct {
					StartCursor string
					HasNextPage bool
					EndCursor   string
				}
				Nodes []struct {
					DismissedAt           *githubv4.DateTime
					DismissReason         string
					SecurityVulnerability struct {
						Advisory struct {
							Summary     string
							Description string
							References  []struct {
								URL string
							}
						}
						Severity               githubv4.SecurityAdvisorySeverity
						VulnerableVersionRange string
						Package                struct {
							Ecosystem githubv4.SecurityAdvisoryEcosystem
							Name      stringittle smaller, and only contains repos we expect to be active.
						}
					}
				}
			} `graphql:"vulnerabilityAlerts(first:10,after:$cursor)"`
		} `graphql:"repository(owner:$repoOwner,name:$repoName)"`
	}

	variables := map[string]interface{}{
		"repoOwner": githubv4.String(rInfo.OwnerLogin),
		"repoName":  githubv4.String(rInfo.Name),
	}

	err := client.Query(context.Background(), &firstQuery, variables)
	if err != nil {
		panic(err)
	}

	if rInfo.CodeOwners == nil {
		rInfo.CodeOwners = append(rInfo.CodeOwners, "none")
	}

	alerts := firstQuery.Repository.VulnerabilityAlert.Nodes
	ittle smaller, and only contains repos we expect to be active.
	moreAlerts := firstQuery.Repository.VulnerabilityAlert.PageInfo.HasNextPage
	variables["cursor"] = githubv4.String(firstQuery.Repository.VulnerabilityAlert.PageInfo.EndCursor)

	for moreAlerts {
		err := client.Query(context.Background(), &nextQuery, variables)
		if err != nil {
			panic(err)
		}
		alerts = append(alerts, nextQuery.Repository.VulnerabilityAlert.Nodes...)
		moreAlerts = nextQuery.Repository.VulnerabilityAlert.PageInfo.HasNextPage
		variables["cursor"] = githubv4.String(nextQuery.Repository.VulnerabilityAlert.PageInfo.EndCursor)
	}

	if len(alerts) > 0 {
		fmt.Printf("\n## Repository: %s\n", firstQuery.Repository.Name)
		fmt.Printf("**Description**: %s\n", firstQuery.Repository.Description)
		fmt.Printf("**Code Owners**: %s\n", strings.Join(rInfo.CodeOwners, ","))
		for _, alert := range alerts {
			advisoryRefs := make([]string, 0)
			if alert.DismissedAt == nil {
				for _, ref := range alert.SecurityVulnerability.Advisory.References {
					advisoryRefs = append(advisoryRefs, ref.URL)
				}
				fmt.Printf("* `%s` %s `%s` `%s` %s\n",
					alert.SecurityVulnerability.Package.Ecosystem,
					alert.SecurityVulnerability.Package.Name,
					alert.SecurityVulnerability.VulnerableVersionRange,
					alert.SecurityVulnerability.Severity,
					strings.Join(advisoryRefs, " , "))
			}
		}
		fmt.Printf("---")
	}
}
ittle smaller, and only contains repos we expect to be active.ittle smaller, and only contains repos we expect to be active.
