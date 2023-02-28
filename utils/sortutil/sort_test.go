package sortutil

import "testing"

type req struct {
	val    string
	sortId int
}

func TestSortStruct(t *testing.T) {
	var data []*req

	data = append(data, &req{"a", 3})
	data = append(data, &req{"b", 1})
	data = append(data, &req{"m", 9})

	Cmp[*req](func(p1, p2 **req) bool {
		return (*p1).sortId < (*p2).sortId
	}).Sort(data)

	var ret []string
	for _, r := range data {
		ret = append(ret, r.val)
	}

	if !stringSliceEqual(ret, []string{"b", "a", "m"}) {
		t.Errorf("sort error\n")
	}

}

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	if (a == nil) != (b == nil) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}
