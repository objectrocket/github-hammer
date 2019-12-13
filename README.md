# github-hammer
This tool utilizes the Github V3 and V4 API's to make changes to a large number of repositories in an organization.

## Why publish the tool?

There was a surprising lack of tools made for manaing a large set of repositories. I presume many exist, and are created like this one originally was, as a one off tool for our own internal use. We're publishing this tool to help aid others who may have similar needs.

## Use Cases

Currently there are three primary functions
* archive reposistories
* enable security alerting
* generate a report of existing security alerts (outputs a format suitable for pasting into confluence...)

## Todo

As mentioned this tool was created initially for our own internal use, and is not ready for general use. It's pre-alpha, it doesn't operate like a proper CLI tool, as it currently just uncommented/commented code for certain functions when those were desired. We'll update this shortly to make it a more properly behaving CLI app.
