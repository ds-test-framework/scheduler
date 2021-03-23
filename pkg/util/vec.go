package util

import "github.com/zeu5/model-checker/pkg/types"

func PathEq(path1 []*types.Event, path2 []*types.Event) bool {
	if len(path1) != len(path2) {
		return false
	}
	for i := 0; i < len(path1); i++ {
		if path1[i].ID != path2[i].ID {
			return false
		}
	}
	return true
}
