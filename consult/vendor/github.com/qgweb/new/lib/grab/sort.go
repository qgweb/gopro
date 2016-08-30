package grab

import (
	"sort"
)

type MapSorter []Item

type Item struct {
	Key string
	Val int
}

func NewMapSorter(m map[string]int) MapSorter {
	ms := make(MapSorter, 0, len(m))

	for k, v := range m {
		ms = append(ms, Item{k, v})
	}

	return ms
}

func (ms MapSorter) Len() int {
	return len(ms)
}

func (ms MapSorter) Less(i, j int) bool {
	return ms[i].Val > ms[j].Val // 按值排序
}

func (ms MapSorter) Swap(i, j int) {
	ms[i], ms[j] = ms[j], ms[i]
}

func (ms *MapSorter) Sort() {
	sort.Sort(ms)
}
