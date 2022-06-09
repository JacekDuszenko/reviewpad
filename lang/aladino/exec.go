// Copyright 2022 Explore.dev Unipessoal Lda. All Rights Reserved.
// Use of this source code is governed by a license that can be
// found in the LICENSE file.

package aladino

import "fmt"

type ExecExpr interface {
	exec(env Env) error
}

func TypeCheckExec(env Env, expr Expr) (ExecExpr, error) {
	switch expr.Kind() {
	case "FunctionCall":
		_, err := TypeInference(env, expr)
		if err != nil {
			return nil, err
		}

		return expr.(*FunctionCall), nil
	}

	return nil, fmt.Errorf("typecheckexec: %v", expr.Kind())
}

func (fc *FunctionCall) exec(env Env) error {
	args := make([]Value, len(fc.arguments))
	for i, elem := range fc.arguments {
		value, err := elem.Eval(env)

		if err != nil {
			return err
		}

		args[i] = value
	}

	action, ok := env.GetBuiltIns().Actions[fc.name.ident]
	if !ok {
		return fmt.Errorf("exec: %v not found. are you sure this is a built-in function?", fc.name.ident)
	}

	env.GetCollector().Collect("Builtin", &map[string]interface{}{
		"builtin": fc.name.ident,
	})

	return action.Code(env, args)
}

func ExecAction(env Env, expr ExecExpr) error {
	return expr.exec(env)
}