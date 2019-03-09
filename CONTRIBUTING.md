We at QLC Chain love reaching out to the open-source community and are open to accepting issues and pull-requests.

# About the code base

It is written in [Golang](https://golang.org). And we use [Modules](https://github.com/golang/go/wiki/Modules) for versioned modules. 

## Development Prerequisites

## Clone repository and build
```bash
git clone https://github.com/qlcchain/go-qlc
cd go-qlc
make build
```

## Running unit tests

Run `./ci/traivs/travis.codecov.sh` to execute the unit tests.

# GitHub and working with the team

If you have an idea of an improvement or new feature, consider discussing it first with the team by adding an issue. Maybe someone is already working on it, or have suggestions on how to improve on the idea. 

and there are the branch protection rules:

- master
  - Require pull request reviews before merging
  - Require status checks to pass before merging
  - Require signed commits

- develop
  - Require pull request reviews before merging
  - Require status checks to pass before merging

## Fork and do all your work on a branch
We prefer the standard GitHub workflow. You create a fork of the QLCWallet repository, make branches for features/issues, and commit and push to the develop branch. 

## Create pull requests
Before:
* Review your code locally. Have you followed the guidelines in this document?
* Run tests. Did you consider adding a test case for your feature?
* Commit and push your fork
* Create pull request on the upstream repository:
    * Make sure you add a description that clearly describes the purpose of the PR.
    * If the PR solves one or more issues, please reference these in the description.

After:
* Check that CI completes successfully. If not, fix the problem and push an update.
* Respond to comments and reviews in a timely fashion.

## Resolve conflicts

If time passes between your pull request (PR) submission and the team accepting it, merge conflicts may occur due to activity on master, such as merging other PR's before yours. In order for your PR to be accepted, you must resolve these conflicts.

The preferred process is to rebase your changes, resolve any conflicts, and push your changes again. <sup>[1](#git_rebase_conflicts)</sup> <sup>[2](#git_merge_conflicts)</sup>

* Check out your branch
* git fetch upstream
* git rebase upstream/master
* Resolve conflicts in your favorite editor
* git add filename
* git rebase --continue
* Commit and push your branch

## Consider squashing or amending commits

In the review process, you're likely to get feedback. You'll commit and push more changes, get more feedback, etc. 

This can lead to a messy git history, and can make stuff like bisecting harder.

Once your PR is OK'ed, please squash the commits into a one <sup>[3](#git_squash)</sup> if there are many "noisy" commits with minor changes and fixups. 

If the individual commits are clean and logically separate, then no squashing is necessary.

Note that you can also update the last commit with `git commit --amend`. Say your last commit had a typo. Instead of committing and having to squash it later, simply commit with amend and push the branch.

# Code standard

For all code contributions, please ensure they adhere as close as possible to the following guidelines:

- **Strictly** follows the formatting and styling rules denoted [here](https://github.com/golang/go/wiki/CodeReviewComments).
- Commit messages are in the format `module_name: Change typed down as a sentence.`  
This allows our maintainers and everyone else to know what specific code changes you wish to address.
    - `wallet: Implemented wallet manager.`
    - `types/hash: Add hash.toString().`
- Consider backwards compatibility. New methods are perfectly fine, though changing the existing public API should only be done should there be a good reason.
