// Copyright 2022 Explore.dev Unipessoal Lda. All Rights Reserved.
// Use of this source code is governed by a license that can be
// found in the LICENSE file.

package engine_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-github/v52/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/reviewpad/go-lib/entities"
	"github.com/reviewpad/reviewpad/v4/codehost"
	gh "github.com/reviewpad/reviewpad/v4/codehost/github"
	"github.com/reviewpad/reviewpad/v4/engine"
	"github.com/reviewpad/reviewpad/v4/lang"
	"github.com/reviewpad/reviewpad/v4/lang/aladino"
	plugins_aladino_actions "github.com/reviewpad/reviewpad/v4/plugins/aladino/actions"
	plugins_aladino_functions "github.com/reviewpad/reviewpad/v4/plugins/aladino/functions"
	"github.com/reviewpad/reviewpad/v4/utils"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestEval_WhenGitHubRequestsFail(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())
	tests := map[string]struct {
		inputReviewpadFilePath string
		inputContext           context.Context
		inputGitHubClient      *gh.GithubClient
		clientOptions          []mock.MockBackendOption
		wantErr                string
	}{
		"when get label request fails": {
			inputReviewpadFilePath: "testdata/exec/reviewpad_with_label_without_description.yml",
			clientOptions: []mock.MockBackendOption{
				mock.WithRequestMatchHandler(
					mock.GetReposLabelsByOwnerByRepoByName,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusInternalServerError)
						engine.MustWriteBytes(w, mock.MustMarshal(github.ErrorResponse{
							// An error response may also consist of a 404 status code.
							// However, in this context, such response means a label does not exist.
							Response: &http.Response{
								StatusCode: 500,
							},
							Message: "GetLabelRequestFailed",
						}))
					}),
				),
			},
			wantErr: "GetLabelRequestFailed",
		},
		"when create label request fails": {
			inputReviewpadFilePath: "testdata/exec/reviewpad_with_label_without_description.yml",
			clientOptions: []mock.MockBackendOption{
				mock.WithRequestMatchHandler(
					mock.GetReposLabelsByOwnerByRepoByName,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusNotFound)
						engine.MustWriteBytes(w, mock.MustMarshal(github.ErrorResponse{
							Response: &http.Response{
								StatusCode: 404,
							},
							Message: "Resource not found",
						}))
					}),
				),
				mock.WithRequestMatchHandler(
					mock.PostReposLabelsByOwnerByRepo,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						mock.WriteError(
							w,
							http.StatusInternalServerError,
							"CreateLabelRequestFailed",
						)
					}),
				),
			},
			wantErr: "CreateLabelRequestFailed",
		},
	}

	codehostClient := aladino.GetDefaultCodeHostClient(t, aladino.GetDefaultPullRequestDetails(), aladino.GetDefaultPullRequestFileList(), nil, nil)

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockedClient := engine.MockGithubClient(test.clientOptions)

			mockedAladinoInterpreter, err := mockAladinoInterpreter(mockedClient, codehostClient)
			if err != nil {
				assert.FailNow(t, fmt.Sprintf("mockAladinoInterpreter: %v", err))
			}

			mockedEnv, err := engine.MockEnvWith(mockedClient, mockedAladinoInterpreter, engine.DefaultMockTargetEntity, engine.DefaultMockEventDetails)
			if err != nil {
				assert.FailNow(t, fmt.Sprintf("engine MockEnvWith: %v", err))
			}

			reviewpadFileData, err := utils.ReadFile(test.inputReviewpadFilePath)
			if err != nil {
				assert.FailNow(t, fmt.Sprintf("Error reading reviewpad file: %v", err))
			}

			reviewpadFile, err := engine.Load(test.inputContext, logger, test.inputGitHubClient, reviewpadFileData)
			if err != nil {
				assert.FailNow(t, "Error parsing reviewpad file: %v", err)
			}

			gotProgram, err := engine.EvalConfigurationFile(reviewpadFile, mockedEnv)

			assert.Nil(t, gotProgram)
			assert.Equal(t, test.wantErr, err.(*github.ErrorResponse).Message)
		})
	}
}

