package tui

import (
	"os"
	"strings"
)

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
	StatusSessions        string
	StatusMatched         string
	ClipboardUnavailable  string
	CopiedSessionID       string
	RecentUserMessages    string
	NoTextPreview         string
	ReadOnlyNotice        string
	FooterBrowse          string
	FooterSearch          string
	FooterHelp            string
	ActionHint            string
	HelpTitle             string
	HelpLines             []string
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
			StatusSessions:        "%d 个会话",
			StatusMatched:         "%d/%d 匹配",
			ClipboardUnavailable:  "剪贴板不可用：%s",
			CopiedSessionID:       "已复制 session id",
			RecentUserMessages:    "最近用户消息",
			NoTextPreview:         "没有找到文本预览",
			ReadOnlyNotice:        "MVP 以只读方式运行，不会修改 OpenCode 数据。",
			FooterBrowse:          "浏览",
			FooterSearch:          "搜索",
			FooterHelp:            "/ 搜索 | Enter 恢复 | p/Ctrl-P 预览 | y 复制 ID | ? 帮助 | q 退出",
			ActionHint:            "Enter 恢复 | p 预览 | y 复制 ID",
			HelpTitle:             "lazyOpencodeSession 帮助",
			HelpLines: []string{
				"q / Ctrl+C     退出",
				"↑/k ↓/j        移动选择",
				"搜索中 Ctrl-K/J 上下移动，j/k 仍输入文字",
				"搜索中 Ctrl-P 预览当前会话",
				"PageUp/Down    翻页",
				"/              搜索元数据，空格分隔多个关键词",
				"Esc            清空搜索 / 关闭帮助",
				"Enter          恢复选中会话",
				"p              预览最近用户消息",
				"y              复制 session id",
				"r              重新加载会话",
				"?              切换帮助",
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
		StatusSessions:        "%d sessions",
		StatusMatched:         "%d/%d matched",
		ClipboardUnavailable:  "clipboard unavailable: %s",
		CopiedSessionID:       "copied session id",
		RecentUserMessages:    "Recent User Messages",
		NoTextPreview:         "No text preview found",
		ReadOnlyNotice:        "MVP is read-only. It does not modify OpenCode data.",
		FooterBrowse:          "browse",
		FooterSearch:          "search",
		FooterHelp:            "/ search | Enter resume | p/Ctrl-P preview | y copy id | ? help | q quit",
		ActionHint:            "Enter resume | p preview | y copy id",
		HelpTitle:             "lazyOpencodeSession Help",
		HelpLines: []string{
			"q / Ctrl+C     Quit",
			"↑/k ↓/j        Move selection",
			"Search Ctrl-K/J Move; j/k still type text",
			"Search Ctrl-P Preview current session",
			"PageUp/Down    Jump list",
			"/              Search metadata; split terms by spaces",
			"Esc            Clear search / close help",
			"Enter          Resume selected session",
			"p              Preview recent user messages",
			"y              Copy session id",
			"r              Reload sessions",
			"?              Toggle help",
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
	}
}
