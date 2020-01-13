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
	githubv3 "github.com/google/go-github/v28/github"
)

// RepoInfo contains basic information about a repository, as well as the repository object
type RepoInfo struct {
	Name       string
	OwnerLogin string
	CodeOwners []string
	Repo       *githubv3.Repository
}

// RepoListOptions controls how the repository list is returned
type RepoListOptions struct {
	Limit           int
	IncludeArchived bool
}
