package cantabular

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

type Codebook struct {
	Dataset  Dataset             `json:"dataset"`
	CodeBook []CodebookDimension `json:"codebook"`
}

type CodebookDimension struct {
	Name         string   `json:"name"`
	Codes        []string `json:"codes"`
	Label        string   `json:"label"`
	Labels       []string `json:"labels"`
	MapFrom      []string `json:"mapFrom"`
	MapFromCodes []string `json:"mapFromCodes"`
}

func (c *Codebook) GetDimension(name string) *CodebookDimension {
	if c.CodeBook == nil || len(c.CodeBook) == 0 {
		return nil
	}

	for _, cb := range c.CodeBook {
		if cb.Name == name {
			return &cb
		}
	}

	return nil
}
