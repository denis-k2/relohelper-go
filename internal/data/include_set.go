package data

type IncludeSet map[string]struct{}

func NewIncludeSet(values ...string) IncludeSet {
	set := make(IncludeSet, len(values))
	for _, value := range values {
		set[value] = struct{}{}
	}

	return set
}

func (s IncludeSet) Has(value string) bool {
	_, ok := s[value]
	return ok
}
