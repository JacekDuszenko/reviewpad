// Copyright 2022 Explore.dev Unipessoal Lda. All Rights Reserved.
// Use of this source code is governed by a license that can be
// found in the LICENSE file.

package plugins_aladino

import (
	"log"

	"github.com/reviewpad/reviewpad/v4/lang/aladino"
	actions "github.com/reviewpad/reviewpad/v4/plugins/aladino/actions"
	functions "github.com/reviewpad/reviewpad/v4/plugins/aladino/functions"
	services "github.com/reviewpad/reviewpad/v4/plugins/aladino/services"
	"google.golang.org/grpc"
)

type PluginConfig struct {
	Services        map[string]interface{}
	grpcConnections []*grpc.ClientConn
}

func DefaultPluginConfig() (*PluginConfig, error) {
	connections := make([]*grpc.ClientConn, 0)
	semanticClient, semanticConnection, err := services.NewSemanticService()
	if err != nil {
		return nil, err
	}

	robinClient, robinConnection, err := services.NewRobinService()
	if err != nil {
		return nil, err
	}

	services := map[string]interface{}{
		services.SEMANTIC_SERVICE_KEY: semanticClient,
		services.ROBIN_SERVICE_KEY:    robinClient,
	}

	connections = append(connections, semanticConnection, robinConnection)

	config := &PluginConfig{
		Services:        services,
		grpcConnections: connections,
	}

	return config, nil
}

func (config *PluginConfig) CleanupPluginConfig() {
	for _, connection := range config.grpcConnections {
		connection.Close()
	}
}

