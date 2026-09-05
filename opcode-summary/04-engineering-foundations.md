# 配置、国际化、平台适配与测试

## 1. 配置系统

配置代码位于 [`internal/app/config.go`](../internal/app/config.go)。默认路径为：

```text
~/.config/lazyocs/config.toml
```

如果设置了 `XDG_CONFIG_HOME`，则使用：

```text
$XDG_CONFIG_HOME/lazyocs/config.toml
```

首次运行时会创建默认配置：

- 目录权限：`0700`
- 文件权限：`0600`
- 使用 `O_EXCL`，不会覆盖已经存在的配置

这体现了两个配置文件原则：默认可发现，同时不破坏用户已有内容。

## 2. 配置优先级

合并流程由 `ResolveOptions` 实现：

```text
程序默认值
  < 默认或显式 TOML 配置
  < 用户明确提供的 CLI flags
```

程序使用 `flag.Visit` 区分“参数具有默认值”和“用户真的输入了这个参数”。例如 `--limit` 默认是 500，如果用户没有输入它，就不能让 CLI 默认值错误覆盖配置文件中的 1000。

恢复会话命令使用 `[resume]` 模块表达：`command` 是可执行文件，`args` 是固定参数，`session_args` 负责注入 `{session_id}`。这避免把 `opencode --proxy` 当作单个可执行文件，也避免通过 shell 解析用户配置。

对应文件：

- [`cmd/lazyocs/main.go`](../cmd/lazyocs/main.go) 的 `visitedFlags`
- [`internal/app/config.go`](../internal/app/config.go) 的 `ResolveOptions`
- [`internal/app/config_test.go`](../internal/app/config_test.go) 的优先级测试

配置还支持详情字段级开关，并接受 TOML 的 `true/false` 和 `1/0`。

当前需要改进：

- 无效字段值目前会被静默忽略，应返回警告。
- 自动创建默认配置失败目前也被忽略，应至少输出可操作错误。
- `--check` 不应产生创建配置文件的副作用。
- 配置中的 `read_only=true` 缺少 CLI 反向覆盖方式，需要明确产品语义。

## 3. 数据库路径解析

数据库路径支持：

```text
--db
LAZYOCS_DB
默认 ~/.local/share/opencode/opencode.db
```

`resolveDBPath` 会：

1. 展开 `~/`。
2. 转换成绝对路径。
3. 检查文件是否存在。
4. 拒绝目录路径。

由于 SQLite 使用 `mode=rw` 而不是 `rwc`，即使路径检查出现竞态，也不会创建新的空数据库。

## 4. 国际化

国际化实现位于 [`internal/tui/i18n.go`](../internal/tui/i18n.go)。项目没有引入重量级 i18n 框架，而是使用一个强类型 `Texts` 结构保存全部文案。

优点：

- 编译期能发现字段缺失或拼写错误。
- TUI 使用 `m.texts.FieldTitle`，不依赖字符串 key 查询。
- 当前只有中文和英文，不需要复杂的复数、时区或翻译资源加载。

语言解析优先读取：

```text
LC_ALL
LC_MESSAGES
LANG
```

也可以通过 `--language auto|en|zh-CN` 显式设置。

当前改进点：

- `strings.Contains(locale, "cn")` 的判断过宽，建议解析规范 locale。
- 中文界面的浮层标题部分仍使用英文，需要决定是保留开发者工具风格还是完全本地化。
- 时间格式和星期已经本地化，但未来如果支持更多语言，应改成按 locale 配置格式。

## 5. 终端样式

样式定义在 [`internal/tui/styles.go`](../internal/tui/styles.go)。主界面参考 lazygit 的高对比终端风格；编辑和删除浮层使用独立的低饱和深色配色。

Lip Gloss 样式与文案分离，使交互逻辑不需要携带 ANSI 颜色代码。文本测量统一使用终端 cell 宽度，避免中文对齐依赖字节长度。

需要注意：TUI 不能控制用户终端字体。JetBrains Mono、Iosevka 或 Cascadia Code 必须由终端模拟器配置，程序只能控制颜色、边框、字符和字重。

## 6. 平台适配

剪贴板适配位于 [`internal/platform/clipboard.go`](../internal/platform/clipboard.go)，Linux 下按顺序尝试：

```text
wl-copy
xclip -selection clipboard
xsel --clipboard --input
```

没有可用工具时，TUI 会显示 session ID，让用户仍然可以手工复制。

当前实现已将 `platform.Copy` 从 TUI 主事件处理路径移入异步 `tea.Cmd`，并使用带两秒超时的 `exec.CommandContext` 控制外部进程。即使剪贴板工具不退出，UI 仍然可以继续处理按键和重绘。

当前调用链：

```text
handleKey
 -> copySessionID tea.Cmd
 -> platform.Copy
 -> clipboardCopiedMsg / clipboardCopyFailedMsg
 -> Update 状态栏
```

