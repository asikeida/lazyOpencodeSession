# 弹窗边框错位分析与修复方案

## 1. 问题现象

在中文界面中打开 `Save as`、`Delete session` 或 `?` 帮助浮层后，改变终端字体缩放比例或窗口尺寸，弹窗的左右竖边框可能在不同内容行上出现一格左右的偏移。

这个问题具有以下特征：

- 同一个弹窗的顶部和底部边框通常完整，但中间竖边可能逐行错位。
- 中文内容比纯英文更容易触发。
- 改变终端缩放比例后，问题可能出现、消失或换到其他行。
- 背景主界面的内容位置会影响浮层边框是否错位。
- 三种浮层共用 `renderOverlay`，因此都可能受到影响。

这说明问题不只是字体把线条绘制得不够平滑，而是浮层在逻辑终端 cell 层发生了位置变化。

## 2. 当前渲染路径

当前 TUI 使用 Bubble Tea 和 Lip Gloss。Bubble Tea 的 `View()` 返回一整个屏幕字符串，因此浮层通过以下步骤模拟：

```text
渲染主界面字符串
  -> 渲染弹窗字符串
  -> 按终端显示宽度切出背景左半段
  -> 插入弹窗行
  -> 拼接背景右半段
```

核心代码位于 [`internal/tui/model.go`](../internal/tui/model.go) 的 `renderOverlay`：

```go
ansi.Cut(background, 0, x) +
    modalLine +
    ansi.Cut(background, x+modalWidth, m.width)
```

其中 `x` 是弹窗左上角的逻辑列：

```go
x := (m.width - modalWidth) / 2
```

## 3. 根本原因：切点落在宽字符内部

终端坐标使用 cell，不使用 UTF-8 字节数量：

```text
ASCII 字符 a      通常占 1 个 cell
中文字符 会       通常占 2 个 cell
部分 emoji        可能占 2 个或更多 cell
组合字符          多个 rune 可能组成一个 grapheme
```

假设背景某一行在第 49 列开始显示一个占 2 格的中文字符，而弹窗左边界 `x` 为 50。切点就落在这个中文字符的两个 cell 中间。

`ansi.Cut` 为了不破坏 UTF-8 和 grapheme，不会只返回字符的一半。它会舍弃或保留完整字符。因此：

```text
期望左段宽度：50 cell
实际左段宽度：49 cell
```

随后拼接的弹窗会在这一行提前一格开始。其他背景行在同一列可能是 ASCII 或空格，仍然返回 50 格，于是竖边框看起来逐行晃动。

右侧切点也存在同样问题，可能让背景右段多一格或少一格。

改变字体缩放比例通常会使终端重新报告新的行列数，`m.width`、`modalWidth` 和 `x` 都会重新计算。新的切点可能落在另一组中文字符中间，因此问题表现会随着缩放变化。

## 4. 次要原因：Lip Gloss 隐式换行

当前 `renderDialog` 使用：

```go
m.styles.ModalBG.
    Width(innerWidth).
    Padding(0, 1).
    Render(line)
```

Lip Gloss 在设置 `Width` 后会进行自动换行。如果传入的 ANSI 文本、中文文案或快捷键提示超过内容宽度，一个逻辑 body 元素可能变成两个物理终端行，但外层只添加一次左右边框。

错误结构可能变成：

```text
│ 第一行内容
第二行内容 │
```

即使当前常见终端宽度没有触发，这也是窄终端和长本地化文案下的潜在错位来源。

## 5. 次要原因：嵌套 ANSI 样式重置

弹窗内容先分别应用文字颜色，再由外层 `ModalBG` 添加背景。内层样式结束时产生 ANSI reset，可能同时取消外层背景色，使一行中出现不连续的背景色块。

这不一定改变逻辑边框坐标，但会让边框附近出现深浅不一致的矩形，从视觉上进一步强化“框没有对齐”的感觉。

正确做法是把弹窗背景作为矩形 cell 区域的属性，而不是依赖一串嵌套 ANSI 转义序列维持。

## 6. Lazygit 的处理方式

本次参考的本机 Lazygit 源码：

```text
/home/asikeida/Downloads/git/lazygit
commit abe9e6c04197a7af3d77113f84ce7c610ef6b5a4
```

Lazygit 使用自维护的 `gocui` 和 `tcell`。弹窗本质上是一个拥有固定矩形坐标的 View：

```go
Gui.SetView(name, x0, y0, x1, y1, overlaps)
```

边框直接按 cell 坐标绘制：

```go
g.SetRune(v.x0, y, '│', ...)
g.SetRune(v.x1, y, '│', ...)
```

因此内容中的中文、颜色或 grapheme 不会改变 `x0` 和 `x1`。

最相关的 Lazygit 代码：

| 文件 | 函数 | 作用 |
| --- | --- | --- |
| `pkg/gocui/gui.go` | `SetView` | 创建或调整固定矩形 View |
| `pkg/gocui/gui.go` | `drawFrameEdges` | 按固定坐标绘制横边和竖边 |
| `pkg/gocui/gui.go` | `drawFrameCorners` | 按固定坐标绘制四个角 |
| `pkg/gocui/gui.go` | `drawTitle` | 在边框 cell 上绘制标题 |
| `pkg/gocui/view.go` | `draw` | 按 grapheme 宽度输出内容 cell |
| `pkg/gocui/view.go` | `clearRunes` | 先清理完整 View 矩形 |
| `pkg/gui/controllers/helpers/confirmation_helper.go` | `getPopupPanelDimensionsAux` | 计算弹窗尺寸和居中坐标 |
| `pkg/gui/controllers/helpers/confirmation_helper.go` | `ResizeCurrentPopupPanels` | resize 后重新计算浮层 |

Lazygit 的关键区别是：

