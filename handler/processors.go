// Copyright (C) 2022 Explore.dev Unipessoal Lda. All Rights Reserved.
// Use of this source code is governed by a license that can be
// found in the LICENSE file.

package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v52/github"
	"github.com/reviewpad/go-lib/entities"
	reviewpad_gh "github.com/reviewpad/reviewpad/v4/codehost/github"
	"github.com/sirupsen/logrus"
)

func ParseEvent(log *logrus.Entry, rawEvent string) (*ActionEvent, error) {
	event := &ActionEvent{}

	log.Infof("parsing event %v", rawEvent)

	err := json.Unmarshal([]byte(rawEvent), &event)
	if err != nil {
		return nil, err
	}

	return event, nil
}

func processUnsupportedEvent(eventPayload interface{}) ([]*entities.TargetEntity, *entities.EventDetails, error) {
	return nil, nil, fmt.Errorf("unsupported event payload type: %T", eventPayload)
}

func processCronEvent(log *logrus.Entry, token string, e *ActionEvent) ([]*entities.TargetEntity, *entities.EventDetails, error) {
	log.Info("processing schedule event")

	ctx, canc := context.WithTimeout(context.Background(), time.Minute*10)
	defer canc()

	ghClient := reviewpad_gh.NewGithubClientFromToken(ctx, token)

	repoParts := strings.SplitN(*e.Repository, "/", 2)

	owner := repoParts[0]
	repo := repoParts[1]

	repository, _, err := ghClient.GetClientREST().Repositories.Get(ctx, owner, repo)
	if err != nil {
		return nil, nil, fmt.Errorf("get repository failed: %w", err)
	}

	issues, _, err := ghClient.ListIssuesByRepo(ctx, owner, repo, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("get pull requests: %w", err)
	}

	log.Infof("fetched %d issues", len(issues))

	targets := make([]*entities.TargetEntity, 0)
	for _, issue := range issues {
		kind := entities.Issue
		if issue.IsPullRequest() {
			kind = entities.PullRequest
		}
		targets = append(targets, &entities.TargetEntity{
			Kind:        kind,
			Number:      *issue.Number,
			Owner:       owner,
			Repo:        repo,
			AccountType: repository.GetOwner().GetType(),
			Visibility:  repository.GetVisibility(),
		})
	}

	log.Infof("found events %v", targets)

	return targets, &entities.EventDetails{
		EventName: *e.EventName,
	}, nil
}

func processIssuesEvent(log *logrus.Entry, e *github.IssuesEvent) ([]*entities.TargetEntity, *entities.EventDetails, error) {
	log.Info("processing issues event")
	log.Infof("found issue %v", *e.Issue.Number)

	return []*entities.TargetEntity{
			{
				Kind:        entities.Issue,
				Number:      *e.Issue.Number,
				Owner:       *e.Repo.Owner.Login,
				Repo:        *e.Repo.Name,
				AccountType: e.GetRepo().GetOwner().GetType(),
				Visibility:  e.GetRepo().GetVisibility(),
			},
		}, &entities.EventDetails{
			EventName:   "issues",
			EventAction: *e.Action,
			Payload:     e,
		}, nil
}

func processIssueCommentEvent(log *logrus.Entry, e *github.IssueCommentEvent) ([]*entities.TargetEntity, *entities.EventDetails, error) {
	log.Info("processing issue_comment event")
	log.Infof("found issue %v", *e.Issue.Number)

	kind := entities.Issue
	if e.Issue.IsPullRequest() {
		kind = entities.PullRequest
	}

	return []*entities.TargetEntity{
			{
				Kind:        kind,
				Number:      *e.Issue.Number,
				Owner:       *e.Repo.Owner.Login,
				Repo:        *e.Repo.Name,
				AccountType: e.GetRepo().GetOwner().GetType(),
				Visibility:  e.GetRepo().GetVisibility(),
			},
		}, &entities.EventDetails{
			EventName:   "issue_comment",
			EventAction: *e.Action,
			Payload:     e,
		}, nil
}

