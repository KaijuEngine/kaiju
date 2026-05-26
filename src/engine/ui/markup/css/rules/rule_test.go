/******************************************************************************/
/* rule_test.go                                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rules

import "testing"

func TestRuleInvokeMatchesComposedState(t *testing.T) {
	state := RuleInvokeHover.With(RuleInvokeActive)
	if !RuleInvokeHover.Matches(state) {
		t.Fatal("hover should match hover+active state")
	}
	if !RuleInvokeActive.Matches(state) {
		t.Fatal("active should match hover+active state")
	}
	if !state.Matches(state) {
		t.Fatal("combined rule should match the exact combined state")
	}
	if RuleInvokeFocus.Matches(state) {
		t.Fatal("focus should not match hover+active state")
	}
}

func TestCloneRulesDoesNotShareValues(t *testing.T) {
	original := []Rule{{
		Property:   "color",
		Invocation: RuleInvokeHover,
		Values: []PropertyValue{{
			Str:  "rgb",
			Args: []string{"1", "2", "3"},
		}},
	}}
	cloned := CloneRules(original)
	cloned[0].Invocation = cloned[0].Invocation.With(RuleInvokeActive)
	cloned[0].Values[0].Args[0] = "9"

	if original[0].Invocation != RuleInvokeHover {
		t.Fatal("clone changed original invocation")
	}
	if original[0].Values[0].Args[0] != "1" {
		t.Fatal("clone shared property value args with original")
	}
}
