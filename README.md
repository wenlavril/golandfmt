# golandfmt

GoLand 风格的 Go 代码换行格式化工具，适用于 Cursor / VS Code。

当一行代码超过最大长度（默认 120）时，自动按 GoLand 的 "过长则换行" 规则进行格式化：

- **函数调用实参** — `println(...)` 中的参数
- **复合字面量** — `[]int{...}` 中的元素
- **函数形参** — `func f(...)` 中的参数
- **函数结果形参** — `func f() (...)` 中的返回值

换行时会在一行内尽可能多地放置元素（不是每个元素独占一行），并在最后一个元素后添加尾逗号。

## 安装

```bash
go install golandfmt@latest
```

安装后二进制文件在 `$GOPATH/bin/golandfmt`。

## Cursor / VS Code 配置

在全局设置 `settings.json` 中添加（`Cmd+Shift+P` → `Open User Settings (JSON)`）：

```json
{
    "go.formatTool": "custom",
    "go.alternateTools": {
        "customFormatter": "golandfmt"
    }
}
```

> `golandfmt` 在 `$GOPATH/bin` 下，确保 `$GOPATH/bin` 在你的 `PATH` 中。

## 命令行用法

```bash
# 从 stdin 读取，输出到 stdout
cat main.go | golandfmt

# 格式化文件（输出到 stdout）
golandfmt main.go

# 格式化文件（直接写回）
golandfmt -w main.go

# 自定义最大行宽（默认 120）
golandfmt -m 100 main.go

# 自定义 tab 宽度（默认 4）
golandfmt -t 8 main.go
```

## 示例

输入（超过 40 列时）：

```go
func f() {
	println(100, 101, 102, 103, 104, 105)
}
```

输出：

```go
func f() {
	println(
		100, 101, 102, 103,
		104, 105,
	)
}
```