macOS 使用系统自带的 `pbcopy`，Windows 使用系统 `clip.exe`。进一步可以优先使用 OSC52，把复制内容通过终端控制序列交给终端模拟器，并保留系统命令作为 fallback。

## 7. 错误处理

当前错误分为三类：

| 层级 | 当前行为 |
| --- | --- |
| CLI/启动 | 输出到 stderr 并返回非零退出码 |
| TUI 异步操作 | 写入状态栏，部分操作保留浮层 |
| 剪贴板不可用 | 状态栏显示 ID 作为降级结果 |

TUI 已使用操作级失败消息保留 query 或 session ID，使各操作能够精确恢复 loading、busy 和弹窗状态。后续仍应把底层错误进一步映射为面向用户的处理建议，例如：

- 数据库被占用：稍后重试或退出正在写入的 OpenCode。
- 数据库版本不兼容：升级 lazyocs 或以只读模式运行。
- 剪贴板不可用：安装工具或使用显示的 session ID。
- 数据库只读：取消 `--read-only` 或修改配置。

## 8. 测试结构

当前测试文件：

| 文件 | 已覆盖内容 |
| --- | --- |
| [`config_test.go`](../internal/app/config_test.go) | 配置读取、CLI 优先级、默认配置创建、只读设置 |
| [`db_test.go`](../internal/opencode/db_test.go) | AND 搜索、统计、标题修改、递归级联删除 |
| [`model_test.go`](../internal/tui/model_test.go) | 帮助对齐、模型格式化、时间、浮层、各操作错误恢复、异步复制调度 |
| [`clipboard_unix_test.go`](../internal/platform/clipboard_unix_test.go) | 剪贴板文本传递、命令超时、工具缺失 |

当前验证命令：

```bash
go test ./...
go test -race ./...
go vet ./...
go build -o lazyocs ./cmd/lazyocs
```

普通测试、race 测试和 `go vet` 当前均通过。覆盖率约为：

| Package | 覆盖率 |
| --- | ---: |
| `internal/app` | 43.0% |
| `internal/opencode` | 52.6% |
| `internal/tui` | 20.0% |
| 总体 | 27.5% |

覆盖率数字不是目标本身，但它反映出关键缺口集中在 TUI 状态转换、正式数据库打开路径、预览查询和平台适配。

## 9. 测试进展

### 9.1 已完成：SQLite 集成测试

测试使用 `t.TempDir()` 创建文件型 SQLite 数据库，并通过正式 `Open()` 打开，已经验证：

- `mode=ro` 确实拒绝更新。
- `mode=rw` 不会创建不存在的数据库。
- `foreign_keys(ON)` 在生产连接中生效。
- 删除能够级联移除关联数据。
- OpenCode schema 缺少必要列时返回友好错误。

### 9.2 Bubble Tea 状态机测试

使用 fake Repository 向 `Model.Update` 发送按键和异步消息，验证：

- 只读模式无法进入写操作。
- 编辑和删除浮层拦截背景快捷键。
- 旧搜索结果被丢弃。
- stats 失败后可以重试。
- 删除失败后可以取消。
- 删除成功后选择索引不会越界。

### 9.3 平台适配测试

把临时 fake executable 放入测试 PATH，验证工具选择顺序、stdin 内容和退出错误。不要要求 CI 机器真的安装 Wayland 或 X11。

### 9.4 性能测试

建立可重复生成测试数据库的 fixture，记录：

- 冷启动列表耗时。
- 连续输入搜索字符的查询次数和延迟。
- 最近消息预览耗时。
- 大 session 统计耗时。
- 删除不同规模 session 树的耗时。

性能结论应来自 benchmark 和 query plan，而不是只根据数据库文件大小推断。

## 10. 代码阅读索引

| 想理解的问题 | 优先查看 |
| --- | --- |
| CLI 参数如何覆盖配置 | `main.go: visitedFlags`、`config.go: ResolveOptions` |
| 数据库如何开启只读模式 | `db.go: Open` |
| 为什么启动不扫描聊天正文 | `db.go: ListSessions` |
| 多关键词 AND 搜索如何生成 | `db.go: metadataSearchWhere` |
| 统计为什么不会启动时全部加载 | `model.go: maybeLoadStats` |
| 消息预览如何限制读取 | `db.go: RecentUserMessages`、`model.go: loadPreview` |
| 标题编辑状态如何切换 | `model.go: handleKey`、`saveTitle` |
| 删除如何递归和级联 | `model.go: deleteSession`、`db.go: DeleteSession` |
| 浮层如何覆盖主界面 | `model.go: renderDialog`、`renderOverlay` |
| 恢复 OpenCode 为什么在 TUI 退出后执行 | `app.go: Run` |
