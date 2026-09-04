package tui

import (
	"os"
	"strings"
)

type HelpLine struct {
	Key         string
	Description string
}

type Texts struct {
	AppTitle              string
	SessionsTitle         string
	DetailsTitle          string
	LoadingSessions       string
	Loading               string
	NoSessions            string
	SelectSession         string
	SearchSessions        string
	ReloadingSessions     string
	LoadingPreview        string
	LoadedPreviewMessages string
	SavingTitle           string
	EditingTitle          string
	TitleSaved            string
	TitleCancelled        string
	TitleReadOnly         string
	TitleEmpty            string
	ConfirmDelete         string
	DeletingSession       string
	LoadingDeleteImpact   string
	SessionDeleted        string
	DeleteCancelled       string
	SaveAsTitle           string
	DeleteDialogTitle     string
	DeleteWarning         string
	DeleteTarget          string
	DeleteSessions        string
	DeleteMessages        string
	DeleteParts           string
	SaveAction            string
	DeleteAction          string
	CancelAction          string
	StatusSessions        string
	StatusMatched         string
	ClipboardUnavailable  string
	CopyingSessionID      string
	CopiedSessionID       string
	MemoryWindow          string
	MemoryLoading         string
	MemoryUnavailable     string
	DetailsScrollHint     string
	RecentUserMessages    string
	NoTextPreview         string
	ReadOnlyNotice        string
	FooterBrowse          string
	FooterSearch          string
	FooterHelp            string
	ActionHint            string
	HelpTitle             string
	HelpLines             []HelpLine
	FieldTitle            string
	FieldSession          string
	FieldProject          string
	FieldDirectory        string
	FieldPathStatus       string
	FieldMessageCount     string
	FieldPartCount        string
	FieldSize             string
	FieldLargeSession     string
	PathExists            string
	PathMissing           string
	Yes                   string
	No                    string
	FieldUpdated          string
	FieldCreated          string
	FieldModel            string
	FieldAgent            string
	FieldCost             string
	FieldTokens           string
	FieldResumeCommand    string
	TokensFormat          string
	Weekdays              [7]string
}

func ResolveLanguage(lang string) string {
	lang = strings.TrimSpace(lang)
	if lang == "" || lang == "auto" {
		locale := strings.ToLower(os.Getenv("LC_ALL") + " " + os.Getenv("LC_MESSAGES") + " " + os.Getenv("LANG"))
		if strings.Contains(locale, "zh") || strings.Contains(locale, "cn") {
			return "zh-CN"
		}
		return "en"
	}
	if strings.EqualFold(lang, "zh") || strings.EqualFold(lang, "zh-cn") || strings.EqualFold(lang, "zh_CN") {
		return "zh-CN"
	}
	return "en"
}

