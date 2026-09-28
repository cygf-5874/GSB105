// Command check 是 linediff 审查题的固定验收入口。
//
// ⚠️ 不要修改本文件。
//
// 它只校验 REVIEW.md 的交付格式与可判定性，以及「编号 → 违反保证 / 函数」是否落在
// 允许集合内；结论是否正确由人按 README 逐条复核。8 个场景：
//
//	unmodified 1：6 个不可修改文件的 SHA-256 未变
//	format     1：REVIEW.md 的表格式合法、恰好 6 行且四列齐全
//	finding    6：D1..D6 各自的「违反保证」编号与「函数」命中允许集合
//
// 用法：
//
//	go run ./check                  # 跑全部场景
//	go run ./check --only finding   # 只跑一组
//	go run ./check -list            # 列出场景
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// baseline 是不可修改文件的 SHA-256（按「去掉 \r 后的内容」计算）。
var baseline = []struct {
	path string
	sum  string
}{
	{"linediff.go", "cfe0e5c888e8c5d21e6901e25c6d78afbbceacffaae771c9a06426c866ca09a6"},
	{"linediff_test.go", "a85bb299182508f56844472811178d1c13b7b2c20da4d3a1f875107b3544656d"},
	{"go.mod", "5948185ef36fa82ea37dadf43f9763ba43098197dbb6fc9881c4ca3778002201"},
	{"README.md", "53b74a909d2b05aab9d8a75f64b30e69f1d9e9a5330fe33d2e92d7ae4e9dc79c"},
	{"PROMPT.md", "9bb0ac5c92848fcaccc963096edac11309a899a6f3929461f4980c74ab7215cf"},
	{"scripts/check.sh", "cc5a06c405a474199d98daa52337b0c60c0867148dd8d5c227c2312aa1acac25"},
}

// allowed 是每个缺陷「违反保证」编号与「函数」的允许集合。
var allowed = map[string]struct {
	guarantees map[int]bool
	funcs      map[string]bool
}{
	"D1": {map[int]bool{2: true}, map[string]bool{"Diff": true}},
	"D2": {map[int]bool{4: true}, map[string]bool{"Diff": true}},
	"D3": {map[int]bool{6: true, 7: true}, map[string]bool{"Diff": true}},
	"D4": {map[int]bool{8: true}, map[string]bool{"Stats": true}},
	"D5": {map[int]bool{3: true}, map[string]bool{"Diff": true}},
	"D6": {map[int]bool{5: true}, map[string]bool{"Diff": true}},
}

var defectIDs = []string{"D1", "D2", "D3", "D4", "D5", "D6"}

var bannedWords = []string{
	"看起来", "可能", "也许", "大概", "疑似", "或许", "建议",
	"需要注意", "有待", "风险", "不太好", "问题不大",
}

type scenario struct {
	group string
	name  string
	run   func() error
}

type review struct {
	rows [][]string // 数据行，每行 4 列
}

func main() {
	only := ""
	list := false
	for i := 1; i < len(os.Args); i++ {
		a := os.Args[i]
		switch {
		case a == "-list":
			list = true
		case a == "--only":
			if i+1 < len(os.Args) {
				only = os.Args[i+1]
				i++
			}
		case strings.HasPrefix(a, "--only="):
			only = strings.TrimPrefix(a, "--only=")
		}
	}

	scenarios := buildScenarios(loadReview())

	if list {
		for _, sc := range scenarios {
			fmt.Printf("%s/%s\n", sc.group, sc.name)
		}
		return
	}

	passed, total := 0, 0
	for _, sc := range scenarios {
		if only != "" && sc.group != only {
			continue
		}
		total++
		if err := sc.run(); err != nil {
			fmt.Printf("FAIL %s/%s  期望=%s\n", sc.group, sc.name, err.Error())
			continue
		}
		passed++
		fmt.Printf("PASS %s/%s\n", sc.group, sc.name)
	}
	fmt.Printf("结果：通过 %d/%d\n", passed, total)
	if passed != total {
		os.Exit(1)
	}
}

func buildScenarios(rv review) []scenario {
	out := []scenario{
		{"unmodified", "unmodified_sources", checkUnmodified},
		{"format", "table_shaped", func() error { return rv.checkShape() }},
	}
	for _, id := range defectIDs {
		id := id
		out = append(out, scenario{"finding", id, func() error { return rv.checkFinding(id) }})
	}
	return out
}

// ------------------------------------------------------------------ 读取

