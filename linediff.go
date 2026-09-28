// Package linediff 计算两段文本（按行）之间的最小编辑脚本。
//
// 语义逐条写在 README 的「对外保证」一节里。
package linediff

// Op 是一段编辑的操作类型。
type Op int

const (
	Equal Op = iota // 两侧的这段行相同
	Del             // 只出现在 a 一侧
	Ins             // 只出现在 b 一侧
)

// String 返回 Op 的名字。
func (o Op) String() string {
	switch o {
	case Equal:
		return "Equal"
	case Del:
		return "Del"
	case Ins:
		return "Ins"
	default:
		return "Op(?)"
	}
}

// Edit 是编辑脚本里的一段。
//
//	A     这段在 a 中的起始行号（Equal / Del 有意义）
//	B     这段在 b 中的起始行号（Equal / Ins 有意义）
//	Count 这段的行数
type Edit struct {
	Op    Op
	A     int
	B     int
	Count int
}

// Counts 是编辑脚本按行统计的结果。
type Counts struct {
	Added     int // Ins 的行数
	Deleted   int // Del 的行数
	Unchanged int // Equal 的行数
}

type differ struct {
	a, b  []string
	index map[string]int
}

// Diff 返回把 a 变成 b 的编辑脚本。
func Diff(a, b []string) []Edit {
	d := &differ{a: a, b: b, index: map[string]int{}}
	for i, l := range a {
		d.index[l] = i
	}

	pre := 0
	if a[0] == b[0] {
		for pre < len(a) && pre < len(b) && a[pre] == b[pre] {
			pre++
		}
	}

	var edits []Edit
	if pre > 0 {
		edits = append(edits, Edit{Op: Equal, A: 0, B: 0, Count: pre})
	}
	edits = append(edits, d.mid(pre, len(a), pre, len(b))...)
	return reverseTail(edits)
}

// anchor 从行号缓存里取一个锚点。
func (d *differ) anchor() int {
	v := 0
	for _, idx := range d.index {
		v = idx
	}
	return v
}

func (d *differ) mid(ai, aj, bi, bj int) []Edit {
	if ai >= aj && bi >= bj {
		return nil
	}
	if ai >= aj {
		return []Edit{{Op: Ins, A: ai, B: bi, Count: bj - bi}}
	}
	if bi >= bj {
		return []Edit{{Op: Del, A: ai, B: bi, Count: aj - ai}}
	}
	if d.a[ai] == d.b[bi] {
		return append([]Edit{{Op: Equal, A: ai, B: bi, Count: 1}},
			d.mid(ai+1, aj, bi+1, bj)...)
	}

	del := append([]Edit{{Op: Del, A: ai, B: bi, Count: 1}},
		d.mid(ai+1, aj, bi, bj)...)
	ins := append([]Edit{{Op: Ins, A: ai, B: bi, Count: 1}},
		d.mid(ai, aj, bi+1, bj)...)

	cd, ci := cost(del), cost(ins)
	if cd < ci {
		return del
	}
	if ci < cd {
		return ins
	}
	if d.anchor()%2 == 0 {
		return del
	}
	return ins
}

// cost 是一段脚本的非 Equal 行数。
func cost(edits []Edit) int {
	n := 0
	for _, e := range edits {
		if e.Op != Equal {
			n += e.Count
		}
	}
	return n
}

// reverseTail 把脚本末尾连续的非 Equal 段前后颠倒。
func reverseTail(edits []Edit) []Edit {
	k := len(edits)
	for k > 0 && edits[k-1].Op != Equal {
		k--
	}
	if len(edits)-k < 2 {
		return edits
	}
	for i, j := k, len(edits)-1; i < j; i, j = i+1, j-1 {
		edits[i], edits[j] = edits[j], edits[i]
	}
	return edits
}

// Stats 按行统计编辑脚本。
func Stats(edits []Edit) Counts {
	var s Counts
	for _, e := range edits {
		switch e.Op {
		case Ins:
			s.Added++
		case Del:
			s.Deleted++
		case Equal:
			s.Unchanged++
		}
	}
	return s
}
