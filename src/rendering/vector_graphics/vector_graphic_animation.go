package vector_graphics

// CalcMode represents animation calculation modes for SVG SMIL animations
type CalcMode int8

const (
	CalcModeDiscrete CalcMode = iota
	CalcModeLinear
	CalcModePaced
	CalcModeSpline
)

type Animate struct {
	AttributeName string
	AttributeType string
	From          string
	To            string
	By            string
	Values        string
	KeyTimes      string
	KeySplines    string
	CalcMode      CalcMode
	Dur           string
	Begin         string
	End           string
	Min           string
	Max           string
	RepeatCount   string
	RepeatDur     string
	Fill          FillMode
	Restart       RestartMode
	Additive      Additive
	Accumulate    Accumulate
}
