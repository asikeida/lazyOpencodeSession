# lazyOpencodeSession MVP 设计说明

## 1. 项目定位

`lazyOpencodeSession` 是一个面向 OpenCode CLI 用户的终端会话管理工具，目标是提供类似 `lazygit`、`lazydocker` 的交互体验，让用户可以在终端中快速浏览、搜索、预览并恢复 OpenCode 历史会话。

一句话定位：

```text
A fast terminal UI for browsing, searching, and resuming OpenCode sessions.
```

建议项目名和命令名分开：

```text
项目名：lazyOpencodeSession
命令名：lazyocs
```

原因是 `lazyOpencodeSession` 语义明确，适合作为仓库名；但日常命令太长，`lazyocs` 更适合作为 CLI 命令，输入成本低，也更接近 `lazygit`、`lazydocker` 的使用感。

## 2. 要解决的问题

OpenCode 已经有基础会话能力，例如：

```bash
opencode session list
opencode --session <session_id>
```

但在真实长期使用后，会出现几个问题：

- 会话越来越多，纯表格列表很难找。
- 只能看到标题和更新时间，难以确认是不是目标会话。
- 搜索标题、项目路径、会话内容不够方便。
- 找到 session id 后还要复制并手动执行恢复命令。
- OpenCode 数据库可能很大，第三方 TUI 如果直接扫描消息正文会卡顿。
- 用户希望在终端中完成完整流程，不想切到浏览器或桌面应用。

本项目的核心价值是：

```text
用最快的方式从大量 OpenCode 历史会话中找回目标会话，并一键恢复。
```

## 3. 技术栈选择

MVP 推荐技术栈：

```text
Go + Bubble Tea + Lip Gloss + SQLite
```

具体依赖建议：

```text
github.com/charmbracelet/bubbletea
github.com/charmbracelet/lipgloss
github.com/charmbracelet/bubbles/list
github.com/mattn/go-sqlite3 或 modernc.org/sqlite
github.com/spf13/cobra
github.com/spf13/viper
```

优先建议使用 `modernc.org/sqlite`，原因是它是纯 Go 实现，跨平台分发更简单，不需要 CGO。但如果后续遇到兼容性或性能问题，可以切换到 `github.com/mattn/go-sqlite3`。

### 3.1 为什么选择 Go

- 可以编译成单个二进制文件，安装和分发简单。
- 启动速度快，适合做 CLI/TUI 工具。
- SQLite 支持成熟。
- 代码复杂度低于 Rust，开发效率更高。
- Bubble Tea 生态成熟，适合做 lazygit 风格的多面板 TUI。
- 比 TypeScript/Bun 更不依赖运行时环境。

### 3.2 为什么不优先选择 TypeScript

TypeScript 更贴近 OpenCode 生态，但不适合作为这个项目的第一选择：

- 用户需要安装 Bun 或 Node。
- 全局安装和运行时依赖更重。
- 面对大 SQLite 数据库时，性能和内存使用更难控制。
- 终端 TUI 的跨平台分发不如 Go 单二进制直接。

### 3.3 为什么不优先选择 Rust

Rust + Ratatui 性能很好，但 MVP 阶段不推荐作为第一选择：

- 开发周期更长。
- 类型和生命周期成本更高。
- 对项目早期快速试错不如 Go 友好。

如果未来项目规模扩大、性能瓶颈明确，再考虑 Rust 重写核心索引层或直接迁移也可以。

## 4. MVP 功能范围

MVP 的关键目标是：

```text
启动快，查找快，恢复快。
```

第一版不要追求功能大而全，应该先把最核心路径做到顺滑。

### 4.1 必须有的功能

#### 4.1.1 自动发现 OpenCode 数据库

默认读取：

```text
~/.local/share/opencode/opencode.db
```

支持通过命令参数指定数据库：

```bash
lazyocs --db /path/to/opencode.db
```

支持通过环境变量指定：

```bash
LAZYOCS_DB=/path/to/opencode.db lazyocs
```

如果找不到数据库，TUI 应展示清晰错误提示，而不是直接崩溃。

#### 4.1.2 会话列表

左侧展示 OpenCode sessions，默认按更新时间倒序排列。

