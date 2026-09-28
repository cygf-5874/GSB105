package linediff

import "testing"

func countOp(edits []Edit, op Op) int {
	n := 0
	for _, e := range edits {
		if e.Op == op {
			n += e.Count
		}
	}
	return n
}

func TestOpString(t *testing.T) {
	if Equal.String() != "Equal" || Del.String() != "Del" || Ins.String() != "Ins" {
		t.Fatalf("Op.String() 结果不对：%q %q %q", Equal, Del, Ins)
	}
}

func TestDiffIdenticalSingle(t *testing.T) {
	got := Diff([]string{"a"}, []string{"a"})
	if len(got) != 1 || got[0].Op != Equal || got[0].Count != 1 {
		t.Fatalf("相同单行应当是一段 Equal(1)，得到 %+v", got)
	}
}

func TestDiffIdenticalMulti(t *testing.T) {
	got := Diff([]string{"a", "b", "c"}, []string{"a", "b", "c"})
	if len(got) != 1 || got[0].Op != Equal || got[0].Count != 3 {
		t.Fatalf("相同多行应当合并成一段 Equal(3)，得到 %+v", got)
	}
}

func TestDiffReplaceStats(t *testing.T) {
	got := Stats(Diff([]string{"x"}, []string{"y"}))
	want := Counts{Added: 1, Deleted: 1, Unchanged: 0}
	if got != want {
		t.Fatalf("替换单行的统计应当是 %+v，得到 %+v", want, got)
	}
}

func TestDiffInsertStats(t *testing.T) {
	got := Stats(Diff([]string{"x"}, []string{"x", "y"}))
	want := Counts{Added: 1, Deleted: 0, Unchanged: 1}
	if got != want {
		t.Fatalf("末尾插入的统计应当是 %+v，得到 %+v", want, got)
	}
}

func TestDiffDeleteStats(t *testing.T) {
	got := Stats(Diff([]string{"x", "y"}, []string{"x"}))
	want := Counts{Added: 0, Deleted: 1, Unchanged: 1}
	if got != want {
		t.Fatalf("末尾删除的统计应当是 %+v，得到 %+v", want, got)
	}
}

func TestDiffDeleteHeadStats(t *testing.T) {
	got := Stats(Diff([]string{"x", "y"}, []string{"y"}))
	want := Counts{Added: 0, Deleted: 1, Unchanged: 1}
	if got != want {
		t.Fatalf("删除首行的统计应当是 %+v，得到 %+v", want, got)
	}
}

func TestDiffCoversBothSides(t *testing.T) {
	a := []string{"p", "q"}
	b := []string{"q", "p"}
	edits := Diff(a, b)
	if got := countOp(edits, Equal) + countOp(edits, Del); got != len(a) {
		t.Fatalf("Equal+Del 的行数应当覆盖 a 的 %d 行，得到 %d", len(a), got)
	}
	if got := countOp(edits, Equal) + countOp(edits, Ins); got != len(b) {
		t.Fatalf("Equal+Ins 的行数应当覆盖 b 的 %d 行，得到 %d", len(b), got)
	}
}

func TestDiffAllDifferentCounts(t *testing.T) {
	a := []string{"a", "b", "c"}
	b := []string{"x", "y", "z"}
	edits := Diff(a, b)
	if got := countOp(edits, Equal) + countOp(edits, Del); got != len(a) {
		t.Fatalf("应当覆盖 a 的 %d 行，得到 %d", len(a), got)
	}
	if got := countOp(edits, Equal) + countOp(edits, Ins); got != len(b) {
		t.Fatalf("应当覆盖 b 的 %d 行，得到 %d", len(b), got)
	}
}

func TestDiffInsertMiddleStats(t *testing.T) {
	edits := Diff([]string{"a", "c"}, []string{"a", "b", "c"})
	want := Counts{Added: 1, Deleted: 0, Unchanged: 2}
	if got := Stats(edits); got != want {
		t.Fatalf("中间插入的统计应当是 %+v，得到 %+v", want, got)
	}
}

func TestDiffOpsAreValid(t *testing.T) {
	edits := Diff([]string{"a", "b"}, []string{"b"})
	for i, e := range edits {
		if e.Op != Equal && e.Op != Del && e.Op != Ins {
			t.Fatalf("第 %d 段的 Op 非法：%v", i, e.Op)
		}
		if e.Count <= 0 {
			t.Fatalf("第 %d 段的 Count 应当为正：%+v", i, e)
		}
	}
}

func TestStatsSumInvariant(t *testing.T) {
	a := []string{"x", "y"}
	b := []string{"y"}
	s := Stats(Diff(a, b))
	if s.Added+s.Deleted+s.Unchanged == 0 {
		t.Fatal("统计不应当全为 0")
	}
	if s.Deleted != 1 || s.Unchanged != 1 {
		t.Fatalf("应当有 1 行删除、1 行未变，得到 %+v", s)
	}
}
