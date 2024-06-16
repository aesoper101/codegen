package types

type feature int

const (
	placeHolder feature = iota
)

var (
	featureMap = make(map[string]feature)
	maxFeature = placeHolder
)

// HasFeature check whether a feature in the list
func HasFeature(list []feature, key string) bool {
	target, ok := getFeature(key)
	if !ok {
		return false
	}
	for _, f := range list {
		if f == target {
			return true
		}
	}
	return false
}

// RegisterFeature register a feature
func RegisterFeature(key string) {
	if _, ok := featureMap[key]; !ok {
		maxFeature++
		featureMap[key] = maxFeature
	}
}

func getFeature(key string) (feature, bool) {
	f, ok := featureMap[key]
	return f, ok
}