每条 session 展示：

```text
标题
项目路径
更新时间
session id 短 ID
```

建议列表样式：

```text
Sessions
────────────────────────────────────────
> opencode 桌面版本与会话管理
  /home/asikeida                         11:03
  ses_f9ff94124ffe...

  Arch KDE Plasma LocalSend 托盘图标空白
  /home/asikeida                         08:48
  ses_fbe4d94efffe...
```

#### 4.1.3 搜索会话元数据

按 `/` 进入搜索模式。

搜索范围包括：

- session title
- session id
- directory
- project id
- model
- agent

MVP 阶段不默认搜索聊天正文，避免卡顿。

搜索应该实时过滤，但要做 debounce，例如 80ms 到 150ms，避免每输入一个字符都同步阻塞 UI。

#### 4.1.4 右侧详情面板

选中左侧 session 后，右侧显示详情。

内容包括：

```text
Title:      opencode 桌面版本与会话管理
Session:    ses_f9ff94124ffeo6VURxgSYjZxGL
Project:    global
Directory:  /home/asikeida
Updated:    2026-09-02 11:03
Created:    2026-09-02 10:50
Model:      openai/gpt-5.5
Agent:      build
Cost:       $0.0123
Tokens:     input 12345 / output 6789 / cache 9999
```

右侧详情只读取 `session` 表，不读取 `part.data`。

#### 4.1.5 最近用户消息预览

MVP 可以支持一个轻量预览，但必须懒加载。

建议行为：

- 默认只显示 session metadata。
- 按 `p` 加载预览。
- 只读取最近 3 到 5 条用户消息。
- 每条消息截断到 300 到 500 字符。
- 如果 session 很大，显示提示：`Large session, preview limited`。

预览示例：

```text
Recent User Messages
────────────────────────────────────────
1. opencode 推出了桌面版本吗？他们的有关回话管理的是怎么做的

2. 我现在使用的是opencode cli，但是我有些在项目中的回话找不到了...

3. 那我想做一个这个tui，你觉得怎么样，对标lazydocker，lazygit
```

#### 4.1.6 回车恢复会话

按 `Enter` 执行：

```bash
opencode --session <session_id>
```

建议提供两种模式：

```text
默认模式：退出 lazyocs，然后在当前终端进入 opencode session
新终端模式：后续支持，在 Kitty/WezTerm/tmux 中新 tab 或 split 打开
```

MVP 先做默认模式即可。

#### 4.1.7 复制 session id

按 `y` 复制当前 session id。

Linux 下可以支持：

```text
wl-copy
xclip
xsel
```

如果没有剪贴板工具，则在底部提示：

```text
Clipboard tool not found. Session ID: ses_xxx
```

#### 4.1.8 帮助面板

按 `?` 显示快捷键。

MVP 快捷键建议：

```text
q / Ctrl+C   退出
↑ / k        上移
↓ / j        下移
/            搜索
Esc          清空搜索或关闭弹窗
Enter        恢复选中会话
p            加载最近消息预览
y            复制 session id
r            刷新列表
?            帮助
```

### 4.2 MVP 暂不做的功能

为了保证第一版速度和稳定性，以下功能建议暂缓：

- 全文搜索所有聊天内容。
- 自动生成 AI 摘要。
- 删除 session。
- 重命名 session。
- fork session。
- diff 详细预览。
- token 成本复杂图表。
- 多数据库合并。
- 远程同步。
- Web UI。

这些功能都可以做，但不应该进入第一版核心路径。

## 5. TUI 效果设计

整体风格对标 `lazygit` 和 `lazydocker`：信息密度高、键盘优先、状态清晰、响应快速。

### 5.1 默认布局

建议使用三栏或两栏布局。

MVP 推荐两栏布局：