func TestEvalCommand(t *testing.T) {
	tests := map[string]struct {
		inputContext      context.Context
		inputGitHubClient *gh.GithubClient
		clientOptions     []mock.MockBackendOption
		targetEntity      *entities.TargetEntity
		eventDetails      *entities.EventDetails
		wantProgram       *engine.Program
		wantErr           string
	}{
		"when invalid assign reviewer command": {
			targetEntity: &entities.TargetEntity{
				Kind:   engine.DefaultMockTargetEntity.Kind,
				Number: engine.DefaultMockTargetEntity.Number,
				Owner:  engine.DefaultMockTargetEntity.Owner,
				Repo:   engine.DefaultMockTargetEntity.Repo,
			},
			eventDetails: &entities.EventDetails{
				EventName:   "issue_comment",
				EventAction: "created",
				Payload: &github.IssueCommentEvent{
					Comment: &github.IssueComment{
						ID:   github.Int64(1),
						Body: github.String("/reviewpad assign-reviewer john, jane, 1 random"),
					},
				},
			},
			clientOptions: []mock.MockBackendOption{
				mock.WithRequestMatch(
					mock.PostReposIssuesCommentsByOwnerByRepoByIssueNumber,
					&github.IssueComment{},
				),
			},
			wantErr: "accepts 1 arg, received 4",
		},
		"when invalid total required reviewers arg": {
			targetEntity: &entities.TargetEntity{
				Kind:   engine.DefaultMockTargetEntity.Kind,
				Number: engine.DefaultMockTargetEntity.Number,
				Owner:  engine.DefaultMockTargetEntity.Owner,
				Repo:   engine.DefaultMockTargetEntity.Repo,
			},
			eventDetails: &entities.EventDetails{
				EventName:   "issue_comment",
				EventAction: "created",
				Payload: &github.IssueCommentEvent{
					Comment: &github.IssueComment{
						ID:   github.Int64(1),
						Body: github.String("/reviewpad assign-reviewer reviewpad"),
					},
				},
			},
			clientOptions: []mock.MockBackendOption{
				mock.WithRequestMatch(
					mock.PostReposIssuesCommentsByOwnerByRepoByIssueNumber,
					&github.IssueComment{},
				),
			},
			wantErr: "invalid argument: reviewpad, number of reviewers must be a number",
		},
		"when total required reviewers arg is 0": {
			targetEntity: &entities.TargetEntity{
				Kind:   engine.DefaultMockTargetEntity.Kind,
				Number: engine.DefaultMockTargetEntity.Number,
				Owner:  engine.DefaultMockTargetEntity.Owner,
				Repo:   engine.DefaultMockTargetEntity.Repo,
			},
			eventDetails: &entities.EventDetails{
				EventName:   "issue_comment",
				EventAction: "created",
				Payload: &github.IssueCommentEvent{
					Comment: &github.IssueComment{
						ID:   github.Int64(1),
						Body: github.String("/reviewpad assign-reviewer 0"),
					},
				},
			},
			clientOptions: []mock.MockBackendOption{
				mock.WithRequestMatch(
					mock.PostReposIssuesCommentsByOwnerByRepoByIssueNumber,
					&github.IssueComment{},
				),
			},
			wantErr: "invalid argument: 0, number of reviewers must be greater than 0",
		},
		"when assign reviewer has 1 total required reviewer": {
			wantProgram: engine.BuildProgram(
				[]*engine.Statement{
					engine.BuildStatement(`$assignCodeAuthorReviewers(1, [], 0)`),
				},
			),
			targetEntity: &entities.TargetEntity{
				Kind:   engine.DefaultMockTargetEntity.Kind,
				Number: engine.DefaultMockTargetEntity.Number,
				Owner:  engine.DefaultMockTargetEntity.Owner,
				Repo:   engine.DefaultMockTargetEntity.Repo,
			},
			eventDetails: &entities.EventDetails{
				EventName:   "issue_comment",
				EventAction: "created",
				Payload: &github.IssueCommentEvent{
					Comment: &github.IssueComment{
						ID:   github.Int64(1),
						Body: github.String("/reviewpad assign-reviewer 1"),
					},
				},
			},
		},
		"when assign reviewer has no total required reviewers provided": {
			wantProgram: engine.BuildProgram(
				[]*engine.Statement{
					engine.BuildStatement(`$assignCodeAuthorReviewers(1, [], 0)`),
				},
			),
			targetEntity: &entities.TargetEntity{
				Kind:   engine.DefaultMockTargetEntity.Kind,
				Number: engine.DefaultMockTargetEntity.Number,
				Owner:  engine.DefaultMockTargetEntity.Owner,
				Repo:   engine.DefaultMockTargetEntity.Repo,
			},
			eventDetails: &entities.EventDetails{
				EventName:   "issue_comment",
				EventAction: "created",
				Payload: &github.IssueCommentEvent{
					Comment: &github.IssueComment{
						ID:   github.Int64(1),
						Body: github.String("/reviewpad assign-reviewer"),
					},
				},
			},
		},
	}

	codehostClient := aladino.GetDefaultCodeHostClient(t, aladino.GetDefaultPullRequestDetails(), aladino.GetDefaultPullRequestFileList(), nil, nil)

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockedClient := engine.MockGithubClient(test.clientOptions)

			mockedAladinoInterpreter, err := mockAladinoInterpreter(mockedClient, codehostClient)
			if err != nil {
				assert.FailNow(t, fmt.Sprintf("mockAladinoInterpreter: %v", err))
			}

			mockedEnv, err := engine.MockEnvWith(mockedClient, mockedAladinoInterpreter, test.targetEntity, test.eventDetails)
			if err != nil {
				assert.FailNow(t, fmt.Sprintf("engine MockEnvWith: %v", err))
			}

			gotProgram, gotErr := engine.EvalCommand(test.eventDetails.Payload.(*github.IssueCommentEvent).GetComment().GetBody(), mockedEnv)

			if gotErr != nil && gotErr.Error() != test.wantErr {
				assert.FailNow(t, fmt.Sprintf("Eval() error = %v, wantErr %v", gotErr, test.wantErr))
			}
			assert.Equal(t, test.wantProgram, gotProgram)
		})
	}
}