func processDiscussionEvent(log *logrus.Entry, e *github.DiscussionEvent) ([]*entities.TargetEntity, *entities.EventDetails, error) {
	log.Info("processing discussion event")
	log.Infof("found discussion %v", *e.Discussion.Number)

	return []*entities.TargetEntity{
			{
				Kind:        entities.Discussion,
				Number:      *e.Discussion.Number,
				Owner:       *e.Repo.Owner.Login,
				Repo:        *e.Repo.Name,
				AccountType: e.GetRepo().GetOwner().GetType(),
				Visibility:  e.GetRepo().GetVisibility(),
			},
		}, &entities.EventDetails{
			EventName:   "discussion",
			EventAction: *e.Action,
			Payload:     e,
		}, nil
}

func processDiscussionCommentEvent(log *logrus.Entry, e *github.DiscussionCommentEvent) ([]*entities.TargetEntity, *entities.EventDetails, error) {
	log.Info("processing discussion_comment event")
	log.Infof("found discussion %v", *e.Discussion.Number)

	return []*entities.TargetEntity{
			{
				Kind:        entities.Discussion,
				Number:      *e.Discussion.Number,
				Owner:       *e.Repo.Owner.Login,
				Repo:        *e.Repo.Name,
				AccountType: e.GetRepo().GetOwner().GetType(),
				Visibility:  e.GetRepo().GetVisibility(),
			},
		}, &entities.EventDetails{
			EventName:   "discussion_comment",
			EventAction: *e.Action,
			Payload:     e,
		}, nil
}

func processPullRequestEvent(log *logrus.Entry, e *github.PullRequestEvent) ([]*entities.TargetEntity, *entities.EventDetails, error) {
	log.Info("processing pull_request event")
	log.Infof("found pull request %v", *e.PullRequest.Number)

	return []*entities.TargetEntity{
			{
				Kind:        entities.PullRequest,
				Number:      *e.PullRequest.Number,
				Owner:       *e.Repo.Owner.Login,
				Repo:        *e.Repo.Name,
				AccountType: e.GetRepo().GetOwner().GetType(),
				Visibility:  e.GetRepo().GetVisibility(),
			},
		}, &entities.EventDetails{
			EventName:   "pull_request",
			EventAction: *e.Action,
			Payload:     e,
		}, nil
}

func processPullRequestReviewEvent(log *logrus.Entry, e *github.PullRequestReviewEvent) ([]*entities.TargetEntity, *entities.EventDetails, error) {
	log.Info("processing pull_request_review event")
	log.Infof("found pull request %v", *e.PullRequest.Number)

	return []*entities.TargetEntity{
			{
				Kind:        entities.PullRequest,
				Number:      *e.PullRequest.Number,
				Owner:       *e.Repo.Owner.Login,
				Repo:        *e.Repo.Name,
				AccountType: e.GetRepo().GetOwner().GetType(),
				Visibility:  e.GetRepo().GetVisibility(),
			},
		}, &entities.EventDetails{
			EventName:   "pull_request_review",
			EventAction: *e.Action,
			Payload:     e,
		}, nil
}

func processPullRequestReviewCommentEvent(log *logrus.Entry, e *github.PullRequestReviewCommentEvent) ([]*entities.TargetEntity, *entities.EventDetails, error) {
	log.Info("processing pull_request_review_comment event")
	log.Infof("found pull request %v", *e.PullRequest.Number)

	return []*entities.TargetEntity{
			{
				Kind:        entities.PullRequest,
				Number:      *e.PullRequest.Number,
				Owner:       *e.Repo.Owner.Login,
				Repo:        *e.Repo.Name,
				AccountType: e.GetRepo().GetOwner().GetType(),
				Visibility:  e.GetRepo().GetVisibility(),
			},
		}, &entities.EventDetails{
			EventName:   "pull_request_review_comment",
			EventAction: *e.Action,
			Payload:     e,
		}, nil
}

