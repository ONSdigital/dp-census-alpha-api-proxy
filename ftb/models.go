package ftb

type Datasets struct {
	Items []*Dataset `json:"items,omitempty"`
}

type Dataset struct {
	Name             string `json:"name,omitempty"`
	Description      string `json:"description,omitempty"`
	Size             int    `json:"size,omitempty"`
	RuleRootVariable string `json:"rule_root_variable,omitempty"`
	Digest           string `json:"digest,omitempty"`
}