// The documentation for the builtins is in:
// https://github.com/reviewpad/docs/blob/main/aladino/builtins.md
// This means that changes to the builtins need to be propagated to that document.
func PluginBuiltInsWithConfig(config *PluginConfig) *aladino.BuiltIns {
	return &aladino.BuiltIns{
		Functions: map[string]*aladino.BuiltInFunction{
			// Pull Request
			"approvalsCount":                functions.ApprovalsCount(),
			"assignees":                     functions.Assignees(),
			"author":                        functions.Author(),
			"base":                          functions.Base(),
			"changed":                       functions.Changed(),
			"checkRunConclusion":            functions.CheckRunConclusion(),
			"commentCount":                  functions.CommentCount(),
			"comments":                      functions.Comments(),
			"commitCount":                   functions.CommitCount(),
			"commits":                       functions.Commits(),
			"context":                       functions.Context(),
			"createdAt":                     functions.CreatedAt(),
			"description":                   functions.Description(),
			"eventType":                     functions.EventType(),
			"fileCount":                     functions.FileCount(),
			"filesPath":                     functions.FilesPath(),
			"hasAnnotation":                 functions.HasAnnotation(),
			"hasAnyCheckRunCompleted":       functions.HasAnyCheckRunCompleted(),
			"hasBinaryFile":                 functions.HasBinaryFile(),
			"hasCodePattern":                functions.HasCodePattern(),
			"hasCodeWithoutSemanticChanges": functions.HasCodeWithoutSemanticChanges(),
			"hasFileExtensions":             functions.HasFileExtensions(),
			"hasFileName":                   functions.HasFileName(),
			"hasFilePattern":                functions.HasFilePattern(),
			"hasGitConflicts":               functions.HasGitConflicts(),
			"hasLinearHistory":              functions.HasLinearHistory(),
			"hasLinkedIssues":               functions.HasLinkedIssues(),
			"hasRequiredApprovals":          functions.HasRequiredApprovals(),
			"hasUnaddressedThreads":         functions.HasUnaddressedThreads(),
			"haveAllChecksRunCompleted":     functions.HaveAllChecksRunCompleted(),
			"head":                          functions.Head(),
			"isBinary":                      functions.IsBinary(),
			"isDraft":                       functions.IsDraft(),
			"isLinkedToProject":             functions.IsLinkedToProject(),
			"isMerged":                      functions.IsMerged(),
			"isUpdatedWithBaseBranch":       functions.IsUpdatedWithBaseBranch(),
			"isWaitingForReview":            functions.IsWaitingForReview(),
			"labels":                        functions.Labels(),
			"lastEventAt":                   functions.LastEventAt(),
			"milestone":                     functions.Milestone(),
			"requestedReviewers":            functions.RequestedReviewers(),
			"reviewers":                     functions.Reviewers(),
			"reviewerStatus":                functions.ReviewerStatus(),
			"size":                          functions.Size(),
			"state":                         functions.State(),
			"title":                         functions.Title(),
			"toJSON":                        functions.ToJSON(),
			"workflowStatus":                functions.WorkflowStatus(),
			// Organization
			"organization": functions.Organization(),
			"team":         functions.Team(),
			// User
			"issueCountBy":             functions.IssueCountBy(),
			"pullRequestCountBy":       functions.PullRequestCountBy(),
			"totalCodeReviews":         functions.TotalCodeReviews(),
			"totalCreatedPullRequests": functions.TotalCreatedPullRequests(),
			// Utilities
			"all":                           functions.All(),
			"any":                           functions.Any(),
			"append":                        functions.AppendString(),
			"contains":                      functions.Contains(),
			"extractMarkdownHeadingContent": functions.ExtractMarkdownHeadingContent(),
			"isElementOf":                   functions.IsElementOf(),
			"join":                          functions.Join(),
			"length":                        functions.Length(),
			"matchString":                   functions.MatchString(),
			"selectFromContext":             functions.SelectFromContext(),
			"selectFromJSON":                functions.SelectFromJSON(),
			"sprintf":                       functions.Sprintf(),
			"startsWith":                    functions.StartsWith(),
			"subMatchesString":              functions.SubMatchesString(),
			"toBool":                        functions.ToBool(),
			"toNumber":                      functions.ToNumber(),
			"toStringArray":                 functions.ToStringArray(),
			// Engine
			"dictionary": functions.Dictionary(),
			"group":      functions.Group(),
			"rule":       functions.Rule(),
			// Internal
			"filter": functions.Filter(),
		},
		Actions: map[string]*aladino.BuiltInAction{
			"addLabel":                  actions.AddLabel(),
			"addToProject":              actions.AddToProject(),
			"approve":                   actions.Approve(),
			"assignAssignees":           actions.AssignAssignees(),
			"assignCodeAuthorReviewers": actions.AssignCodeAuthorReviewers(),
			"assignRandomReviewer":      actions.AssignRandomReviewer(),
			"assignReviewer":            actions.AssignReviewer(),
			"assignTeamReviewer":        actions.AssignTeamReviewer(),
			"close":                     actions.Close(),
			"comment":                   actions.Comment(),
			"commentOnce":               actions.CommentOnce(),
			"commitLint":                actions.CommitLint(),
			"deleteHeadBranch":          actions.DeleteHeadBranch(),
			"disableActions":            actions.DisableActions(),
			"error":                     actions.ErrorMsg(),
			"fail":                      actions.Fail(),
			"failCheckStatus":           actions.FailCheckStatus(),
			"info":                      actions.Info(),
			"merge":                     actions.Merge(),
			"rebase":                    actions.Rebase(),
			"removeFromProject":         actions.RemoveFromProject(),
			"removeLabel":               actions.RemoveLabel(),
			"removeLabels":              actions.RemoveLabels(),
			"review":                    actions.Review(),
			"robinPrompt":               actions.RobinPrompt(),
			"robinRawPrompt":            actions.RobinRawPrompt(),
			"robinSummarize":            actions.RobinSummarize(),
			"setProjectField":           actions.SetProjectField(),
			"titleLint":                 actions.TitleLint(),
			"triggerWorkflow":           actions.TriggerWorkflow(),
			"warn":                      actions.Warn(),
		},
		Services: config.Services,
	}
}

func PluginBuiltIns() *aladino.BuiltIns {
	config, err := DefaultPluginConfig()
	if err != nil {
		log.Fatal("Error loading default plugin config: ", err)
	}
	return PluginBuiltInsWithConfig(config)
}
