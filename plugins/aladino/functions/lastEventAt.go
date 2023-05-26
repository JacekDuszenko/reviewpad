// Copyright 2022 Explore.dev Unipessoal Lda. All Rights Reserved.
// Use of this source code is governed by a license that can be
// found in the LICENSE file.

package plugins_aladino_functions

import (
	"time"

	"github.com/reviewpad/go-lib/entities"
	"github.com/reviewpad/reviewpad/v4/lang"
	"github.com/reviewpad/reviewpad/v4/lang/aladino"
)

func LastEventAt() *aladino.BuiltInFunction {
	return &aladino.BuiltInFunction{
		Type:           lang.BuildFunctionType([]lang.Type{}, lang.BuildIntType()),
		Code:           lastEventAtCode,
		SupportedKinds: []entities.TargetEntityKind{entities.PullRequest, entities.Issue},
	}
}

func lastEventAtCode(e aladino.Env, args []lang.Value) (lang.Value, error) {
	updatedAtTime, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", e.GetTarget().GetUpdatedAt())
	if err != nil {
		return nil, err
	}

	return lang.BuildIntValue(int(updatedAtTime.Unix())), nil
}