func TestExecConfigurationFile(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())
	tests := map[string]struct {
		inputReviewpadFilePath string
		inputContext           context.Context
		inputGitHubClient      *gh.GithubClient
		clientOptions          []mock.MockBackendOption
		targetEntity           *entities.TargetEntity
		eventDetails           *entities.EventDetails
		wantProgram            *engine.Program
		wantErr                string
		wantExitStatus         engine.ExitStatus
	}{
		"reviewpad with workflow extra actions then": {
			inputReviewpadFilePath: "testdata/exec/reviewpad_with_workflow_extra_actions_then.yml",
			wantProgram: engine.BuildProgram(
				[]*engine.Statement{
					engine.BuildStatement(`$addLabel("then-clause")`),
					engine.BuildStatement(`$addLabel("extra-actions")`),
				},
			),
			targetEntity:   engine.DefaultMockTargetEntity,
			wantExitStatus: engine.ExitStatusSuccess,
		},
		"reviewpad with workflow else": {
			inputReviewpadFilePath: "testdata/exec/reviewpad_with_workflow_else.yml",
			wantProgram: engine.BuildProgram(
				[]*engine.Statement{
					engine.BuildStatement(`$addLabel("else-clause")`),
					engine.BuildStatement(`$addLabel("after-if")`),
				},
			),
			targetEntity:   engine.DefaultMockTargetEntity,
			wantExitStatus: engine.ExitStatusSuccess,
		},
		"reviewpad with for each workflow": {
			inputReviewpadFilePath: "testdata/exec/reviewpad_with_foreach_workflow.yml",
			wantProgram: engine.BuildProgram(
				[]*engine.Statement{
					engine.BuildStatement(`$addLabel($teamName)`),
					engine.BuildStatement(`$addLabel($teamName)`),
				},
			),
			targetEntity:   engine.DefaultMockTargetEntity,
			wantExitStatus: engine.ExitStatusSuccess,
		},
		"reviewpad with nested for each workflow": {
			inputReviewpadFilePath: "testdata/exec/reviewpad_with_nested_foreach_workflow.yml",
			wantProgram: engine.BuildProgram(
				[]*engine.Statement{
					engine.BuildStatement(`$addLabel($teamName)`),
					engine.BuildStatement(`$assignReviewer($team($teamName2), 1, "random")`),
					engine.BuildStatement(`$assignReviewer($team($teamName2), 1, "random")`),
					engine.BuildStatement(`$addLabel($teamName)`),
					engine.BuildStatement(`$assignReviewer($team($teamName2), 1, "random")`),
					engine.BuildStatement(`$assignReviewer($team($teamName2), 1, "random")`),
				},
			),
			targetEntity:   engine.DefaultMockTargetEntity,
			wantExitStatus: engine.ExitStatusSuccess,
		},
		"reviewpad with conditional for each workflow": {
			inputReviewpadFilePath: "testdata/exec/reviewpad_with_conditional_foreach_workflow.yml",
			wantProgram: engine.BuildProgram(
				[]*engine.Statement{
					engine.BuildStatement(`$addLabel($teamName)`),
					engine.BuildStatement(`$addLabel($teamName)`),
				},
			),
			targetEntity:   engine.DefaultMockTargetEntity,
			wantExitStatus: engine.ExitStatusSuccess,
		},
		"reviewpad with pipelines": {
			inputReviewpadFilePath: "testdata/exec/reviewpad_with_pipelines.yml",
			wantProgram: engine.BuildProgram(
				[]*engine.Statement{
					engine.BuildStatement(`$addLabel("pipeline with no trigger")`),
					engine.BuildStatement(`$addLabel("pipeline with trigger")`),
					engine.BuildStatement(`$addLabel("pipeline with single stage and single action")`),
					engine.BuildStatement(`$addLabel("pipeline with single stage and multiple actions - 1/2")`),
					engine.BuildStatement(`$addLabel("pipeline with single stage and multiple actions - 2/2")`),
					engine.BuildStatement(`$addLabel("pipeline with multiple stages - stage 2")`),
					engine.BuildStatement(`$addLabel("pipeline with multiple stages and multiple actions - stage 2 - 1/2")`),
					engine.BuildStatement(`$addLabel("pipeline with multiple stages and multiple actions - stage 2 - 2/2")`),
					engine.BuildStatement(`$addLabel("pipeline with single stage and concise actions")`),
					engine.BuildStatement(`$addLabel("pipeline with multiple stages and concise actions - stage 2")`),
					engine.BuildStatement(`$addLabel("pipeline with more than 2 stages - stage 2")`),
				},
			),
			targetEntity:   engine.DefaultMockTargetEntity,
			wantExitStatus: engine.ExitStatusSuccess,
		},
		"reviewpad with dictionaries and for each": {
			inputReviewpadFilePath: "testdata/exec/reviewpad_with_dictionary_and_foreach_workflow.yml",
			wantProgram: engine.BuildProgram(
				[]*engine.Statement{
					engine.BuildStatement(`$addLabel($teamName)`),
					engine.BuildStatement(`$addLabel($teamMember)`),
					engine.BuildStatement(`$addLabel($teamMember)`),
					engine.BuildStatement(`$addLabel($teamName)`),
					engine.BuildStatement(`$addLabel($teamMember)`),
					engine.BuildStatement(`$addLabel($teamMember)`),
				},
			),
			targetEntity:   engine.DefaultMockTargetEntity,
			wantExitStatus: engine.ExitStatusSuccess,
		},
		"reviewpad with referenced dictionaries and for each": {
			inputReviewpadFilePath: "testdata/exec/reviewpad_with_referenced_dictionary_and_foreach_workflow.yml",
			wantProgram: engine.BuildProgram(
				[]*engine.Statement{
					engine.BuildStatement(`$addLabel($groupMember)`),
					engine.BuildStatement(`$addLabel($groupMember)`),
					engine.BuildStatement(`$addLabel($groupMember)`),
					engine.BuildStatement(`$addLabel($groupMember)`),
					engine.BuildStatement(`$addLabel($groupMember)`),
					engine.BuildStatement(`$addLabel($groupMember)`),
					engine.BuildStatement(`$addLabel($groupMember)`),
					engine.BuildStatement(`$addLabel($groupMember)`),
				},
			),
			targetEntity:   engine.DefaultMockTargetEntity,
			wantExitStatus: engine.ExitStatusSuccess,
		},
	}

	codehostClient := aladino.GetDefaultCodeHostClient(t, aladino.GetDefaultPullRequestDetails(), aladino.GetDefaultPullRequestFileList(), nil, nil)

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockedClient := engine.MockGithubClient(test.clientOptions)

			builtInName := "addLabel"
			builtIns := &aladino.BuiltIns{
				Actions: map[string]*aladino.BuiltInAction{
					builtInName: {
						Type: lang.BuildFunctionType([]lang.Type{lang.BuildStringType()}, lang.BuildArrayOfType(lang.BuildStringType())),
						Code: func(e aladino.Env, args []lang.Value) error {
							return nil
						},
						SupportedKinds: []entities.TargetEntityKind{entities.PullRequest},
					},
					"assignReviewer": plugins_aladino_actions.AssignReviewer(),
				},
				Functions: map[string]*aladino.BuiltInFunction{
					"group":      plugins_aladino_functions.Group(),
					"team":       plugins_aladino_functions.Team(),
					"dictionary": plugins_aladino_functions.Dictionary(),
				},
			}

			dryRun := true
			mockedAladinoInterpreter, err := aladino.NewInterpreter(
				engine.DefaultMockCtx,
				logger,
				dryRun,
				mockedClient,
				codehostClient,
				engine.DefaultMockCollector,
				engine.DefaultMockTargetEntity,
				engine.DefaultMockEventPayload,
				builtIns,
				nil,
			)
			if err != nil {
				assert.FailNow(t, fmt.Sprintf("mockAladinoInterpreter: %v", err))
			}

			mockedEnv, err := engine.MockEnvWith(mockedClient, mockedAladinoInterpreter, test.targetEntity, test.eventDetails)
			if err != nil {
				assert.FailNow(t, fmt.Sprintf("engine MockEnvWith: %v", err))
			}
			mockedEnv.DryRun = dryRun

			reviewpadFileData, err := utils.ReadFile(test.inputReviewpadFilePath)
			if err != nil {
				assert.FailNow(t, fmt.Sprintf("Error reading reviewpad file: %v", err))
			}

			reviewpadFile, err := engine.Load(test.inputContext, logger, test.inputGitHubClient, reviewpadFileData)
			if err != nil {
				assert.FailNow(t, fmt.Sprintf("Error parsing reviewpad file: %v", err))
			}

			gotExitStatus, gotProgram, gotErr := engine.ExecConfigurationFile(mockedEnv, reviewpadFile)

			if gotErr != nil && gotErr.Error() != test.wantErr {
				assert.FailNow(t, fmt.Sprintf("Eval() error = %v, wantErr %v", gotErr, test.wantErr))
			}
			assert.Equal(t, test.wantExitStatus, gotExitStatus)
			assert.Equal(t, test.wantProgram, gotProgram)
		})
	}
}

