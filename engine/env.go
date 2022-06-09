// Copyright 2022 Explore.dev Unipessoal Lda. All Rights Reserved.
// Use of this source code is governed by a license that can be
// found in the LICENSE file.

package engine

import (
	"context"

	"github.com/google/go-github/v42/github"
	"github.com/reviewpad/reviewpad/collector"
	"github.com/shurcooL/githubv4"
)

type GroupKind string
type GroupType string

const GroupKindDeveloper GroupKind = "developer"
const GroupTypeStatic GroupType = "static"
const GroupTypeFilter GroupType = "filter"

type Interpreter interface {
	ProcessGroup(name string, kind GroupKind, typeOf GroupType, expr, paramExpr, whereExpr string) error
	EvalExpr(kind, expr string) (bool, error)
	ExecActions(program *[]string) error
}

type Env struct {
	Ctx          context.Context
	Client       *github.Client
	ClientGQL    *githubv4.Client
	Collector    collector.Collector
	PullRequest  *github.PullRequest
	Interpreters map[string]Interpreter
}

func NewEvalEnv(
	ctx context.Context,
	client *github.Client,
	clientGQL *githubv4.Client,
	collector collector.Collector,
	pullRequest *github.PullRequest,
	interpreters map[string]Interpreter,
) (*Env, error) {
	input := &Env{
		Ctx:          ctx,
		Client:       client,
		ClientGQL:    clientGQL,
		Collector:    collector,
		PullRequest:  pullRequest,
		Interpreters: interpreters,
	}

	return input, nil
}