func loadReview() review {
	var rv review
	data, err := os.ReadFile("REVIEW.md")
	if err != nil {
		return rv
	}
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	for _, line := range strings.Split(text, "\n") {
		t := strings.TrimSpace(line)
		if !strings.HasPrefix(t, "|") {
			continue
		}
		parts := strings.Split(t, "|")
		// parts[0] 与 parts[len-1] 是首尾的空串
		if len(parts) < 3 {
			continue
		}
		cells := make([]string, 0, len(parts)-2)
		for i := 1; i < len(parts)-1; i++ {
			cells = append(cells, strings.TrimSpace(parts[i]))
		}
		joined := strings.Join(cells, "")
		if strings.Contains(joined, "编号") || strings.Contains(joined, "违反保证") ||
			strings.Contains(joined, "可判定证据") {
			continue // 表头
		}
		allSep := true
		for _, c := range cells {
			if strings.Trim(c, "-: ") != "" {
				allSep = false
			}
		}
		if allSep {
			continue // 分隔行
		}
		rv.rows = append(rv.rows, cells)
	}
	return rv
}

// -------------------------------------------------------------- unmodified

func checkUnmodified() error {
	for _, b := range baseline {
		data, err := os.ReadFile(b.path)
		if err != nil {
			return fmt.Errorf("%s 存在且未改动 实际=读取失败 %v", b.path, err)
		}
		sum := sha256Hex(data)
		if sum != b.sum {
			return fmt.Errorf("%s 保持原样 实际=SHA-256 与基线不一致", b.path)
		}
	}
	return nil
}

func sha256Hex(data []byte) string {
	clean := make([]byte, 0, len(data))
	for _, c := range data {
		if c != '\r' {
			clean = append(clean, c)
		}
	}
	sum := sha256.Sum256(clean)
	return hex.EncodeToString(sum[:])
}

// ----------------------------------------------------------------- format

func (rv review) rowFor(id string) ([]string, bool) {
	for _, row := range rv.rows {
		if len(row) >= 1 && strings.EqualFold(strings.TrimSpace(row[0]), id) {
			return row, true
		}
	}
	return nil, false
}

func (rv review) checkShape() error {
	if len(rv.rows) != 6 {
		return fmt.Errorf("REVIEW.md 的表恰好 6 个数据行 实际=%d 行", len(rv.rows))
	}
	seen := map[string]bool{}
	for i, row := range rv.rows {
		if len(row) != 4 {
			return fmt.Errorf("第 %d 行 4 列 实际=%d 列", i+1, len(row))
		}
		for c, cell := range row {
			if strings.TrimSpace(cell) == "" {
				return fmt.Errorf("第 %d 行第 %d 列非空 实际=空", i+1, c+1)
			}
		}
		seen[strings.ToUpper(strings.TrimSpace(row[0]))] = true
	}
	for _, id := range defectIDs {
		if !seen[id] {
			return fmt.Errorf("编号 %s 出现 实际=缺失", id)
		}
	}
	return nil
}

// ---------------------------------------------------------------- finding

func (rv review) checkFinding(id string) error {
	row, ok := rv.rowFor(id)
	if !ok {
		return fmt.Errorf("REVIEW.md 有编号 %s 的行 实际=缺失", id)
	}
	if len(row) != 4 {
		return fmt.Errorf("%s 行 4 列 实际=%d 列", id, len(row))
	}
	al := allowed[id]

	g, err := strconv.Atoi(strings.TrimSpace(row[1]))
	if err != nil || !al.guarantees[g] {
		return fmt.Errorf("%s 的「违反保证」∈ %s 实际=%q", id, guaranteeSet(al.guarantees), row[1])
	}
	fn := strings.TrimSpace(row[2])
	if !al.funcs[fn] {
		return fmt.Errorf("%s 的「函数」∈ %s 实际=%q", id, funcSet(al.funcs), row[2])
	}
	ev := strings.TrimSpace(row[3])
	if len([]rune(ev)) < 8 {
		return fmt.Errorf("%s 的「可判定证据」足够具体（>=8 字） 实际=%q", id, ev)
	}
	for _, w := range bannedWords {
		if strings.Contains(ev, w) {
			return fmt.Errorf("%s 的「可判定证据」不含无法核对的措辞 实际=出现「%s」", id, w)
		}
	}
	if !strings.ContainsAny(ev, "0123456789\"[]") {
		return fmt.Errorf("%s 的「可判定证据」含具体输入或计数（数字 / 引号 / 切片） 实际=%q", id, ev)
	}
	return nil
}

func guaranteeSet(m map[int]bool) string {
	ids := []int{}
	for k := range m {
		ids = append(ids, k)
	}
	// 升序，避免 map 顺序影响提示文本
	for i := 0; i < len(ids); i++ {
		for j := i + 1; j < len(ids); j++ {
			if ids[j] < ids[i] {
				ids[i], ids[j] = ids[j], ids[i]
			}
		}
	}
	parts := make([]string, len(ids))
	for i, v := range ids {
		parts[i] = strconv.Itoa(v)
	}
	return "{" + strings.Join(parts, ",") + "}"
}

func funcSet(m map[string]bool) string {
	names := []string{}
	for k := range m {
		names = append(names, k)
	}
	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			if names[j] < names[i] {
				names[i], names[j] = names[j], names[i]
			}
		}
	}
	return "{" + strings.Join(names, ",") + "}"
}