func processPullRequestTargetEvent(log *logrus.Entry, e *github.PullRequestTargetEvent) ([]*entities.TargetEntity, *entities.EventDetails, error) {
	log.Infof(`processing "pull_request_target" event`)
	log.Infof("found pull request %v", *e.PullRequest.Number)

	return []*entities.TargetEntity{
			{
				Kind:        entities.PullRequest,
				Number:      *e.PullRequest.Number,
				Owner:       *e.Repo.Owner.Login,
				Repo:        *e.Repo.Name,
				AccountType: e.GetRepo().GetOwner().GetType(),
				Visibility:  e.GetRepo().GetVisibility(),
			},
		}, &entities.EventDetails{
			EventName:   "pull_request_target",
			EventAction: *e.Action,
			Payload:     e,
		}, nil
}

func processStatusEvent(log *logrus.Entry, token string, e *github.StatusEvent) ([]*entities.TargetEntity, *entities.EventDetails, error) {
	log.Info("processing status event")

	ctx, canc := context.WithTimeout(context.Background(), time.Minute*10)
	defer canc()

	ghClient := reviewpad_gh.NewGithubClientFromToken(ctx, token)

	prs, err := ghClient.GetPullRequests(ctx, *e.Repo.Owner.Login, *e.Repo.Name)
	if err != nil {
		return nil, nil, fmt.Errorf("get pull requests: %w", err)
	}

	log.Infof("fetched %v pull requests", len(prs))

	eventDetails := &entities.EventDetails{
		EventName: "status",
		Payload:   e,
	}

	for _, pr := range prs {
		if *pr.Head.SHA == *e.SHA {
			log.Infof("found pull request %v", *pr.Number)
			return []*entities.TargetEntity{
				{
					Kind:        entities.PullRequest,
					Number:      *pr.Number,
					Owner:       *pr.Base.Repo.Owner.Login,
					Repo:        *pr.Base.Repo.Name,
					AccountType: pr.GetBase().GetRepo().GetOwner().GetType(),
					Visibility:  pr.GetBase().GetRepo().GetVisibility(),
				},
			}, eventDetails, nil
		}
	}

	log.Infof("no pr found with the head sha %v", *e.SHA)

	return []*entities.TargetEntity{}, eventDetails, nil
}

func processWorkflowRunEvent(log *logrus.Entry, token string, e *github.WorkflowRunEvent) ([]*entities.TargetEntity, *entities.EventDetails, error) {
	log.Info("processing workflow_run event")

	ctx, canc := context.WithTimeout(context.Background(), time.Minute*10)
	defer canc()
	ghClient := reviewpad_gh.NewGithubClientFromToken(ctx, token)

	prs, err := ghClient.GetPullRequests(ctx, *e.Repo.Owner.Login, *e.Repo.Name)
	if err != nil {
		return nil, nil, fmt.Errorf("get pull requests: %w", err)
	}

	eventDetail := &entities.EventDetails{
		EventName:   "workflow_run",
		EventAction: e.GetAction(),
		Payload:     e,
	}

	log.Infof("fetched %v prs", len(prs))

	for _, pr := range prs {
		if *pr.Head.SHA == *e.WorkflowRun.HeadSHA {
			log.Infof("found pull request %v", *pr.Number)
			return []*entities.TargetEntity{
				{
					Kind:        entities.PullRequest,
					Number:      *pr.Number,
					Owner:       *pr.Base.Repo.Owner.Login,
					Repo:        *pr.Base.Repo.Name,
					AccountType: pr.GetBase().GetRepo().GetOwner().GetType(),
					Visibility:  pr.GetBase().GetRepo().GetVisibility(),
				},
			}, eventDetail, nil
		}
	}

	log.Infof("no pr found with the head sha %v", *e.WorkflowRun.HeadSHA)

	return []*entities.TargetEntity{}, eventDetail, nil
}

