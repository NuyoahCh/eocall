package rag

import (
	"testing"
)

func TestService_splitIntoChunks(t *testing.T) {
	s := &Service{
		chunkSize:    10,
		chunkOverlap: 2,
	}

	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{
			name:     "short text",
			text:     "hello",
			expected: 1,
		},
		{
			name:     "exact chunk size",
			text:     "0123456789",
			expected: 1,
		},
		{
			name:     "multiple chunks",
			text:     "0123456789abcdefghij",
			expected: 3, // 0-9, 8-17, 16-19
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunks := s.splitIntoChunks(tt.text)
			if len(chunks) != tt.expected {
				t.Errorf("expected %d chunks, got %d", tt.expected, len(chunks))
			}
		})
	}
}

func TestService_FormatContext(t *testing.T) {
	s := &Service{}

	docs := []Document{
		{ID: "1", Content: "first document", Metadata: map[string]string{"source": "doc1.md"}},
		{ID: "2", Content: "second document", Metadata: map[string]string{"source": "doc2.md"}},
	}

	result := s.FormatContext(docs)

	if result == "" {
		t.Error("expected non-empty context")
	}

	// 检查是否包含文档内容
	if !contains(result, "first document") || !contains(result, "second document") {
		t.Error("context should contain document content")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
