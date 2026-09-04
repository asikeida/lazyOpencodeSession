# 近期用户消息记忆搜索

## 目标

用户通常记得自己最近问过什么，却不一定记得 session 标题或项目目录。lazyocs 因此把最近 N 天的用户消息作为临时搜索记忆，让原有 `/` 搜索同时覆盖元数据和自然语言提示。

默认窗口是 7 天，可在配置中修改：

```toml
[search]
recent_days = 7
```

允许范围为 0-365，`0` 表示关闭。

## 极简交互

没有新增页面或搜索模式。用户仍然按 `/` 输入关键词：

```text
/ cobalt migration  [memory 7d]
```

多关键词保持 AND 语义。关键词可以分布在不同来源，例如 `cobalt` 命中标题，`migration` 命中用户消息。

只有记忆参与命中时，列表才显示一行低对比度摘要：

```text
> Migration notes
  /workspace/project  09-03 21:30
  │ remember the cobalt migration issue
```

纯元数据命中继续使用原来的两行布局。

## 数据流

```text
启动 / 刷新
  -> 异步加载 session 元数据
  -> 异步加载最近 N 天用户文本
  -> 子 session 消息映射到根 session
  -> 仅保存在当前进程内存

输入搜索词
  -> 100ms debounce
  -> 本地组合匹配元数据和记忆
  -> 不再访问 SQLite
```

## 性能边界

直接连接最近 message 和 part 时，SQLite 曾选择扫描整个 part 表，在真实 3.1GB 数据库上约需 3 秒。当前查询从根 session 递归展开，再使用现有索引逐 session 读取：

```text
session_parent_idx
message_session_time_created_id_idx
part_message_id_id_idx
```

真实数据中最近 7 天约有 310 条用户文本，查询约 20-30ms。查询不读取 assistant 文本，只解析命中的 user text part。

为防止配置过大导致内存无界增长：

- 最多加载 5,000 条消息。
- 单条最多保留 4,000 个 rune。
- 查询超时为 5 秒。
- 不创建 sidecar 索引，不持久化用户消息。

## 搜索范围

元数据包括 session ID、project ID、标题、工作目录、model 和 agent。记忆包括最近 N 天的用户文本。子 session 中的用户文本归属到对应根 session，因为列表恢复的对象是根 session。

## 降级行为

- 记忆仍在加载时，搜索提示显示 `loading`，元数据搜索保持可用。
- 记忆加载失败时，提示显示 `unavailable`，不会阻止浏览和元数据搜索。
- `recent_days = 0` 时完全不执行记忆查询。
- 按 `r` 会同时刷新 session 和近期记忆。

## 与右侧预览的关系

右侧 `p` 预览仍然是独立能力，不等于近期记忆搜索的全部数据集。它现在支持通过配置文件调整显示条数：

```toml
[preview]
recent_messages_limit = 5
```

调大这个值可以让用户在详情区看到更多最近用户消息，但会增加每次预览查询和渲染成本，因此默认值仍保持紧凑的 5 条。