func processPushEvent(log *logrus.Entry, token string, e *github.PushEvent) ([]*entities.TargetEntity, *entities.EventDetails, error) {
	log.Info("processing push event")

	ctx, canc := context.WithTimeout(context.Background(), time.Minute*10)
	defer canc()

	ghClient := reviewpad_gh.NewGithubClientFromToken(ctx, token)

	repoParts := strings.SplitN(*e.GetRepo().FullName, "/", 2)

	owner := repoParts[0]
	repo := repoParts[1]

	prs, err := ghClient.GetPullRequests(ctx, owner, repo)
	if err != nil {
		return nil, nil, fmt.Errorf("get pull requests: %w", err)
	}

	log.Infof("fetched %d pull requests", len(prs))

	targets := make([]*entities.TargetEntity, 0)
	for _, pr := range prs {
		if fmt.Sprintf("refs/heads/%s", pr.GetHead().GetRef()) == e.GetRef() {
			targets = append(targets, &entities.TargetEntity{
				Kind:        entities.PullRequest,
				Number:      *pr.Number,
				Owner:       owner,
				Repo:        repo,
				AccountType: pr.GetBase().GetRepo().GetOwner().GetType(),
				Visibility:  pr.GetBase().GetRepo().GetVisibility(),
			})
		}
	}

	// since the push event is not necessarily tied to a pull request, for example when you push into a default branch
	// we also need to add the repo as a target so we can handle push events that are not tied to a pull request
	repoTarget := &entities.TargetEntity{
		Kind:        entities.Repository,
		Owner:       e.GetRepo().GetOwner().GetLogin(),
		Repo:        e.GetRepo().GetName(),
		AccountType: e.GetRepo().GetOwner().GetType(),
		Visibility:  "public",
	}

	if e.GetRepo().GetPrivate() {
		repoTarget.Visibility = "private"
	}

	targets = append(targets, repoTarget)

	log.Infof("found events %v", targets)

	return targets, &entities.EventDetails{
		EventName: "push",
		Payload:   e,
	}, nil
}

func processInstallationEvent(log *logrus.Entry, event *github.InstallationEvent) ([]*entities.TargetEntity, *entities.EventDetails, error) {
	log.Info("processing installation event")

	targetEntities, err := extractTargetEntitiesFromRepositories(event.Repositories, event.GetInstallation().GetAccount().GetType())
	if err != nil {
		return nil, nil, err
	}

	return targetEntities, &entities.EventDetails{
		EventName:   "installation",
		EventAction: event.GetAction(),
		Payload:     event,
	}, nil
}

func processInstallationRepositoriesEvent(event *github.InstallationRepositoriesEvent) ([]*entities.TargetEntity, *entities.EventDetails, error) {
	repositories := make([]*github.Repository, 0)

	if event.GetAction() == "added" {
		repositories = event.RepositoriesAdded
	}

	if event.GetAction() == "removed" {
		repositories = event.RepositoriesRemoved
	}

	targetEntities, err := extractTargetEntitiesFromRepositories(repositories, event.GetInstallation().GetAccount().GetType())
	if err != nil {
		return nil, nil, err
	}

	return targetEntities, &entities.EventDetails{
		EventName:   "installation_repositories",
		EventAction: event.GetAction(),
		Payload:     event,
	}, nil
}

func processCheckRunEvent(log *logrus.Entry, token string, event *github.CheckRunEvent) ([]*entities.TargetEntity, *entities.EventDetails, error) {
	log.Info("processing check_run event")

	targetEntities := []*entities.TargetEntity{}

	eventDetails := &entities.EventDetails{
		EventName:   "check_run",
		EventAction: event.GetAction(),
		Payload:     event,
	}

	//  if the head is from a forked repository the pull_requests array will be empty on check run events
	if len(event.CheckRun.PullRequests) == 0 {
		prs, err := getPullRequests(token, event.GetRepo().GetFullName())
		if err != nil {
			return nil, nil, err
		}

		log.Infof("fetched %d pull requests", len(prs))

		for _, pr := range prs {
			if pr.GetHead().GetSHA() == event.CheckRun.GetHeadSHA() {
				log.Infof("found pull request %v", pr.GetNumber())
				return []*entities.TargetEntity{
					{
						Kind:        entities.PullRequest,
						Number:      pr.GetNumber(),
						Owner:       event.GetRepo().GetOwner().GetLogin(),
						Repo:        event.GetRepo().GetName(),
						AccountType: pr.GetBase().GetRepo().GetOwner().GetType(),
						Visibility:  pr.GetBase().GetRepo().GetVisibility(),
					},
				}, eventDetails, nil
			}
		}

		log.Infof("no pr found with the head sha %v", event.CheckRun.GetHeadSHA())

		return []*entities.TargetEntity{}, eventDetails, nil
	}

	for _, pr := range event.CheckRun.PullRequests {
		targetEntities = append(targetEntities, &entities.TargetEntity{
			Kind:        entities.PullRequest,
			Owner:       event.GetRepo().GetOwner().GetLogin(),
			Repo:        event.GetRepo().GetName(),
			Number:      pr.GetNumber(),
			AccountType: event.GetRepo().GetOwner().GetType(),
			Visibility:  event.GetRepo().GetVisibility(),
		})
	}

	return targetEntities, eventDetails, nil
}

