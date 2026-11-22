package configuration

type FilterRule struct {
	Key             string `json:"key"`
	UnparsedMatcher string `json:"match"`
	Name            string `json:"name"`
}

type SegmentRule struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type MaskingRule struct {
	Key        string `json:"key"`
	Fields     []int  `json:"fields"`
	Everything bool   `json:"everything"`
	Name       string `json:"name"`
}

type Configuration struct {
	GatedURL       string        `json:"gated_url"`
	HeaderRules    []FilterRule  `json:"header_rules"`
	ParameterRules []FilterRule  `json:"parameter_rules"`
	SegmentRules   []SegmentRule `json:"segment_rules"`
	MaskingRules   []MaskingRule `json:"masking_rules"`
}
