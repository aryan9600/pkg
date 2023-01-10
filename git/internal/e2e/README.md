# Git E2E Tests

This folder contains E2E tests for `pkg/git/gogit` and `pkg/git/libgit2`. The current
tests are run against the following providers:

* [GitHub](https://github.com)
* [Gitlab](https://gitlab.com)
* GitLab CE (self-hosted)
* Bitbucket Server (self-hosted)
* Gitkit (a test Git server to test custom configuration)

## Usage

### Gitkit
The Git server runs in process, hence there are no user inputs required.
```
GO_TESTS='-run TestGitKitE2E' ./run.sh
```

### GitLab CE
Gitlab CE is run inside a Docker container and all necessary authentication credentials
are automatically provided to the tests, hence no user inputs are required.
```
GO_TESTS='-run TestGitLabCEE2E' ./run.sh
```

### GitHub
You need to create a PAT associated with your account. You can do so by following this
[guide](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token).
Specify the token, username and org name as environment variables for the script. Please make sure that the
org already exists as it won't be created by the script itself.
```
GO_TESTS='-run TestGitHubE2E' GITHUB_USER='***' GITHUB_ORG='***' GITHUB_TOKEN='***' ./run.sh 
```

### GitLab
You need to create an access token associated with your account. You can do so by following this
[guide](https://docs.gitlab.com/ee/user/project/settings/project_access_tokens.html).
Specify the token, username and group name as environment variables for the script. Please make sure that the
group already exists as it won't be created by the script itself.
```
GO_TESTS='-run TestGitHubE2E' GITLAB_USER='***' GITLAB_GROUP='***' GITLAB_PAT='***' ./run.sh 
```

### Bitbucket Server
You need to create an HTTP Access Token associated with your account. You can do so by following this
[guide](https://confluence.atlassian.com/bitbucketserver/personal-access-tokens-939515499.html).
Specify the token, username, project key and the domain where your server can be reached as
environment variables for the script. Please make sure that the project already exists as it
won't be created by the script itself.
```
GO_TESTS='-run TestBitbucketServerE2E' STASH_USER='***' STASH_TOKEN='***' STASH_DOMAIN='***' STASH_PROJECT_KEY='***' ./run.sh
```