func processCheckSuiteEvent(log *logrus.Entry, token string, event *github.CheckSuiteEvent) ([]*entities.TargetEntity, *entities.EventDetails, error) {
	log.Info("processing check_suite event")

	targetEntities := []*entities.TargetEntity{}

	eventDetails := &entities.EventDetails{
		EventName:   "check_suite",
		EventAction: event.GetAction(),
		Payload:     event,
	}

	if event.Sender.GetLogin() == "github-merge-queue[bot]" && event.GetAction() == "requested" {
		return processCheckSuiteEventForMergeQueue(log, token, event, eventDetails)
	}

	// When the check suite is from a head of a forked repository the pull_requests array will be empty.
	// We need to fetch all the pull requests for the repository and find the one that matches the head sha.
	if len(event.CheckSuite.PullRequests) == 0 {
		log.Infof("no pull requests found in check suite event. fetching all pull requests for repository %v", event.GetRepo().GetFullName())

		prs, err := getPullRequests(token, event.GetRepo().GetFullName())
		if err != nil {
			return nil, nil, err
		}

		log.Infof("fetched %d pull requests", len(prs))

		for _, pr := range prs {
			if pr.GetHead().GetSHA() == event.CheckSuite.GetHeadSHA() {
				log.Infof("found pull request %v", pr.GetNumber())
				return []*entities.TargetEntity{
					{
						Kind:        entities.PullRequest,
						Number:      pr.GetNumber(),
						Owner:       event.GetRepo().GetOwner().GetLogin(),
						Repo:        event.GetRepo().GetName(),
						AccountType: pr.GetBase().GetRepo().GetOwner().GetType(),
						Visibility:  pr.GetBase().GetRepo().GetVisibility(),
					},
				}, eventDetails, nil
			}
		}

		log.Infof("no pull request found with the head sha %v", event.CheckSuite.GetHeadSHA())

		return []*entities.TargetEntity{}, eventDetails, nil
	} else {
		for _, pr := range event.CheckSuite.PullRequests {
			targetEntities = append(targetEntities, &entities.TargetEntity{
				Kind:        entities.PullRequest,
				Owner:       event.GetRepo().GetOwner().GetLogin(),
				Repo:        event.GetRepo().GetName(),
				Number:      pr.GetNumber(),
				AccountType: event.GetRepo().GetOwner().GetType(),
				Visibility:  event.GetRepo().GetVisibility(),
			})
		}
	}

	return targetEntities, eventDetails, nil
}

func processCheckSuiteEventForMergeQueue(log *logrus.Entry, token string, event *github.CheckSuiteEvent, eventDetails *entities.EventDetails) ([]*entities.TargetEntity, *entities.EventDetails, error) {
	// The GitHub Merge Queue creates a temporary branch where the head SHA of the check suite is from the temporary branch.
	// In order to find the pull request associated with the event, we can use the temporary branch name created by the merge queue.
	// The format of a GitHub Merge Queue temporary branch is: gh-readonly-queue/{target_branch}/pr-{pr_number}-{head_SHA_source_branch}
	// We can extract the pr number from the temporary branch name and find the pull request associated with it.
	prIdentifier := strings.Split(event.GetCheckSuite().GetHeadBranch(), "/")[2]
	prNumber, err := strconv.Atoi(strings.Split(prIdentifier, "-")[1])
	if err != nil {
		return nil, nil, fmt.Errorf("error converting pr number to int: %w", err)
	}

	pr, err := getPullRequest(token, event.GetRepo().GetFullName(), prNumber)
	if err != nil {
		return nil, nil, err
	}

	return []*entities.TargetEntity{
		{
			Kind:        entities.PullRequest,
			Number:      pr.GetNumber(),
			Owner:       event.GetRepo().GetOwner().GetLogin(),
			Repo:        event.GetRepo().GetName(),
			AccountType: pr.GetBase().GetRepo().GetOwner().GetType(),
			Visibility:  pr.GetBase().GetRepo().GetVisibility(),
		},
	}, eventDetails, nil
}

