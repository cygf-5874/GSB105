# linediff

按行计算两段文本之间的**编辑脚本**（Go 1.24，仅标准库，`go.mod` 无 `require`）。

`Diff(a, b []string) []Edit` 返回把 `a` 变成 `b` 的最小编辑脚本，
每一段是 `Equal` / `Del` / `Ins` 之一；`Stats` 按行统计脚本。

```go
edits := linediff.Diff([]string{"a", "b"}, []string{"a", "c"})
// [Equal a, Del b, Ins c]（示意）
linediff.Stats(edits) // {Added:1, Deleted:1, Unchanged:1}
```

## 怎么跑

```bash
go build ./...            # 编译
go test ./...             # 既有用例
go run ./check            # 固定验收程序（8 个场景）
go run ./check -list      # 列出全部场景
go run ./check --only finding
bash scripts/check.sh     # 薄封装：go run ./check "$@"
```

前提：Go 1.24。`check/`、`scripts/` 是固定验收的一部分，**请勿修改**。

## 本次交付物

这次**不写代码**：请通读 `linediff.go`，把 `REVIEW.md` 里的那张表**填满** ——
每个问题一行，写清它**违反哪一条保证**、出在**哪个函数**、以及**一句可判定的证据**。

- 表头已给定：`| 编号 | 违反保证 | 函数 | 可判定证据 |`，编号固定为 `D1`..`D6`。
- `违反保证` 填 1~8 的**编号**；`函数` 填函数名（`Diff` / `Stats`）。
- `可判定证据` 必须**可复现**：给出一句话，含具体的输入形状、计数或可观察差异，
  不许写「看起来不对」「可能有风险」这类无法核对的话。
- **源码一个字节都不许改**（`linediff.go`、`linediff_test.go` 等），`check/` 会校验哈希。

## 对外保证

下面 8 条是 `linediff.go` 这份实现的契约。**它们是契约，不是「当前行为」的转述。**

1. **最小**：`Diff(a, b []string) []Edit` 返回的编辑脚本必须**最小** ——
   非 `Equal` 段覆盖的行数（即 `Del` + `Ins` 的行数）最少。
2. **确定性**：同一输入任意多次调用结果**完全相同**（逐段相同），不得依赖 `map` 迭代顺序。
3. **合法且连续**：`Edit.Op` 只能是 `Equal` / `Del` / `Ins`；把脚本里所有 `Equal`/`Del`
   段的 `A` 依次取出，必须**连续递增**（前一段 `A + Count == 后一段 A`），
   `Equal`/`Ins` 段的 `B` 也必须**连续递增**。
4. **`Equal` 尽量长**：连续的相等行必须**合并成一段** `Equal`，不得拆成多段。
5. **空输入**：两侧都空、或一侧为空时，必须返回空脚本或纯 `Ins` / `Del` 脚本，**不得 panic**。
6. **有限资源**：含大量重复行（例如几百行全是同一字符串）的输入必须在**有限内存与有限
   调用次数**内完成 —— 不得递归到爆栈，也不得让递归调用次数**指数膨胀**。
7. **近似 O(n)**：对编辑距离小的输入，耗时随行数近似线性，不得退化成指数级回溯。
8. **统计自洽**：`Stats(edits []Edit) Counts` 的 `Added` / `Deleted` / `Unchanged`
   分别是 `Ins` / `Del` / `Equal` 段的**行数**之和；并且对 `a`、`b` 与 `edits := Diff(a, b)`
   必须满足 `Deleted + Unchanged == len(a)`、`Added + Unchanged == len(b)`。

## 固定件场景（8 个）

| 分组 | 场景 | 覆盖 |
| --- | --- | --- |
| `unmodified` | `unmodified_sources` | 6 个不可修改文件的 SHA-256 未变 |
| `format` | `table_shaped` | `REVIEW.md` 的表格式合法、恰好 6 行且四列齐全 |
| `finding` | `D1`..`D6` | 每个缺陷的「违反保证」编号与「函数」命中允许集合，证据可判定 |

`--only <组名>`（`unmodified` / `format` / `finding`）可以只跑一组。
8 个场景都在同一个进程里跑，失败不早退。

## 目录

```
.
├── .gitattributes
├── .gitignore
├── go.mod               module linediff
├── linediff.go          待评审的实现（只审不改）
├── linediff_test.go     既有用例
├── PROMPT.md            本题的 User Prompt（逐字）
├── README.md
├── REVIEW.md            本次交付物（把表填满）
├── check/main.go        固定验收程序（勿改）
└── scripts/check.sh     薄封装：go run ./check "$@"
```
