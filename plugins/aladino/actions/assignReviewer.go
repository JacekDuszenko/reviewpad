// Copyright 2022 Explore.dev Unipessoal Lda. All Rights Reserved.
// Use of this source code is governed by a license that can be
// found in the LICENSE file.

package plugins_aladino_actions

import (
	"fmt"
	"log"

	"github.com/google/go-github/v42/github"
	"github.com/reviewpad/reviewpad/v2/lang/aladino"
	"github.com/reviewpad/reviewpad/v2/utils"
)

func AssignReviewer() *aladino.BuiltInAction {
	return &aladino.BuiltInAction{
		Type: aladino.BuildFunctionType([]aladino.Type{aladino.BuildArrayOfType(aladino.BuildStringType()), aladino.BuildIntType()}, nil),
		Code: assignReviewerCode,
	}
}

func assignReviewerCode(e aladino.Env, args []aladino.Value) error {
	if len(args) < 1 {
		return fmt.Errorf("assignReviewer: expecting at least 1 argument")
	}

	arg := args[0]
	if !arg.HasKindOf(aladino.ARRAY_VALUE) {
		return fmt.Errorf("assignReviewer: requires array argument, got %v", arg.Kind())
	}

	if !args[1].HasKindOf(aladino.INT_VALUE) {
		return fmt.Errorf("assignReviewer: the parameter total is required to be an int, instead got %v", args[1].Kind())
	}

	totalRequiredReviewers := args[1].(*aladino.IntValue).Val

	availableReviewers := arg.(*aladino.ArrayValue).Vals

	for _, reviewer := range availableReviewers {
		if !reviewer.HasKindOf(aladino.STRING_VALUE) {
			return fmt.Errorf("assignReviewer: requires array of strings, got array with value of %v", reviewer.Kind())
		}
	}

	// Remove pull request author from provided reviewers list
	for index, reviewer := range availableReviewers {
		if reviewer.(*aladino.StringValue).Val == *e.GetPullRequest().User.Login {
			availableReviewers = append(availableReviewers[:index], availableReviewers[index+1:]...)
			break
		}
	}

	totalAvailableReviewers := len(availableReviewers)
	if totalRequiredReviewers > totalAvailableReviewers {
		log.Printf("assignReviewer: total required reviewers %v exceeds the total available reviewers %v", totalRequiredReviewers, totalAvailableReviewers)
		totalRequiredReviewers = totalAvailableReviewers
	}

	prNum := utils.GetPullRequestNumber(e.GetPullRequest())
	owner := utils.GetPullRequestOwnerName(e.GetPullRequest())
	repo := utils.GetPullRequestRepoName(e.GetPullRequest())

	reviewers := []string{}

	reviews, _, err := e.GetClient().PullRequests.ListReviews(e.GetCtx(), owner, repo, prNum, nil)
	if err != nil {
		return err
	}

	// Re-request current reviewers if mention on the provided reviewers list
	for _, review := range reviews {
		for index, availableReviewer := range availableReviewers {
			if availableReviewer.(*aladino.StringValue).Val == *review.User.Login {
				totalRequiredReviewers--
				reviewers = append(reviewers, *review.User.Login)
				availableReviewers = append(availableReviewers[:index], availableReviewers[index+1:]...)
				break
			}
		}
	}

	// Skip current requested reviewers if mention on the provided reviewers list
	currentRequestedReviewers := e.GetPullRequest().RequestedReviewers
	for _, requestedReviewer := range currentRequestedReviewers {
		for index, availableReviewer := range availableReviewers {
			if availableReviewer.(*aladino.StringValue).Val == *requestedReviewer.Login {
				totalRequiredReviewers--
				availableReviewers = append(availableReviewers[:index], availableReviewers[index+1:]...)
				break
			}
		}
	}

	// Select random reviewers from the list of all provided reviewers
	for i := 0; i < totalRequiredReviewers; i++ {
		selectedElementIndex := utils.GenerateRandom(len(availableReviewers))

		selectedReviewer := availableReviewers[selectedElementIndex]
		availableReviewers = append(availableReviewers[:selectedElementIndex], availableReviewers[selectedElementIndex+1:]...)

		reviewers = append(reviewers, selectedReviewer.(*aladino.StringValue).Val)
	}

	if len(reviewers) == 0 {
		log.Printf("assignReviewer: skipping request reviewers. the pull request already has reviewers")
		return nil
	}

	_, _, err = e.GetClient().PullRequests.RequestReviewers(e.GetCtx(), owner, repo, prNum, github.ReviewersRequest{
		Reviewers: reviewers,
	})

	return err
}