func getPullRequest(token, fullName string, prNumber int) (*github.PullRequest, error) {
	ctx, canc := context.WithTimeout(context.Background(), time.Minute*10)
	defer canc()

	ghClient := reviewpad_gh.NewGithubClientFromToken(ctx, token)

	repoParts := strings.SplitN(fullName, "/", 2)

	owner := repoParts[0]
	repo := repoParts[1]

	pr, _, err := ghClient.GetPullRequest(ctx, owner, repo, prNumber)
	if err != nil {
		return nil, fmt.Errorf("get pull request: %w", err)
	}

	return pr, nil
}

func getPullRequests(token, fullName string) ([]*github.PullRequest, error) {
	ctx, canc := context.WithTimeout(context.Background(), time.Minute*10)
	defer canc()

	ghClient := reviewpad_gh.NewGithubClientFromToken(ctx, token)

	repoParts := strings.SplitN(fullName, "/", 2)

	owner := repoParts[0]
	repo := repoParts[1]

	prs, err := ghClient.GetPullRequests(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("get pull requests: %w", err)
	}

	return prs, nil
}

func extractTargetEntitiesFromRepositories(repos []*github.Repository, accountType string) ([]*entities.TargetEntity, error) {
	targetEntities := make([]*entities.TargetEntity, 0)

	for _, repo := range repos {
		parts := strings.Split(repo.GetFullName(), "/")

		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid full repository name: %v", repo.GetFullName())
		}

		visibility := "public"
		if repo.GetPrivate() {
			visibility = "private"
		}

		targetEntities = append(targetEntities, &entities.TargetEntity{
			Owner:       parts[0],
			Repo:        parts[1],
			AccountType: accountType,
			Visibility:  visibility,
		})
	}

	return targetEntities, nil
}