func TestEvalConfigurationFile(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())
	tests := map[string]struct {
		inputReviewpadFilePath string
		inputContext           context.Context
		inputGitHubClient      *gh.GithubClient
		clientOptions          []mock.MockBackendOption
		targetEntity           *entities.TargetEntity
		eventDetails           *entities.EventDetails
		wantProgram            *engine.Program
		wantErr                string
	}{
		"when label has no name": {
			inputReviewpadFilePath: "testdata/exec/reviewpad_with_unnamed_label.yml",
			clientOptions: []mock.MockBackendOption{
				mockGetReposLabelsByOwnerByRepoByName("bug", ""),
				mockPatchReposLabelsByOwnerByRepo("bug", ""),
			},
			wantProgram: engine.BuildProgram(
				[]*engine.Statement{
					engine.BuildStatement(`$addLabel("test-unnamed-label")`),
				},
			),
			targetEntity: engine.DefaultMockTargetEntity,
		},
		"when label exists": {
			inputReviewpadFilePath: "testdata/exec/reviewpad_with_label_without_description.yml",
			clientOptions:          []mock.MockBackendOption{mockGetReposLabelsByOwnerByRepoByName("bug", "")},
			wantProgram: engine.BuildProgram(
				[]*engine.Statement{
					engine.BuildStatement(`$addLabel("test-label-without-description")`),
				},
			),
			targetEntity: engine.DefaultMockTargetEntity,
		},
		"when label does not exist": {
			inputReviewpadFilePath: "testdata/exec/reviewpad_with_label_without_description.yml",
			clientOptions: []mock.MockBackendOption{
				mockGetReposLabelsByOwnerByRepoByName("bug", ""),
				mockPostReposLabelsByOwnerByRepo("bug", ""),
			},
			wantProgram: engine.BuildProgram(
				[]*engine.Statement{
					engine.BuildStatement(`$addLabel("test-label-without-description")`),
				},
			),
			targetEntity: engine.DefaultMockTargetEntity,
		},
		"when label has updates": {
			inputReviewpadFilePath: "testdata/exec/reviewpad_with_label_with_diff_repo_description.yml",
			clientOptions: []mock.MockBackendOption{
				mockGetReposLabelsByOwnerByRepoByName("bug", "Something is wrong"),
				mockPatchReposLabelsByOwnerByRepo("bug", "Something is definitely wrong"),
			},
			wantProgram: engine.BuildProgram(
				[]*engine.Statement{
					engine.BuildStatement(`$addLabel("test-label-update")`),
				},
			),
			targetEntity: engine.DefaultMockTargetEntity,
		},
		"when label has no updates": {
			inputReviewpadFilePath: "testdata/exec/reviewpad_with_label_with_repo_description.yml",
			clientOptions:          []mock.MockBackendOption{mockGetReposLabelsByOwnerByRepoByName("bug", "Something is wrong")},
			wantProgram: engine.BuildProgram(
				[]*engine.Statement{
					engine.BuildStatement(`$addLabel("test-label-with-no-updates")`),
				},
			),
			targetEntity: engine.DefaultMockTargetEntity,
		},
		"when group is valid": {
			inputReviewpadFilePath: "testdata/exec/reviewpad_with_valid_group.yml",
			wantProgram: engine.BuildProgram(
				[]*engine.Statement{
					engine.BuildStatement(`$addLabel("test-valid-group")`),
				},
			),
			targetEntity: engine.DefaultMockTargetEntity,
		},
		"when group is invalid": {
			inputReviewpadFilePath: "testdata/exec/reviewpad_with_invalid_group.yml",
			wantErr:                "ProcessGroup:evalGroup expression is not a valid group",
			targetEntity:           engine.DefaultMockTargetEntity,
		},
		"when workflow is invalid": {
			inputReviewpadFilePath: "testdata/exec/reviewpad_with_invalid_workflow.yml",
			wantErr:                "expression 1 is not a condition",
			targetEntity:           engine.DefaultMockTargetEntity,
		},
		"when no workflow is activated": {
			inputReviewpadFilePath: "testdata/exec/reviewpad_with_no_activated_workflows.yml",
			wantProgram: engine.BuildProgram(
				[]*engine.Statement{},
			),
			targetEntity: engine.DefaultMockTargetEntity,
		},
		"when one workflow is activated": {
			inputReviewpadFilePath: "testdata/exec/reviewpad_with_one_activated_workflow.yml",
			wantProgram: engine.BuildProgram(
				[]*engine.Statement{
					engine.BuildStatement(`$addLabel("activate-one-workflow")`),
				},
			),
			targetEntity: engine.DefaultMockTargetEntity,
		},
		"when multiple workflows are activated": {
			inputReviewpadFilePath: "testdata/exec/reviewpad_with_multiple_activated_workflows.yml",
			wantProgram: engine.BuildProgram(
				[]*engine.Statement{
					engine.BuildStatement(`$addLabel("activated-workflow-a")`),
					engine.BuildStatement(`$addLabel("activated-workflow-b")`),
				},
			),
			targetEntity: engine.DefaultMockTargetEntity,
		},
		"when workflow is activated and has extra actions": {
			inputReviewpadFilePath: "testdata/exec/reviewpad_with_activated_workflow_with_extra_actions.yml",
			wantProgram: engine.BuildProgram(
				[]*engine.Statement{
					engine.BuildStatement(`$addLabel("activated-workflow")`),
					engine.BuildStatement(`$addLabel("workflow-with-extra-actions")`),
				},
			),
			targetEntity: engine.DefaultMockTargetEntity,
		},
		"when workflow is skipped": {
			inputReviewpadFilePath: "testdata/exec/reviewpad_with_skipped_workflow.yml",
			wantProgram: engine.BuildProgram(
				[]*engine.Statement{
					engine.BuildStatement(`$addLabel("activated-workflow")`),
				},
			),
			targetEntity: engine.DefaultMockTargetEntity,
		},
		"when run is a single action": {
			inputReviewpadFilePath: "testdata/exec/reviewpad_with_string_run.yml",
			wantProgram: engine.BuildProgram(
				[]*engine.Statement{
					engine.BuildStatement(`$comment("reviewpad")`),
				},
			),
			targetEntity: engine.DefaultMockTargetEntity,
		},
		"when run is a list of actions": {
			inputReviewpadFilePath: "testdata/exec/reviewpad_with_list_of_actions_run.yml",
			wantProgram: engine.BuildProgram(
				[]*engine.Statement{
					engine.BuildStatement(`$comment("reviewpad")`),
					engine.BuildStatement(`$addLabel("test-label")`),
				},
			),
			targetEntity: engine.DefaultMockTargetEntity,
		},
		"when run is a single conditional": {
			inputReviewpadFilePath: "testdata/exec/reviewpad_with_single_conditional_run.yml",
			wantProgram: engine.BuildProgram(
				[]*engine.Statement{
					engine.BuildStatement(`$assignRandomReviewer()`),
					engine.BuildStatement(`$fail("failing for reason")`),
					engine.BuildStatement(`$comment("old workflow style")`),
					engine.BuildStatement(`$comment("reviewpad")`),
				},
			),
			targetEntity: engine.DefaultMockTargetEntity,
		},
		"when run is a single conditional with else": {
			inputReviewpadFilePath: "testdata/exec/reviewpad_with_single_conditional_run_else.yml",
			wantProgram: engine.BuildProgram(
				[]*engine.Statement{
					engine.BuildStatement(`$comment("else")`),
				},
			),
			targetEntity: engine.DefaultMockTargetEntity,
		},
		"when run is a combination": {
			inputReviewpadFilePath: "testdata/exec/reviewpad_with_combination_run.yml",
			wantProgram: engine.BuildProgram(
				[]*engine.Statement{
					engine.BuildStatement(`$comment("reviewpad")`),
					engine.BuildStatement(`$addLabel("label")`),
					engine.BuildStatement(`$removeLabel("remove-label")`),
					engine.BuildStatement(`$assignRandomReviewer()`),
					engine.BuildStatement(`$info("reviewpad supports nested conditions")`),
					engine.BuildStatement(`$comment("combination-run-2")`),
					engine.BuildStatement(`$addLabel("bug")`),
					engine.BuildStatement(`$addLabel("documentation")`),
				},
			),
			targetEntity: engine.DefaultMockTargetEntity,
		},
	}

	codehostClient := aladino.GetDefaultCodeHostClient(t, aladino.GetDefaultPullRequestDetails(), aladino.GetDefaultPullRequestFileList(), nil, nil)

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockedClient := engine.MockGithubClient(test.clientOptions)

			mockedAladinoInterpreter, err := mockAladinoInterpreter(mockedClient, codehostClient)
			if err != nil {
				assert.FailNow(t, fmt.Sprintf("mockAladinoInterpreter: %v", err))
			}

			mockedEnv, err := engine.MockEnvWith(mockedClient, mockedAladinoInterpreter, test.targetEntity, test.eventDetails)
			if err != nil {
				assert.FailNow(t, fmt.Sprintf("engine MockEnvWith: %v", err))
			}

			reviewpadFileData, err := utils.ReadFile(test.inputReviewpadFilePath)
			if err != nil {
				assert.FailNow(t, fmt.Sprintf("Error reading reviewpad file: %v", err))
			}

			reviewpadFile, err := engine.Load(test.inputContext, logger, test.inputGitHubClient, reviewpadFileData)
			if err != nil {
				assert.FailNow(t, fmt.Sprintf("Error parsing reviewpad file: %v", err))
			}

			gotProgram, gotErr := engine.EvalConfigurationFile(reviewpadFile, mockedEnv)

			if gotErr != nil && gotErr.Error() != test.wantErr {
				assert.FailNow(t, fmt.Sprintf("Eval() error = %v, wantErr %v", gotErr, test.wantErr))
			}
			assert.Equal(t, test.wantProgram, gotProgram)
		})
	}
}

