package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/asikeida/lazyOpencodeSession/internal/opencode"
)

func BenchmarkApplySearch(b *testing.B) {
	model := Model{
		query:        "windows checksum",
		catalog:      make([]opencode.Session, 0, 500),
		memories:     make(map[string][]opencode.UserMemory, 500),
		memorySearch: make(map[string]string, 500),
	}
	for index := range 500 {
		id := fmt.Sprintf("ses_bench_%03d", index)
		title := fmt.Sprintf("Benchmark session %03d", index)
		memory := "routine benchmark message"
		if index%10 == 0 {
			title += " Windows"
			memory += " with checksum verification"
		}
		model.catalog = append(model.catalog, opencode.Session{ID: id, Title: title})
		model.memories[id] = []opencode.UserMemory{{SessionID: id, Text: memory}}
		model.memorySearch[id] = strings.ToLower(memory)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		model.applySearch()
	}
}
