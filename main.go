package main

import (
	"context"
	"fmt"
	"github.com/google/go-github/v18/github"
	"golang.org/x/oauth2"
	"os"
)

const (
	DEFAULT_BRANCH = "master"
)

func main() {

	// Make sure no mandatory environment variables are missing.
	for _, envVar := range []string{
		"ACCESS_TOKEN",
	} {
		if os.Getenv(envVar) == "" {
			panic(fmt.Sprintf("Missing environment variable! %s", envVar))
		}
	}
	token := os.Getenv("ACCESS_TOKEN")

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 10},
	}
	// get all pages of results
	var allRepos []*github.Repository

	for {
		repos, resp, err := client.Repositories.ListByOrg(ctx, "randomtestorg", opt)
		if err != nil {
			panic(err)
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	for i := 0; i < len(allRepos); i++ {
		repository := allRepos[i]
		updateBranchProtection(repository, client, ctx)
	}
}

func updateBranchProtection(repo *github.Repository, client *github.Client, ctx context.Context) {
	fmt.Printf("Branch protection: %s \n", repo.GetName())
	slice := make([]string, 0)

	statusCheck := &github.RequiredStatusChecks{
		Strict:   true,
		Contexts: slice, // empty list
	}
	restrictionsRequest := &github.DismissalRestrictionsRequest{
		Users: nil, // empty list
		Teams: nil, // empty list
	}
	pullRequestEnforcemnet := &github.PullRequestReviewsEnforcementRequest{
		DismissalRestrictionsRequest: restrictionsRequest,
		DismissStaleReviews:          false,
		RequiredApprovingReviewCount: 2,
	}
	userRestrictions := &github.BranchRestrictionsRequest{
		Users: slice,
		Teams: slice,
	}
	pr := &github.ProtectionRequest{
		RequiredStatusChecks:       statusCheck,
		RequiredPullRequestReviews: pullRequestEnforcemnet,
		EnforceAdmins:              false,
		Restrictions:               userRestrictions,
	}

	_, _, e := client.Repositories.UpdateBranchProtection(ctx, repo.GetOwner().GetLogin(), repo.GetName(), DEFAULT_BRANCH, pr)

	if e != nil {
		fmt.Errorf("Failed to update", e)
	}
}