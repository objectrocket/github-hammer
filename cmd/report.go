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
	"strings"

	"github.com/objectrocket/github-hammer/internal/ghammer"
	"github.com/shurcooL/githubv4"
	"github.com/spf13/cobra"
)

var (
	reportCmd = &cobra.Command{
		Use:           "report",
		Short:         "display vulnerability report",
		Long:          `Display a report of vulnerabilities, with information designed to make triage easy. Output suitable for pasting into confluence.`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReport()
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return CheckRequiredFlags()
		},
	}
)

func init() {
	rootCmd.AddCommand(reportCmd)
	reportCmd.Flags().BoolVar(&includeArchived, "include-archived", false, "Include archived repositories in report")
	reportCmd.Flags().IntVar(&listLimit, "limit", 5000, "Limit report to this many repositories")
}

// runReport generates the vulnerability report
func runReport() error {
	repoList, err := ghammer.GetRepoList(ghammer.RepoListOptions{Limit: listLimit, IncludeArchived: includeArchived})
	if err != nil {
		return err
	}

	// populate codeowners
	for i := 0; i < len(repoList); i++ {
		repoList[i].CodeOwners, err = ghammer.GetCodeOwners(repoList[i].Name, repoList[i].OwnerLogin)
		if err != nil {
			return err
		}
	}

	// display the report
	for _, repo := range repoList {
		repoDetail(repo)
	}

	return nil
}

func repoDetail(rInfo ghammer.RepoInfo) {
	client := ghammer.GetV4Client()

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
							Name      string
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