// reviewpad-an: critical
// output: the list of pull requests/issues that are affected by the event.
func ProcessEvent(log *logrus.Entry, event *ActionEvent) ([]*entities.TargetEntity, *entities.EventDetails, error) {
	// These events do not have an equivalent in the GitHub webhooks, thus
	// parsing them with github.ParseWebhook would return an error.
	// These are the webhook events: https://docs.github.com/en/developers/webhooks-and-events/webhooks/webhook-events-and-payloads
	// And these are the "workflow events": https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows
	switch *event.EventName {
	case "schedule":
		return processCronEvent(log, *event.Token, event)
	}

	eventPayload, err := github.ParseWebHook(*event.EventName, *event.EventPayload)
	if err != nil {
		return nil, nil, fmt.Errorf("parse github webhook: %w", err)
	}

	switch payload := eventPayload.(type) {
	// Handle github events triggered by actions
	// For more information, visit: https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows
	case *github.BranchProtectionRuleEvent:
		return processUnsupportedEvent(payload)
	case *github.CheckRunEvent:
		return processCheckRunEvent(log, *event.Token, payload)
	case *github.CheckSuiteEvent:
		return processCheckSuiteEvent(log, *event.Token, payload)
	case *github.CommitCommentEvent:
		return processUnsupportedEvent(payload)
	case *github.ContentReferenceEvent:
		return processUnsupportedEvent(payload)
	case *github.CreateEvent:
		return processUnsupportedEvent(payload)
	case *github.DeleteEvent:
		return processUnsupportedEvent(payload)
	case *github.DeployKeyEvent:
		return processUnsupportedEvent(payload)
	case *github.DeploymentEvent:
		return processUnsupportedEvent(payload)
	case *github.DeploymentStatusEvent:
		return processUnsupportedEvent(payload)
	case *github.DiscussionEvent:
		return processDiscussionEvent(log, payload)
	case *github.DiscussionCommentEvent:
		return processDiscussionCommentEvent(log, payload)
	case *github.ForkEvent:
		return processUnsupportedEvent(payload)
	case *github.GitHubAppAuthorizationEvent:
		return processUnsupportedEvent(payload)
	case *github.GollumEvent:
		return processUnsupportedEvent(payload)
	case *github.InstallationEvent:
		return processInstallationEvent(log, payload)
	case *github.InstallationRepositoriesEvent:
		return processInstallationRepositoriesEvent(payload)
	case *github.IssueCommentEvent:
		return processIssueCommentEvent(log, payload)
	case *github.IssuesEvent:
		return processIssuesEvent(log, payload)
	case *github.LabelEvent:
		return processUnsupportedEvent(payload)
	case *github.MarketplacePurchaseEvent:
		return processUnsupportedEvent(payload)
	case *github.MemberEvent:
		return processUnsupportedEvent(payload)
	case *github.MembershipEvent:
		return processUnsupportedEvent(payload)
	case *github.MetaEvent:
		return processUnsupportedEvent(payload)
	case *github.MilestoneEvent:
		return processUnsupportedEvent(payload)
	case *github.OrganizationEvent:
		return processUnsupportedEvent(payload)
	case *github.OrgBlockEvent:
		return processUnsupportedEvent(payload)
	case *github.PackageEvent:
		return processUnsupportedEvent(payload)
	case *github.PageBuildEvent:
		return processUnsupportedEvent(payload)
	case *github.PingEvent:
		return processUnsupportedEvent(payload)
	case *github.ProjectEvent:
		return processUnsupportedEvent(payload)
	case *github.ProjectCardEvent:
		return processUnsupportedEvent(payload)
	case *github.ProjectColumnEvent:
		return processUnsupportedEvent(payload)
	case *github.PublicEvent:
		return processUnsupportedEvent(payload)
	case *github.PullRequestEvent:
		return processPullRequestEvent(log, payload)
	case *github.PullRequestReviewEvent:
		return processPullRequestReviewEvent(log, payload)
	case *github.PullRequestReviewCommentEvent:
		return processPullRequestReviewCommentEvent(log, payload)
	case *github.PullRequestReviewThreadEvent:
		return processUnsupportedEvent(payload)
	case *github.PullRequestTargetEvent:
		return processPullRequestTargetEvent(log, payload)
	case *github.PushEvent:
		return processPushEvent(log, *event.Token, payload)
	case *github.ReleaseEvent:
		return processUnsupportedEvent(payload)
	case *github.RepositoryEvent:
		return processUnsupportedEvent(payload)
	case *github.RepositoryDispatchEvent:
		return processUnsupportedEvent(payload)
	case *github.RepositoryImportEvent:
		return processUnsupportedEvent(payload)
	case *github.RepositoryVulnerabilityAlertEvent:
		return processUnsupportedEvent(payload)
	case *github.SecretScanningAlertEvent:
		return processUnsupportedEvent(payload)
	case *github.StarEvent:
		return processUnsupportedEvent(payload)
	case *github.StatusEvent:
		return processStatusEvent(log, *event.Token, payload)
	case *github.TeamEvent:
		return processUnsupportedEvent(payload)
	case *github.TeamAddEvent:
		return processUnsupportedEvent(payload)
	case *github.UserEvent:
		return processUnsupportedEvent(payload)
	case *github.WatchEvent:
		return processUnsupportedEvent(payload)
	case *github.WorkflowDispatchEvent:
		return processUnsupportedEvent(payload)
	case *github.WorkflowJobEvent:
		return processUnsupportedEvent(payload)
	case *github.WorkflowRunEvent:
		return processWorkflowRunEvent(log, *event.Token, payload)
	}

	return processUnsupportedEvent(eventPayload)
}

