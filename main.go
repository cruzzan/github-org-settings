package main

import (
	"context"
	"fmt"
	"github.com/google/go-github/v21/github"
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
		"GITHUB_ORG",
	} {
		if os.Getenv(envVar) == "" {
			panic(fmt.Sprintf("Missing environment variable! %s", envVar))
		}
	}
	token := os.Getenv("ACCESS_TOKEN")
	organizationName := os.Getenv("GITHUB_ORG")

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
		repos, resp, err := client.Repositories.ListByOrg(ctx, organizationName, opt)
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
		editRepository(repository, client, ctx)
	}
}
func Bool(v bool) *bool       { return &v }
func String(v string) *string { return &v }

func editRepository(repository *github.Repository, client *github.Client, ctx context.Context) {

	input := &github.Repository{
		HasIssues:        Bool(false),
		DefaultBranch:    String(DEFAULT_BRANCH),
		MasterBranch:     String(DEFAULT_BRANCH),
		AllowRebaseMerge: Bool(true),
		AllowSquashMerge: Bool(true),
		AllowMergeCommit: Bool(true),
	}
	fmt.Printf("Updating repository settings for %s \n", repository.GetName())

	_, _, err := client.Repositories.Edit(ctx, repository.GetOwner().GetLogin(), repository.GetName(), input)
	if err != nil {
		fmt.Errorf("Failed to update repository", err)
	}
}

func updateBranchProtection(repo *github.Repository, client *github.Client, ctx context.Context) {
	fmt.Printf("Branch protection: %s \n", repo.GetName())
	slice := []string{}

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
		EnforceAdmins:              true,
		Restrictions:               userRestrictions,
	}

	_, _, e := client.Repositories.UpdateBranchProtection(ctx, repo.GetOwner().GetLogin(), repo.GetName(), DEFAULT_BRANCH, pr)
	if e != nil {
		fmt.Errorf("Failed to update branch protection", e)
	}
	sreq := &github.RequiredStatusChecksRequest{
		Strict:   Bool(true),
		Contexts: slice,
	}
	_, _, error := client.Repositories.UpdateRequiredStatusChecks(ctx, repo.GetOwner().GetLogin(), repo.GetName(), DEFAULT_BRANCH, sreq)
	if error != nil {
		fmt.Errorf("require status check", e)
	}

	_, _, errrr := client.Repositories.RequireSignaturesOnProtectedBranch(ctx, repo.GetOwner().GetLogin(), repo.GetName(), DEFAULT_BRANCH)
	if errrr != nil {
		fmt.Errorf("Failed to update branch protection", errrr)
	}
}