func mockAladinoInterpreter(githubClient *gh.GithubClient, codehostClient *codehost.CodeHostClient) (engine.Interpreter, error) {
	dryRun := false
	logger := logrus.NewEntry(logrus.New())
	mockedAladinoInterpreter, err := aladino.NewInterpreter(
		engine.DefaultMockCtx,
		logger,
		dryRun,
		githubClient,
		codehostClient,
		engine.DefaultMockCollector,
		engine.DefaultMockTargetEntity,
		engine.DefaultMockEventPayload,
		aladino.MockBuiltIns(),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("aladino NewInterpreter returned unexpected error: %v", err)
	}

	return mockedAladinoInterpreter, nil
}

func mockGetReposLabelsByOwnerByRepoByName(name, description string) mock.MockBackendOption {
	labelsMockedResponse := &github.Label{}

	if name != "" {
		labelsMockedResponse = &github.Label{
			Name:        github.String(name),
			Description: github.String(description),
		}
	}

	return mock.WithRequestMatch(
		mock.GetReposLabelsByOwnerByRepoByName,
		labelsMockedResponse,
	)
}

func mockPatchReposLabelsByOwnerByRepo(name, description string) mock.MockBackendOption {
	labelsMockedResponse := &github.Label{
		Name:        github.String(name),
		Description: github.String(description),
	}

	return mock.WithRequestMatch(
		mock.PatchReposLabelsByOwnerByRepoByName,
		labelsMockedResponse,
	)
}

func mockPostReposLabelsByOwnerByRepo(name, description string) mock.MockBackendOption {
	labelsMockedResponse := &github.Label{
		Name:        github.String(name),
		Description: github.String(description),
	}

	return mock.WithRequestMatch(
		mock.PostReposLabelsByOwnerByRepo,
		labelsMockedResponse,
	)
}
