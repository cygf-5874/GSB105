# 代码审查结论 · linediff

> 本文件是本次任务的交付物。`check/` 会校验它的格式：表格必须恰好有 6 个数据行，
> 每行四列都要填，`编号` 一列固定为 `D1`..`D6`。
> **源码（`linediff.go`、`linediff_test.go`）一个字节都不许改**，`check/` 会校验哈希。

| 编号 | 违反保证 | 函数 | 可判定证据 |
| --- | --- | --- | --- |
| D1 | 2 | Diff | 对同一输入 `Diff([]string{"a","b"}, []string{"b","a"})` 连续调用 2000 次，实测 253 次返回的首段与首次不同：等代价时 `anchor()` 遍历 `map[string]int`，Go map 迭代顺序随机，结果随迭代顺序变。 |
| D2 | 4 | Diff | `Diff([]string{"x","m","n"}, []string{"m","n"})` 返回 3 段，其中第 2、3 段是相邻的两个 `Equal`（`{Equal A:1 B:0 Count:1}` 后紧跟 `{Equal A:2 B:1 Count:1}`），未合并成 1 段 Count=2。 |
| D3 | 6 | Diff | 两侧各 n 行全重复且首行不等（如 n 个 `"x"` 对 n 个 `"y"`）时 `mid` 无记忆化双路递归：实测 n=8/9/10/11/12 的 mid 调用次数为 25739/97239/369511/1410863/5408311（=2·C(2n,n)-1，约 4 倍/行），n=300 永不返回。 |
| D4 | 8 | Stats | `Diff([]string{"x","m"}, []string{"m","n","p"})` 末尾是 1 个 `Ins` 段 Count=2，`Stats` 却给出 `Added:1`（按段数加 1 而非按 `e.Count` 加 2），且 1+Unchanged(1)=2 ≠ len(b)=3。 |
| D5 | 3 | Diff | `Diff([]string{"x","x","x"}, []string{"y"})` 返回 `[{Ins A:3 B:0} {Del A:2} {Del A:1} {Del A:0}]`，Equal/Del 段的 A 起点为 2、1、0，非连续递增（应为 0、1、2 连续覆盖 a）。 |
| D6 | 5 | Diff | `Diff(nil, nil)` 直接 panic：`runtime error: index out of range [0] with length 0`，因为函数体无判空就读 `a[0]`、`b[0]`；任一侧为空切片同样 panic。 |