```text
┌─ lazyOpencodeSession ─ Sessions ─────────────┐┌─ Details ───────────────────────────────┐
│ / search: opencode                           ││ Title: opencode 桌面版本与会话管理        │
│                                              ││ Session: ses_f9ff94124ffeo6VURxgSYjZxGL │
│ > opencode 桌面版本与会话管理                  ││ Directory: /home/asikeida                │
│   /home/asikeida                     11:03   ││ Updated: 2026-09-02 11:03               │
│                                              ││ Model: openai/gpt-5.5                   │
│   Arch KDE Plasma LocalSend 托盘图标空白      ││ Cost: $0.012                             │
│   /home/asikeida                     08:48   ││ Tokens: 12.3k in / 6.7k out             │
│                                              ││                                          │
│   Linux 一键重启到 Windows                    ││ Press p to preview recent user messages │
│   /home/asikeida                     08:39   ││ Press Enter to resume this session      │
└──────────────────────────────────────────────┘└──────────────────────────────────────────┘
  383 sessions | / search | Enter resume | p preview | y copy id | ? help | q quit
```

布局比例建议：

```text
左侧 sessions：45%
右侧 details：55%
底部 status bar：1 行
```

在窄屏终端中自动切换为单栏：

```text
宽度 >= 110：双栏
宽度 < 110：单栏，只显示列表，按 Enter 或 Tab 看详情
```

### 5.2 视觉风格

视觉应该干净、克制、实用，不要过度花哨。

建议默认主题：

```text
背景：终端默认背景
普通文本：浅灰
弱文本：暗灰
高亮项：蓝色或青色
边框：灰色
警告：黄色
危险：红色
成功：绿色
```

选中行应该明显，但不能刺眼：

```text
> Title                         11:03
  /path/to/project
```

如果终端支持真彩色，可以使用更细腻的颜色；如果不支持，自动退化到 8 色或 16 色。

### 5.3 信息优先级

列表中最重要的是：

```text
标题 > 项目路径 > 更新时间 > session id
```

右侧详情中最重要的是：

```text
标题 > 路径 > 更新时间 > 模型/成本 > session id
```

session id 很长，不应该在列表中抢占主要视觉空间，只显示短 ID 即可。

### 5.4 搜索体验

按 `/` 后底部状态栏切换为搜索输入：

```text
Search: opencode_
```

搜索结果实时刷新。

匹配项可以高亮，但 MVP 可以先不做复杂高亮，只要过滤速度够快。

没有结果时显示：

```text
No sessions matched "xxx"
```

### 5.5 空状态和错误状态

找不到数据库：

```text
OpenCode database not found

Checked:
  ~/.local/share/opencode/opencode.db

Run with:
  lazyocs --db /path/to/opencode.db
```

数据库无法读取：

```text
Failed to read OpenCode database

Path: /home/user/.local/share/opencode/opencode.db
Error: database is locked

Try:
  lazyocs --readonly
```

没有 session：

```text
No OpenCode sessions found.
```

## 6. 架构设计

架构目标：

```text
UI、业务逻辑、数据库访问、外部命令执行必须分层，避免 TUI 代码直接写 SQL 或直接拼 shell。
```

推荐目录结构：

```text
lazyOpencodeSession/
├── cmd/
│   └── lazyocs/
│       └── main.go
├── internal/
│   ├── app/
│   │   ├── app.go
│   │   ├── config.go
│   │   └── errors.go
│   ├── opencode/
│   │   ├── db.go
│   │   ├── repository.go
│   │   ├── session.go
│   │   ├── message.go
│   │   └── schema.go
│   ├── service/
│   │   ├── session_service.go
│   │   ├── search_service.go
│   │   └── resume_service.go
│   ├── tui/
│   │   ├── model.go
│   │   ├── update.go
│   │   ├── view.go
│   │   ├── keymap.go
│   │   ├── styles.go
│   │   ├── layout.go
│   │   └── components/
│   │       ├── session_list.go
│   │       ├── details_panel.go
│   │       ├── preview_panel.go
│   │       ├── help_modal.go
│   │       └── status_bar.go
│   ├── theme/
│   │   ├── theme.go
│   │   ├── builtin.go
│   │   └── loader.go
│   ├── clipboard/
│   │   └── clipboard.go
│   └── platform/
│       ├── path.go
│       └── terminal.go
├── docs/
├── opcode-summary/
├── go.mod
├── go.sum
├── README.md
└── LICENSE
```

### 6.1 分层说明

#### 6.1.1 cmd 层

职责：