func GetEventSender(eventPayload interface{}) string {
	switch payload := eventPayload.(type) {
	case *github.BranchProtectionRuleEvent:
		return payload.Sender.GetLogin()
	case *github.CheckRunEvent:
		return payload.Sender.GetLogin()
	case *github.CheckSuiteEvent:
		return payload.Sender.GetLogin()
	case *github.CommitCommentEvent:
		return payload.Sender.GetLogin()
	case *github.ContentReferenceEvent:
		return payload.Sender.GetLogin()
	case *github.CreateEvent:
		return payload.Sender.GetLogin()
	case *github.DeleteEvent:
		return payload.Sender.GetLogin()
	case *github.DeployKeyEvent:
		return payload.Sender.GetLogin()
	case *github.DeploymentEvent:
		return payload.Sender.GetLogin()
	case *github.DeploymentStatusEvent:
		return payload.Sender.GetLogin()
	case *github.DiscussionEvent:
		return payload.Sender.GetLogin()
	case *github.ForkEvent:
		return payload.Sender.GetLogin()
	case *github.GitHubAppAuthorizationEvent:
		return payload.Sender.GetLogin()
	case *github.GollumEvent:
		return payload.Sender.GetLogin()
	case *github.InstallationEvent:
		return payload.Sender.GetLogin()
	case *github.InstallationRepositoriesEvent:
		return payload.Sender.GetLogin()
	case *github.IssueCommentEvent:
		return payload.Sender.GetLogin()
	case *github.IssuesEvent:
		return payload.Sender.GetLogin()
	case *github.LabelEvent:
		return payload.Sender.GetLogin()
	case *github.MarketplacePurchaseEvent:
		return payload.Sender.GetLogin()
	case *github.MemberEvent:
		return payload.Sender.GetLogin()
	case *github.MembershipEvent:
		return payload.Sender.GetLogin()
	case *github.MetaEvent:
		return payload.Sender.GetLogin()
	case *github.MilestoneEvent:
		return payload.Sender.GetLogin()
	case *github.OrganizationEvent:
		return payload.Sender.GetLogin()
	case *github.OrgBlockEvent:
		return payload.Sender.GetLogin()
	case *github.PackageEvent:
		return payload.Sender.GetLogin()
	case *github.PageBuildEvent:
		return payload.Sender.GetLogin()
	case *github.PingEvent:
		return payload.Sender.GetLogin()
	case *github.ProjectEvent:
		return payload.Sender.GetLogin()
	case *github.ProjectCardEvent:
		return payload.Sender.GetLogin()
	case *github.ProjectColumnEvent:
		return payload.Sender.GetLogin()
	case *github.PublicEvent:
		return payload.Sender.GetLogin()
	case *github.PullRequestEvent:
		return payload.Sender.GetLogin()
	case *github.PullRequestReviewEvent:
		return payload.Sender.GetLogin()
	case *github.PullRequestReviewCommentEvent:
		return payload.Sender.GetLogin()
	case *github.PullRequestReviewThreadEvent:
		return payload.Sender.GetLogin()
	case *github.PullRequestTargetEvent:
		return payload.Sender.GetLogin()
	case *github.PushEvent:
		return payload.Sender.GetLogin()
	case *github.ReleaseEvent:
		return payload.Sender.GetLogin()
	case *github.RepositoryEvent:
		return payload.Sender.GetLogin()
	case *github.RepositoryDispatchEvent:
		return payload.Sender.GetLogin()
	case *github.RepositoryImportEvent:
		return payload.Sender.GetLogin()
	case *github.RepositoryVulnerabilityAlertEvent:
		return payload.Sender.GetLogin()
	case *github.SecretScanningAlertEvent:
		return payload.Sender.GetLogin()
	case *github.StarEvent:
		return payload.Sender.GetLogin()
	case *github.StatusEvent:
		return payload.Sender.GetLogin()
	case *github.TeamEvent:
		return payload.Sender.GetLogin()
	case *github.TeamAddEvent:
		return payload.Sender.GetLogin()
	case *github.UserEvent:
		return payload.Sender.GetLogin()
	case *github.WatchEvent:
		return payload.Sender.GetLogin()
	case *github.WorkflowDispatchEvent:
		return payload.Sender.GetLogin()
	case *github.WorkflowJobEvent:
		return payload.Sender.GetLogin()
	case *github.WorkflowRunEvent:
		return payload.Sender.GetLogin()
	}

	return ""
}
