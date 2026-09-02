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
	FieldUpdated          string
	FieldCreated          string
	FieldModel            string
	FieldAgent            string
	FieldCost             string
	FieldTokens           string
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
			ClipboardUnavailable:  "剪贴板不可用：%s",
			CopiedSessionID:       "已复制 session id",
			RecentUserMessages:    "最近用户消息",
			NoTextPreview:         "没有找到文本预览",
			ReadOnlyNotice:        "MVP 以只读方式运行，不会修改 OpenCode 数据。",
			FooterBrowse:          "浏览",
			FooterSearch:          "搜索",
			FooterHelp:            "/ 搜索 | Enter 恢复 | p 预览 | y 复制 ID | ? 帮助 | q 退出",
			ActionHint:            "Enter 恢复 | p 预览 | y 复制 ID",
			HelpTitle:             "lazyOpencodeSession 帮助",
			HelpLines: []string{
				"q / Ctrl+C     退出",
				"↑/k ↓/j        移动选择",
				"PageUp/Down    翻页",
				"/              搜索元数据",
				"Esc            清空搜索 / 关闭帮助",
				"Enter          恢复选中会话",
				"p              预览最近用户消息",
				"y              复制 session id",
				"r              重新加载会话",
				"?              切换帮助",
			},
			FieldTitle:     "标题",
			FieldSession:   "会话",
			FieldProject:   "项目",
			FieldDirectory: "目录",
			FieldUpdated:   "更新",
			FieldCreated:   "创建",
			FieldModel:     "模型",
			FieldAgent:     "代理",
			FieldCost:      "费用",
			FieldTokens:    "Token",
			TokensFormat:   "输入 %d / 输出 %d / 推理 %d / 缓存 %d",
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
		ClipboardUnavailable:  "clipboard unavailable: %s",
		CopiedSessionID:       "copied session id",
		RecentUserMessages:    "Recent User Messages",
		NoTextPreview:         "No text preview found",
		ReadOnlyNotice:        "MVP is read-only. It does not modify OpenCode data.",
		FooterBrowse:          "browse",
		FooterSearch:          "search",
		FooterHelp:            "/ search | Enter resume | p preview | y copy id | ? help | q quit",
		ActionHint:            "Enter resume | p preview | y copy id",
		HelpTitle:             "lazyOpencodeSession Help",
		HelpLines: []string{
			"q / Ctrl+C     Quit",
			"↑/k ↓/j        Move selection",
			"PageUp/Down    Jump list",
			"/              Search metadata",
			"Esc            Clear search / close help",
			"Enter          Resume selected session",
			"p              Preview recent user messages",
			"y              Copy session id",
			"r              Reload sessions",
			"?              Toggle help",
		},
		FieldTitle:     "Title",
		FieldSession:   "Session",
		FieldProject:   "Project",
		FieldDirectory: "Directory",
		FieldUpdated:   "Updated",
		FieldCreated:   "Created",
		FieldModel:     "Model",
		FieldAgent:     "Agent",
		FieldCost:      "Cost",
		FieldTokens:    "Tokens",
		TokensFormat:   "in %d / out %d / reasoning %d / cache %d",
	}
}
