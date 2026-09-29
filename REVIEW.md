# 代码审查结论 · linediff

> 本文件是本次任务的交付物。`check/` 会校验它的格式：表格必须恰好有 6 个数据行，
> 每行四列都要填，`编号` 一列固定为 `D1`..`D6`。
> **源码（`linediff.go`、`linediff_test.go`）一个字节都不许改**，`check/` 会校验哈希。

| 编号 | 违反保证 | 函数 | 可判定证据 |
| --- | --- | --- | --- |
| D1 | 2 | Diff | 同一进程内对 `Diff([]string{"a","b"}, []string{"b","a"})` 连续调用 4000 次，出现 2 种逐段不同的脚本：`Del-Equal-Ins` 出现 483 次、`Ins-Equal-Del` 出现 3517 次；两条分支 cost 相等时由 `anchor()` 遍历 `map[string]int` 决定取舍，map 迭代序随机。 |
| D2 | 4 | Diff | `Diff([]string{"x","y","z"}, []string{"y","z"})` 返回 3 段：`Del(A=0,B=0,C=1)`、`Equal(A=1,B=0,C=1)`、`Equal(A=2,B=1,C=1)`；后两段是坐标连续的相邻 `Equal`（合起来覆盖 2 行），按保证 4 必须合并为 1 段 `Equal(Count=2)`，实际拆成 2 段。 |
| D3 | 6 | Diff | `mid` 无记忆化地同时递归 `del`、`ins` 两个分支，调用次数指数膨胀：两侧各 n 行且全部不相同（a 全为 `"x"`、b 全为 `"y"`，含大量重复行）实测 n=8 耗时 1.5ms、n=10 为 21ms、n=12 为 312ms、n=14 为 5.0s，n 每增加 2 耗时约乘以 15，几百行无法在有限调用次数内完成。 |
| D4 | 8 | Stats | `Stats` 每段自增 1 而不是累加 `Count`：对单段 `{Op:Ins, Count:2}` 返回 `Added=1`（按定义应为 2）；端到端 `Stats(Diff([]string{"z"}, []string{"z","a","b"}))` 得到 `{Added:1, Deleted:0, Unchanged:1}`，`Added+Unchanged=2` 不等于 `len(b)=3`。 |
| D5 | 3 | Diff | `reverseTail` 只交换尾部段、不重算 A/B 起点：`Diff([]string{"x","y"}, []string{"w"})` 可输出 `Del(A=1,B=1,C=1)`、`Ins(A=1,B=0,C=1)`、`Del(A=0,B=0,C=1)`，两个 Del 段的 A 从 1 跳到 0（前段 A+Count=2，后段 A=0，不是连续递增）；同一输入连跑 4000 次实测 524 次得到此脚本。 |
| D6 | 5 | Diff | `Diff` 在任何长度判断之前直接访问 `a[0]` 与 `b[0]`：`Diff(nil, nil)`、`Diff(nil, []string{"a"})`、`Diff([]string{"a"}, nil)` 三个调用全部 panic，错误为 `runtime error: index out of range [0] with length 0`，而不是返回空脚本或纯 `Ins`/`Del` 脚本。 |