func NewTexts(lang string) Texts {
	if ResolveLanguage(lang) == "zh-CN" {
		return Texts{
			AppTitle:              "lazyOpencodeSession",
			SessionsTitle:         "会话列表",
			DetailsTitle:          "会话详情",
			LoadingSessions:       "正在加载会话...",
			Loading:               "正在加载...",
			NoSessions:            "没有找到会话",
			SelectSession:         "请选择一个会话",
			SearchSessions:        "搜索会话",
			ReloadingSessions:     "正在重新加载会话...",
			LoadingPreview:        "正在加载预览...",
			LoadedPreviewMessages: "已加载 %d 条预览消息",
			SavingTitle:           "正在保存标题...",
			EditingTitle:          "编辑标题：Enter 保存，Esc 取消",
			TitleSaved:            "标题已保存",
			TitleCancelled:        "已取消标题编辑",
			TitleReadOnly:         "当前为只读模式，无法修改会话，请取消 --read-only 或修改 read_only 配置",
			TitleEmpty:            "标题不能为空",
			ConfirmDelete:         "确认删除当前会话？按 y/Enter 确认，n/Esc 取消",
			DeletingSession:       "正在删除会话...",
			LoadingDeleteImpact:   "正在统计删除影响...",
			SessionDeleted:        "会话已删除",
			DeleteCancelled:       "已取消删除",
			SaveAsTitle:           "Save as",
			DeleteDialogTitle:     "Delete session",
			DeleteWarning:         "此操作将永久删除该会话及其所有子会话。",
			DeleteTarget:          "目标",
			DeleteSessions:        "会话",
			DeleteMessages:        "消息",
			DeleteParts:           "内容块",
			SaveAction:            "保存",
			DeleteAction:          "删除",
			CancelAction:          "取消",
			StatusSessions:        "%d 个会话",
			StatusMatched:         "%d/%d 匹配",
			ClipboardUnavailable:  "剪贴板不可用：%s",
			CopyingSessionID:      "正在复制 session id...",
			CopiedSessionID:       "已复制 session id",
			MemoryWindow:          "记忆 %d天",
			MemoryLoading:         "加载中",
			MemoryUnavailable:     "不可用",
			DetailsScrollHint:     "h/l 切换面板  j/k 滚动",
			RecentUserMessages:    "最近用户消息",
			NoTextPreview:         "没有找到文本预览",
			ReadOnlyNotice:        "默认允许修改标题；使用 --read-only 可进入只读模式。",
			FooterBrowse:          "浏览",
			FooterSearch:          "搜索",
			FooterHelp:            "/ 搜索 | Enter 恢复 | p/Ctrl-P 预览 | y 复制 ID | ? 帮助 | q 退出",
			ActionHint:            "e 编辑标题 | Enter 恢复 | p 预览 | y 复制 ID",
			HelpTitle:             "lazyOpencodeSession 帮助",
			HelpLines: []HelpLine{
				{Key: "q / Ctrl+C", Description: "退出"},
				{Key: "↑/k ↓/j", Description: "移动选择"},
				{Key: "搜索中 Ctrl-K/J", Description: "上下移动，j/k 仍输入文字"},
				{Key: "搜索中 Ctrl-P", Description: "预览当前会话"},
				{Key: "PageUp/Down", Description: "翻页"},
				{Key: "/", Description: "搜索元数据和近期用户消息"},
				{Key: "Esc", Description: "清空搜索 / 关闭帮助"},
				{Key: "Enter", Description: "恢复选中会话"},
				{Key: "e", Description: "编辑当前会话标题"},
				{Key: "d", Description: "删除当前会话（需要确认）"},
				{Key: "p", Description: "预览最近用户消息"},
				{Key: "y", Description: "复制 session id"},
				{Key: "r", Description: "重新加载会话"},
				{Key: "?", Description: "切换帮助"},
			},
			FieldTitle:         "标题",
			FieldSession:       "会话",
			FieldProject:       "项目",
			FieldDirectory:     "目录",
			FieldPathStatus:    "路径状态",
			FieldMessageCount:  "消息数",
			FieldPartCount:     "片段数",
			FieldSize:          "大小",
			FieldLargeSession:  "大型会话",
			PathExists:         "存在",
			PathMissing:        "不存在",
			Yes:                "是",
			No:                 "否",
			FieldUpdated:       "更新",
			FieldCreated:       "创建",
			FieldModel:         "模型",
			FieldAgent:         "代理",
			FieldCost:          "费用",
			FieldTokens:        "Token",
			FieldResumeCommand: "恢复命令",
			TokensFormat:       "输入 %d / 输出 %d / 推理 %d / 缓存 %d",
			Weekdays:           [7]string{"周日", "周一", "周二", "周三", "周四", "周五", "周六"},
		}
	}

	return Texts{
		AppTitle:              "lazyOpencodeSession",
		SessionsTitle:         "Sessions",
		DetailsTitle:          "Details",
		LoadingSessions:       "loading sessions...",
		Loading:               "loading...",
		NoSessions:            "No sessions found",
		SelectSession:         "Select a session",
		SearchSessions:        "search sessions",
		ReloadingSessions:     "reloading sessions...",
		LoadingPreview:        "loading preview...",
		LoadedPreviewMessages: "loaded %d preview messages",
		SavingTitle:           "saving title...",
		EditingTitle:          "editing title: Enter save, Esc cancel",
		TitleSaved:            "title saved",
		TitleCancelled:        "title edit cancelled",
		TitleReadOnly:         "read-only mode; restart without --read-only to modify sessions",
		TitleEmpty:            "title cannot be empty",
		ConfirmDelete:         "Delete this session? Press y/Enter to confirm, n/Esc to cancel",
		DeletingSession:       "deleting session...",
		LoadingDeleteImpact:   "calculating delete impact...",
		SessionDeleted:        "session deleted",
		DeleteCancelled:       "delete cancelled",
		SaveAsTitle:           "Save as",
		DeleteDialogTitle:     "Delete session",
		DeleteWarning:         "This permanently deletes the session and all child sessions.",
		DeleteTarget:          "Target",
		DeleteSessions:        "Sessions",
		DeleteMessages:        "Messages",
		DeleteParts:           "Parts",
		SaveAction:            "save",
		DeleteAction:          "delete",
		CancelAction:          "cancel",
		StatusSessions:        "%d sessions",
		StatusMatched:         "%d/%d matched",
		ClipboardUnavailable:  "clipboard unavailable: %s",
		CopyingSessionID:      "copying session id...",
		CopiedSessionID:       "copied session id",
		MemoryWindow:          "memory %dd",
		MemoryLoading:         "loading",
		MemoryUnavailable:     "unavailable",
		DetailsScrollHint:     "h/l switch  j/k scroll",
		RecentUserMessages:    "Recent User Messages",
		NoTextPreview:         "No text preview found",
		ReadOnlyNotice:        "Title editing is enabled by default; use --read-only to prevent writes.",
		FooterBrowse:          "browse",
		FooterSearch:          "search",
		FooterHelp:            "/ search | Enter resume | p/Ctrl-P preview | y copy id | ? help | q quit",
		ActionHint:            "e edit title | Enter resume | p preview | y copy id",
		HelpTitle:             "lazyOpencodeSession Help",
		HelpLines: []HelpLine{
			{Key: "q / Ctrl+C", Description: "Quit"},
			{Key: "↑/k ↓/j", Description: "Move selection"},
			{Key: "Search Ctrl-K/J", Description: "Move; j/k still type text"},
			{Key: "Search Ctrl-P", Description: "Preview current session"},
			{Key: "PageUp/Down", Description: "Jump list"},
			{Key: "/", Description: "Search metadata and recent user messages"},
			{Key: "Esc", Description: "Clear search / close help"},
			{Key: "Enter", Description: "Resume selected session"},
			{Key: "e", Description: "Edit current session title"},
			{Key: "d", Description: "Delete current session (confirmation required)"},
			{Key: "p", Description: "Preview recent user messages"},
			{Key: "y", Description: "Copy session id"},
			{Key: "r", Description: "Reload sessions"},
			{Key: "?", Description: "Toggle help"},
		},
		FieldTitle:         "Title",
		FieldSession:       "Session",
		FieldProject:       "Project",
		FieldDirectory:     "Directory",
		FieldPathStatus:    "Path Status",
		FieldMessageCount:  "Messages",
		FieldPartCount:     "Parts",
		FieldSize:          "Size",
		FieldLargeSession:  "Large Session",
		PathExists:         "Exists",
		PathMissing:        "Missing",
		Yes:                "Yes",
		No:                 "No",
		FieldUpdated:       "Updated",
		FieldCreated:       "Created",
		FieldModel:         "Model",
		FieldAgent:         "Agent",
		FieldCost:          "Cost",
		FieldTokens:        "Tokens",
		FieldResumeCommand: "Resume Command",
		TokensFormat:       "in %d / out %d / reasoning %d / cache %d",
		Weekdays:           [7]string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"},
	}
}