- 解析命令行参数。
- 初始化配置。
- 启动 TUI。
- 返回退出码。

不应该做：

- 不直接查询数据库。
- 不写 TUI 业务逻辑。
- 不处理复杂搜索逻辑。

#### 6.1.2 app 层

职责：

- 组装依赖。
- 管理应用生命周期。
- 读取配置。
- 建立数据库连接。

`app` 层相当于 glue code，避免 `main.go` 变成大杂烩。

#### 6.1.3 opencode 层

职责：

- 封装 OpenCode SQLite schema。
- 提供 repository 接口。
- 把数据库行转换为领域对象。

示例接口：

```go
type SessionRepository interface {
    ListSessions(ctx context.Context, filter SessionFilter) ([]Session, error)
    GetSession(ctx context.Context, id string) (Session, error)
    ListRecentUserMessages(ctx context.Context, sessionID string, limit int) ([]MessagePreview, error)
}
```

UI 层只能调用接口，不能直接写 SQL。

#### 6.1.4 service 层

职责：

- 组织业务流程。
- 搜索排序。
- 恢复会话。
- 判断大 session 是否需要限制预览。

示例：

```go
type SessionService struct {
    repo SessionRepository
}

func (s *SessionService) Search(ctx context.Context, query string) ([]Session, error)
func (s *SessionService) Preview(ctx context.Context, sessionID string) (SessionPreview, error)
```

#### 6.1.5 tui 层

职责：

- Bubble Tea model/update/view。
- 管理焦点、键盘事件、弹窗状态。
- 调用 service 层获取数据。
- 把数据渲染成终端 UI。

不应该做：

- 不直接写 SQL。
- 不直接拼 `opencode --session` 命令。
- 不直接读取配置文件。

#### 6.1.6 theme 层

职责：

- 内置主题。
- 加载用户主题。
- 把主题映射为 Lip Gloss 样式。

#### 6.1.7 platform 层

职责：

- 跨平台路径处理。
- 检查外部命令是否存在。
- 处理剪贴板工具。
- 后续处理 Kitty/WezTerm/tmux 打开新 tab 或 split。

### 6.2 数据模型

MVP 可以先定义精简模型：

```go
type Session struct {
    ID                string
    ProjectID         string
    ParentID          *string
    Title             string
    Directory         string
    CreatedAt         time.Time
    UpdatedAt         time.Time
    Model             string
    Agent             string
    Cost              float64
    TokensInput       int64
    TokensOutput      int64
    TokensReasoning   int64
    TokensCacheRead   int64
    TokensCacheWrite  int64
}
```

消息预览模型：

```go
type MessagePreview struct {
    ID        string
    Role      string
    Text      string
    CreatedAt time.Time
}
```

搜索过滤：

```go
type SessionFilter struct {
    Query     string
    ProjectID string
    Directory string
    Limit     int
    Offset    int
    SortBy    string
    Desc      bool
}
```

### 6.3 数据库读取策略

启动时只执行轻量 SQL：

```sql
select
  id,
  project_id,
  parent_id,
  title,
  directory,
  time_created,
  time_updated,
  model,
  agent,
  cost,
  tokens_input,
  tokens_output,
  tokens_reasoning,
  tokens_cache_read,
  tokens_cache_write
from session
where parent_id is null
order by time_updated desc
limit ? offset ?;
```

搜索元数据时：

```sql
select ...
from session
where
  title like ?
  or directory like ?
  or id like ?
  or project_id like ?
order by time_updated desc
limit ?;
```

最近消息预览按需加载：

```sql
select p.id, p.data, p.time_created
from part p
join message m on m.id = p.message_id
where p.session_id = ?
order by p.time_created desc
limit ?;
```

注意：OpenCode 的 `part.data` 是 JSON 字符串，里面可能很大。MVP 应该限制读取数量，并对文本长度做截断。

### 6.4 性能原则

这是项目成败的关键。

必须遵守：

- 启动时不扫描 `part.data`。
- 启动时不扫描 `event.data`。
- 列表只读 `session` 表。
- 预览按需加载。
- 全文搜索不进 MVP。
- 大字段读取必须限制数量和长度。
- UI 线程不能直接执行慢 SQL，使用 Bubble Tea command 异步加载。
- 每次搜索结果限制数量，例如默认 200 条。
- 搜索输入做 debounce。

