package cantabular

const maxDepth = 2

type Hierarchy struct {
	Label    string
	Children []*Node
}

type Node struct {
	Type     string
	Name     string
	Code     string
	Children []*Node
}

type Index struct {
	Start int
	End   int
	Count int
}

func BuildHierarchyFrom(rootDim *Dimension, cb *Codebook) *Hierarchy {
	h := &Hierarchy{
		Label:    rootDim.Name,
		Children: make([]*Node, 0),
	}

	for i, code := range rootDim.Codes {
		n := &Node{
			Type:     rootDim.Name,
			Name:     rootDim.Labels[i],
			Code:     code,
			Children: rootDim.GetChildrenForOption(code, cb, 1),
		}

		h.Children = append(h.Children, n)
	}

	return h
}

func (d *Dimension) GetChildrenForOption(parentCode string, cb *Codebook, depth int) []*Node {
	children := make([]*Node, 0)
	if depth >= maxDepth {
		return children
	}

	depth += 1

	index, found := d.GetDescendantCodeIndices(parentCode)
	if !found {
		return children
	}

	childDim := cb.GetDimension(d.MapFrom[0])

	for i := index.Start; i <= index.End; i++ {
		children = append(children, &Node{
			Type:     childDim.Name,
			Name:     childDim.Labels[i],
			Code:     childDim.Codes[i],
			Children: childDim.GetChildrenForOption(childDim.Codes[i], cb, depth),
		})
	}
	return children
}

func (d *Dimension) GetDescendantCodeIndices(parentCode string) (*Index, bool) {
	var index *Index

	found := false
	i := 0

	for ; i < len(d.MapFromCodes); i++ {
		code := d.MapFromCodes[i]
		if code == parentCode {
			found = true
			index = &Index{Start: i, End: i, Count: 1}
			continue
		}

		if found && code == "" {
			index.End = i
			index.Count += 1
		} else if found && code != "" {
			break
		}
	}

	return index, found
}
