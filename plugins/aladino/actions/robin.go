// Copyright 2023 Explore.dev Unipessoal Lda. All Rights Reserved.
// Use of this source code is governed by a license that can be
// found in the LICENSE file.

package plugins_aladino_actions

import (
	"fmt"

	"github.com/reviewpad/api/go/entities"
	api "github.com/reviewpad/api/go/services"
	converter "github.com/reviewpad/go-lib/converters"
	"github.com/reviewpad/reviewpad/v4/handler"
	"github.com/reviewpad/reviewpad/v4/lang/aladino"
	plugins_aladino_services "github.com/reviewpad/reviewpad/v4/plugins/aladino/services"
)

func Robin() *aladino.BuiltInAction {
	return &aladino.BuiltInAction{
		Type:           aladino.BuildFunctionType([]aladino.Type{aladino.BuildStringType()}, nil),
		Code:           robinCode,
		SupportedKinds: []handler.TargetEntityKind{handler.PullRequest, handler.Issue},
	}
}

func robinCode(e aladino.Env, args []aladino.Value) error {
	target := e.GetTarget()
	targetEntity := target.GetTargetEntity()
	prompt := args[0].(*aladino.StringValue).Val

	service, ok := e.GetBuiltIns().Services[plugins_aladino_services.ROBIN_SERVICE_KEY]
	if !ok {
		return fmt.Errorf("robin service not found")
	}

	robinClient := service.(api.RobinClient)
	req := &api.PromptRequest{
		Prompt: prompt,
		Token:  e.GetGithubClient().GetToken(),
		Target: &entities.TargetEntity{
			Owner:  targetEntity.Owner,
			Repo:   targetEntity.Repo,
			Kind:   converter.ToEntityKind(targetEntity.Kind),
			Number: int32(targetEntity.Number),
		},
	}

	_, err := robinClient.Prompt(e.GetCtx(), req)
	return err
}