```text
Lazygit：固定 cell 矩形 + z-order 覆盖
当前实现：ANSI 字符串切片 + 行字符串拼接
```

## 7. 选定的解决方案

项目继续保留 Bubble Tea，不迁移到 gocui。浮层合成改用 `github.com/charmbracelet/x/cellbuf`，在 Bubble Tea 返回最终字符串之前，先建立一个逻辑终端 cell buffer。

该依赖已经由 Bubble Tea/Lip Gloss 间接引入，不需要增加新的第三方技术体系。

新的渲染流程：

```text
主界面 ANSI 字符串
  -> 解析为 width × height cell buffer
  -> 对背景 cell 应用 faint
  -> 计算弹窗矩形 x/y/width/height
  -> 清理弹窗矩形
  -> 将弹窗内容写入该矩形
  -> 强制整个矩形使用统一弹窗背景色
  -> 按屏幕行重新输出 ANSI 字符串
```

cellbuf 的 `Line.Set` 对宽字符被部分覆盖做了专门处理：当覆盖发生在宽字符占位 cell 上时，会先把残缺字符对应区域清为空格，避免留下半个宽字符或造成后续 cell 偏移。

## 8. 弹窗行宽契约

除了替换 overlay 算法，还要让 `renderDialog` 满足严格契约：

```text
一个 body 元素只产生一个物理终端行
每个弹窗物理行的显示宽度都严格等于 dialogWidth
内容超过宽度时按 grapheme 截断，不允许隐式 wrap
不足宽度时显式补空格
左右边框始终位于固定 cell
```

实现方式：

1. 用 `ansi.Truncate` 截断已经带样式的内容。
2. 用 `ansi.StringWidth` 计算真实显示宽度。
3. 显式补齐内容区域空格。
4. 使用 inline 样式，关闭 Lip Gloss 的自动换行。
5. 最后添加左右边框。

## 9. 尺寸约束

当前 `dialogWidth` 在极窄终端中可能返回大于终端本身的宽度。修复后需要保证：

```text
0 <= dialogWidth <= terminalWidth
0 <= dialogHeight <= terminalHeight
```

正常终端仍保持当前约 75% 宽度；窄终端优先保证边框不越界和内容安全截断。

终端 resize 或字体缩放产生新的 `tea.WindowSizeMsg` 后，`View()` 会使用新的 `m.width/m.height` 重新计算弹窗矩形，不保留旧坐标。

## 10. 测试方案

需要增加以下回归测试：

1. 弹窗左边界恰好穿过中文双宽字符。
2. 弹窗右边界恰好穿过中文双宽字符。
3. 每一行最终显示宽度都等于终端宽度。
4. 弹窗左右竖边在所有 body 行上坐标一致。
5. 中文帮助面板不会改变边框位置。
6. 长提示文本不会在 `renderDialog` 内隐式换行。
7. 10、20、40、100 列终端下弹窗不会超过屏幕。
8. resize 前后分别基于新宽度重新居中。
9. Save as、删除和帮助仍使用同一渲染骨架。

## 11. 不采用的方案

### 11.1 只给 `ansi.Cut` 结果补空格

这个方案可以缓解当前症状，但仍然在字符串层模拟图层。以后遇到 emoji、ZWJ、组合字符和复杂 ANSI 样式时仍可能出现新的边界问题。

### 11.2 迁移到 gocui/tcell

Lazygit 的 cell View 模型非常稳定，但整个项目已经使用 Bubble Tea。为一个浮层问题更换 TUI 框架会影响状态模型、事件循环和全部渲染代码，成本和风险不合理。

### 11.3 去掉背景上下文

使用 `lipgloss.Place` 在空白全屏中居中弹窗可以绕开 overlay，但会隐藏主界面，不符合当前“浮层覆盖原应用”的交互目标。

## 12. 验收标准

修复完成后必须满足：

- 在中文和英文界面中，弹窗左右边框逐行位于同一 cell。
- 改变终端窗口尺寸后，弹窗重新居中且不残留旧边框。
- 常见字体缩放比例下不再出现整 cell 级别的边框位移。
- Save as、删除和帮助浮层保持现有视觉风格。
- `e`、`d`、`?`、`Enter`、`Esc`、`y`、`n` 的行为不变。
- 标题保存、删除 SQL 和 Repository 不发生变化。
- `go test ./...`、`go test -race ./...` 和构建全部通过。

如果修复后只剩非整数缩放下非常细微的像素级线条接缝，那属于终端字体的 box-drawing glyph 栅格化，不是 cell 坐标错位；程序层面需要消除的是截图中这种整格偏移。

## 13. 实施结果

修复前新增的测试稳定复现了三个问题：

```text
宽字符边界：20 格屏幕输出为 19 格
长中文 body：一个逻辑内容行被隐式扩成 7 个物理行
极窄终端：1 格终端计算出 12 格弹窗
```

修复已经完成：

- `renderOverlay` 已从 `ansi.Cut` 字符串拼接改为 cell buffer 矩形覆盖。
- 背景 faint 现在逐 cell 应用，不受内部 ANSI reset 影响。
- 弹窗矩形背景色逐 cell 统一设置，不再依赖嵌套样式维持。
- `renderDialog` 会显式截断、补齐并禁止隐式换行。
- 换行、回车和 Tab 在单行弹窗内容中被安全转换为空格。
- `dialogWidth` 已保证不会超过终端宽度。
- 测试覆盖了左右边界穿过中文字符、固定边框坐标和 resize 重新居中。

自动验证结果：

```text
go test ./...       通过
go test -race ./... 通过
go build            通过
```

最终仍建议在真实终端中对多个字体缩放比例做一次视觉验收，用来区分逻辑 cell 偏移与终端字体自身的像素级 box-drawing 接缝。