性能目标：

```text
数据库 3GB、sessions 500、parts 50000 时：
启动到可操作 < 500ms
元数据搜索 < 100ms
切换选中项无明显卡顿
预览最近消息 < 300ms
```

如果未来加入全文搜索，应使用单独索引：

```text
SQLite FTS5 或独立 bleve index
```

并且必须后台增量构建，不能启动时重建。

## 7. 配置设计

配置文件建议路径：

```text
~/.config/lazyocs/config.toml
```

MVP 配置示例：

```toml
db = "~/.local/share/opencode/opencode.db"
command = "opencode"
theme = "default"
limit = 300

[preview]
enabled = false
max_messages = 5
max_chars_per_message = 500

[keymap]
quit = ["q", "ctrl+c"]
search = ["/"]
resume = ["enter"]
preview = ["p"]
copy_id = ["y"]
refresh = ["r"]
help = ["?"]
```

命令行参数优先级应高于配置文件：

```text
CLI 参数 > 环境变量 > 配置文件 > 默认值
```

## 8. 主题支持

MVP 应支持修改主题，但不要把主题系统做复杂。

建议内置主题：

```text
default
dark
light
tokyo-night
catppuccin
gruvbox
```

MVP 至少实现：

```text
default
```

主题配置示例：

```toml
theme = "default"

[themes.default]
foreground = "#d0d0d0"
muted = "#777777"
accent = "#7aa2f7"
border = "#555555"
selected_fg = "#ffffff"
selected_bg = "#264f78"
warning = "#e0af68"
danger = "#f7768e"
success = "#9ece6a"
```

实现上不要让业务代码接触颜色字符串。应该由 theme 层输出样式对象：

```go
type Styles struct {
    Base      lipgloss.Style
    Muted     lipgloss.Style
    Accent    lipgloss.Style
    Border    lipgloss.Style
    Selected  lipgloss.Style
    Warning   lipgloss.Style
    Danger    lipgloss.Style
    Success   lipgloss.Style
}
```

这样以后更换主题不会影响 TUI 业务逻辑。

## 9. 可维护性设计

这个项目容易失控的地方是：TUI 状态、数据库查询、外部命令、快捷键逻辑全部混在一起。

要保持可维护，需要遵守以下规则。

### 9.1 UI 状态集中管理

Bubble Tea 的 `Model` 中只保存 UI 状态和必要数据：

```go
type Model struct {
    width     int
    height    int
    sessions  []Session
    selected  int
    query     string
    mode      Mode
    preview   *SessionPreview
    loading   bool
    err       error
}
```

不要在 `Model` 中保存数据库连接细节。

### 9.2 Update 只做状态转移

`Update` 函数负责：

- 响应按键。
- 切换模式。
- 发起异步 command。
- 接收加载结果。

不要在 `Update` 中写复杂 SQL。

### 9.3 View 只做渲染

`View` 函数负责根据当前 model 渲染字符串。

不要在 `View` 中发起 IO。

不要在 `View` 中修改状态。

### 9.4 Repository 可测试

数据库层应该用接口封装，方便用临时 SQLite 数据库做测试。

测试重点：

- session 列表排序正确。
- 搜索过滤正确。
- 时间戳转换正确。
- 缺字段或空值时不崩溃。
- 大 `part.data` 被截断。

### 9.5 OpenCode schema 变化隔离

OpenCode 数据库 schema 未来可能变化。应该把 schema 相关 SQL 集中放在 `internal/opencode`，不要散落到 UI 层。

启动时可以做简单 schema 检查：

```sql
pragma table_info(session);
```

如果缺少关键字段，提示用户当前 OpenCode 版本不兼容。

## 10. CLI 行为设计

MVP 主命令：

```bash
lazyocs
```

常用参数：

```bash
lazyocs --db ~/.local/share/opencode/opencode.db
lazyocs --limit 500
lazyocs --theme dark
lazyocs --no-preview
lazyocs --version
lazyocs --help
```

后续可以添加非交互命令：

