# SQLite 数据访问、安全与性能

## 1. 数据层边界

数据访问代码集中在 [`internal/opencode`](../internal/opencode/)：

| 文件 | 职责 |
| --- | --- |
| [`session.go`](../internal/opencode/session.go) | 定义 session、预览、统计和查询条件 |
| [`repository.go`](../internal/opencode/repository.go) | 定义 TUI 所需的数据访问接口 |
| [`db.go`](../internal/opencode/db.go) | SQLite 连接、SQL 和 JSON 解析 |
| [`db_test.go`](../internal/opencode/db_test.go) | 查询、统计、修改和删除测试 |

Repository 接口让 TUI 只依赖能力，不依赖 SQLite：

```go
type Repository interface {
    ListSessions(ctx context.Context, filter SessionFilter) ([]Session, error)
    CountSessions(ctx context.Context, filter SessionFilter) (int, error)
    SessionStats(ctx context.Context, sessionID string) (SessionStats, error)
    UpdateSessionTitle(ctx context.Context, sessionID string, title string) error
    DeleteSession(ctx context.Context, sessionID string) error
    RecentUserMessages(ctx context.Context, sessionID string, limit int, maxChars int) ([]MessagePreview, error)
    Close() error
}
```

这使后续测试可以使用 fake Repository，也为未来支持 OpenCode API 或 sidecar 索引保留替换点。

## 2. SQLite 连接策略

`opencode.Open` 使用 URL 形式构造 DSN：

```text
file:/path/to/opencode.db?mode=ro|rw&cache=shared&_pragma=...
```

关键设置如下：

### 2.1 `mode=ro` 与 `mode=rw`

- `mode=ro` 从 SQLite 层拒绝写入。
- `mode=rw` 允许修改，但数据库不存在时不会创建新文件。
- 项目没有使用 `rwc`，避免路径写错时生成一个空数据库并误导用户。

只读模式不是单纯隐藏按键，而是同时通过 TUI 和 SQLite 两层限制写操作。

### 2.2 `foreign_keys(ON)`

SQLite 外键默认可能关闭。删除 session 时必须显式开启，否则 message、part 等关联记录可能成为孤儿数据。

```go
q.Add("_pragma", "foreign_keys(ON)")
```

外键设置是连接级别状态，因此 Repository 把最大连接数设为 1：

```go
db.SetMaxOpenConns(1)
```

除了减少 TUI 工具对 OpenCode 数据库的并发压力，它也让 pragma 行为更容易推理。

### 2.3 `busy_timeout(1000)`

OpenCode 可能正在向同一数据库写入。`busy_timeout` 允许 lazyocs 短暂等待锁释放，而不是立刻报 `SQLITE_BUSY`。

当前值为 1 秒，读取通常足够，但大会话级联删除可能需要更长等待。公开发布前建议对写操作使用约 5 秒超时，并在 UI 中给出可重试提示。

## 3. 为什么大数据库仍能快速启动

数据库总文件大小主要可能来自：

- `message.data`
- `part.data`
- 长对话产生的大量 payload

