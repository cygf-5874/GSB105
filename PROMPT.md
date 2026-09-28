这次不写代码，只出一份评审意见。linediff 是 Go 1.24 的行级差异库，
只用标准库（无第三方依赖），构建走 go 工具链，
自检走 `scripts/check.sh`（`check/` 是固定验收程序，别改），既有用例走 `go test ./...`。

README 里的「对外保证」8 条是这份实现的契约，`linediff.go` 里有若干处没有守住。

任务：通读源码，把 `REVIEW.md` 里那张表填满 —— 每个问题一行，
写清它违反哪一条保证、出在哪个函数、以及一句可判定的证据（要能复现，别人照你说的输入能看出区别）。

验收：
- go run ./check 退出码 0，8 个场景全过（unmodified 1 + format 1 + finding 6）；
- 源码文件一个字节都不许改，`check/` 会校验哈希。

约束：
1. 只改 `REVIEW.md`；`linediff.go`、`linediff_test.go` 一律不动。
2. 不改 `check/`。
3. 不引入第三方依赖，`go.mod` 里不许出现 `require`。
4. 证据一列不要写「看起来不对」「可能有风险」这类话，要写可复现的输入或可数的量。