```bash
lazyocs list
lazyocs search "keyword"
lazyocs resume <session_id>
lazyocs preview <session_id>
lazyocs config path
```

但第一版可以先只做 TUI。

## 11. 恢复会话的实现方式

MVP 的恢复逻辑：

1. 用户选中 session。
2. 按 `Enter`。
3. TUI 退出 raw mode。
4. 当前进程执行或启动：

```bash
opencode --session <session_id>
```

Go 中有两种实现方式。

方案 A：退出 TUI 后用 `exec.Command` 启动 opencode，并把 stdio 接到当前终端。

```go
cmd := exec.Command("opencode", "--session", sessionID)
cmd.Stdin = os.Stdin
cmd.Stdout = os.Stdout
cmd.Stderr = os.Stderr
return cmd.Run()
```

方案 B：只输出命令，让 shell 执行。这个体验较差，不推荐。

MVP 建议使用方案 A。

如果用户配置了自定义命令：

```toml
command = "opencode"
```

恢复时使用该命令。

## 12. 安全和数据保护

MVP 默认只读数据库，不做任何写操作。

必须坚持：

```text
默认 readonly。
```

数据库连接建议使用只读模式：

```text
file:/home/user/.local/share/opencode/opencode.db?mode=ro
```

这样即使程序有 bug，也不会破坏 OpenCode 数据。

删除、重命名、清理这些写操作后续再做，并且必须有确认和备份。

## 13. 后续路线图

### 13.1 v0.1 MVP

- 自动发现数据库。
- 快速 session 列表。
- 元数据搜索。
- 右侧详情面板。
- 最近用户消息懒加载预览。
- 回车恢复 session。
- 复制 session id。
- 基础主题。
- 帮助面板。

### 13.2 v0.2 搜索增强

- 可选全文搜索。
- SQLite FTS5 索引。
- 后台增量索引。
- 搜索结果定位到消息。
- 支持按项目、模型、时间过滤。

### 13.3 v0.3 会话管理

- 重命名 session。
- 删除 session。
- 导出 session。
- fork session。
- 打开项目目录。
- 显示 session diff 概览。

### 13.4 v0.4 lazygit 风格增强

- 命令面板。
- 多面板焦点切换。
- 自定义 keymap。
- 多主题。
- tmux/Kitty/WezTerm 新窗口恢复。
- session 分组和收藏。

### 13.5 v1.0 稳定版

- 跨平台二进制发布。
- 完整 README 和安装脚本。
- CI 测试。
- 性能 benchmark。
- OpenCode schema 兼容性检测。
- 默认只读安全策略。

## 14. 验收标准

MVP 完成后，应该能满足以下标准：

- 在 3GB OpenCode 数据库上启动不明显卡顿。
- 可以看到所有 root sessions。
- 可以用 `/` 搜索标题、路径、session id。
- 可以选中 session 后按 `Enter` 直接恢复。
- 可以按 `p` 查看最近几条用户消息。
- 可以按 `y` 复制 session id。
- 不会修改 OpenCode 数据库。
- 找不到数据库时有明确提示。
- 窄屏终端不会完全不可用。
- 代码分层清楚，数据库、服务、TUI 不互相污染。

## 15. 最小开发计划

建议按以下顺序开发：

1. 初始化 Go 项目和 CLI 入口。
2. 实现配置读取和数据库路径解析。
3. 实现 SQLite 只读连接。
4. 实现 `SessionRepository.ListSessions`。
5. 做一个最简单 Bubble Tea 列表。
6. 加右侧详情面板。
7. 加 `/` 搜索。
8. 加 `Enter` 恢复 session。
9. 加 `p` 懒加载消息预览。
10. 加主题和帮助面板。
11. 用真实 3GB 数据库做性能测试。

## 16. 最重要的设计取舍

这个项目第一版最重要的取舍是：

```text
宁可功能少，也不能卡。
```

如果一个功能会导致启动慢，例如全文搜索、token 复杂统计、完整聊天预览，就不要默认开启。OpenCode 用户使用这个工具的第一动机是快速找回会话，因此速度是最高优先级。

最终理想体验应该是：

```text
打开 lazyocs，输入几个关键词，选中目标会话，回车，立即回到 OpenCode。
```
