# github-hammer
This tool utilizes the Github V3 and V4 API's to make changes to a large number of repositories in an organization.

## Why publish the tool?

There was a surprising lack of tools made for manaing a large set of repositories. I presume many exist, and are created like this one originally was, as a one off tool for our own internal use. We're publishing this tool to help aid others who may have similar needs and as a reference for utilizing various GitHub API interfaces (the v3 and v4 API's are used by this application).

## Use Cases

Currently there are three primary functions
* archive reposistories
* enable security alerting
* generate a report of existing security alerts (outputs a format suitable for pasting into confluence...)

## Todo

* copy a repo and enforce default settings (template repos do not allow branch protection, etc.)
* report output improvements
* optimize archive functionality
* add tests