启动列表查询只读取 `session` 表的元数据：

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
where parent_id is null or parent_id = ''
order by time_updated desc
limit ?;
```

这里的核心策略是“控制查询边界”，而不是在读取所有数据后再优化 Go 代码。只要不在启动路径对 `message` 和 `part` 做聚合，数 GB 数据库并不意味着必须把数 GB 内容读入内存。

## 4. 元数据搜索

`SearchTerms` 使用空白字符分词、统一小写并去重。例如：

```text
Windows ISO windows
```

会得到：

```text
["windows", "iso"]
```

每个关键词匹配以下字段拼接后的文本：

```text
id project_id title directory model agent
```

多个关键词之间使用 SQL AND：

```sql
and lower(metadata) like '%windows%'
and lower(metadata) like '%iso%'
```

优点是语义直观、实现简单，缺点是前后通配符无法有效利用普通 B-tree 索引。当前 session 数量通常远小于 message/part 数量，因此第一优先级是增加输入 debounce，而不是立即引入复杂 FTS。

如果未来达到数万甚至更多 session，可以考虑：

1. 在内存中缓存精简元数据并搜索。
2. 建立独立 sidecar SQLite FTS5 索引。
3. 使用 OpenCode 官方 API，前提是它提供稳定搜索接口。

不建议直接在 OpenCode 数据库中创建 lazyocs 私有索引或表，这会增加对上游内部实现的侵入。

## 5. 懒加载统计

`SessionStats` 计算：

- message 数量。
- part 数量。
- message 和 part 的 `data` 字节总长度。

这类 SQL 必须访问 payload 相关表，因此默认详情配置关闭这些字段。只有用户启用字段并选中 session 时才查询，结果缓存在：

```go
map[string]opencode.SessionStats
```

`statsBusy` 用于防止同一个 session 重复发起统计请求。

这种按选中项懒加载的方式比启动时为所有 session 做 `GROUP BY` 更适合交互式工具。

## 6. 最近消息预览

`RecentUserMessages` 连接 message 和 part，但只针对一个 session：

```sql
from message m
join part p on p.message_id = m.id
where m.session_id = ?
order by m.time_created desc, p.time_created asc
limit ?;
```

SQL 使用 JSON 字符串特征做初步过滤，Go 端仍通过 `encoding/json` 解析 part，确认 `type == "text"` 后才返回文本。最终还会限制：

- 默认最多 5 条。
- 单条默认最多 500 个 rune。

这里按 rune 截断而不是按 byte 截断，避免把 UTF-8 中文截断成非法字符串。

当前字符串 `LIKE` 过滤依赖 OpenCode JSON 序列化格式。后续更稳妥的实现可以使用 SQLite JSON 函数，例如 `json_extract`，但需要先验证所有支持平台的 SQLite 驱动能力和实际查询性能。

## 7. 标题修改

标题修改使用参数化 SQL：

```sql
update session set title = ? where id = ?;
```

安全特性包括：

- 参数绑定，避免 SQL 注入。
- UI 层拒绝空标题。
- 检查 `RowsAffected`，ID 不存在时返回错误。
- 只有数据库成功后才更新 TUI 内存状态。

后续还应过滤换行和终端控制字符，避免异常粘贴内容影响 OpenCode 或 TUI 展示。

## 8. 递归删除

OpenCode session 可以通过 `parent_id` 形成父子关系。删除根 session 时，如果只删除根记录，子 session 可能残留；因此使用递归 CTE 计算整棵子树：

```sql
with recursive descendants(id) as (
  select id from session where id = ?
  union
  select s.id
  from session s
  join descendants d on s.parent_id = d.id
)
delete from session
where id in (select id from descendants);
```

同一 SQL 语句具备原子性：不会出现只删除一半 session 树的正常提交状态。message、part 和其他外键关联数据由 SQLite `ON DELETE CASCADE` 处理。

当前已经完成递归去重，仍需要补强正式连接路径的外键测试：

1. 递归 CTE 使用 `union` 去重，损坏数据库中的 parent 环也能收敛，并已有回归测试。
2. 删除测试还应通过正式 `Open()` 创建 Repository，验证 DSN 中的 `foreign_keys(ON)` 确实生效，而不是只在测试里手动开启 pragma。

## 9. 数据安全策略

正式发布应采用 fail-closed 原则：

- schema 未通过兼容检查时，只允许读取安全字段或直接拒绝启动。
- schema 写兼容检查失败时，禁止标题修改和删除。
- 删除前展示根 session、子 session 数量及关联数据范围。
- 自动备份使用 SQLite Backup API 或 `VACUUM INTO`，不能在 WAL 模式下只复制主 `.db` 文件。
- 不向 OpenCode 数据库增加私有“软删除”字段；回收站应放在独立 sidecar 数据库或导出文件中。

## 10. Schema 漂移风险

lazyocs 依赖的是 OpenCode 内部 SQLite schema，而不是稳定公开协议。这是项目最大的长期风险。

建议 `Open()` 后执行：

```sql
pragma table_info(session);
pragma foreign_key_list(message);
pragma foreign_key_list(part);
```

并按能力划分兼容等级：

| 等级 | 条件 | 行为 |
| --- | --- | --- |
| Read compatible | 列表所需字段存在 | 允许浏览和搜索 |
| Preview compatible | message/part 关系和 JSON 数据可识别 | 允许预览 |
| Write compatible | 标题字段、parent 关系和级联外键符合预期 | 允许修改和删除 |

这样即使 OpenCode 升级，工具也能给出明确错误，而不是把底层的 `no such column` 直接暴露给用户。

## 11. 性能优化顺序

推荐按投入产出排序：

1. 搜索增加 80-150ms debounce。
2. 使用操作类型消息取消或忽略过期查询。
3. 为启动、搜索、预览和统计建立 benchmark。
4. 目录存在检查改为只处理可见项或选中项。
5. session 数量真正达到瓶颈后，再考虑 sidecar FTS 和分页。

不建议因为数据库文件大就提前更换 SQLite 驱动、引入缓存服务或重写为 Rust。先通过 query plan 和 benchmark 定位瓶颈，再决定优化手段。
