package rules

type RuleInvoke = int

const (
	RuleInvokeImmediate RuleInvoke = iota
	RuleInvokeHover
)

type PropertyValue struct {
	Str  string
	Args []string
}

func (p PropertyValue) IsFunction() bool {
	return len(p.Args) > 0
}

type Rule struct {
	Property   string
	Values     []PropertyValue
	Invocation RuleInvoke
}
