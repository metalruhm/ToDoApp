package model

import (
	"strings"
	"testing"
)

func TestNormalizeTitle(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"通常", "牛乳を買う", "牛乳を買う", false},
		{"前後の空白は除去", "  牛乳  ", "牛乳", false},
		{"途中の空白は残す", "牛乳 パン", "牛乳 パン", false},
		{"空", "", "", true},
		{"半角空白のみ", "   ", "", true},
		{"全角空白のみ", "　　", "", true},
		{"タブと改行のみ", "\t\n", "", true},
		{"1文字", "a", "a", false},
		{"100文字(半角)", strings.Repeat("a", 100), strings.Repeat("a", 100), false},
		{"101文字(半角)", strings.Repeat("a", 101), "", true},
		{"100文字(日本語)", strings.Repeat("あ", 100), strings.Repeat("あ", 100), false},
		{"101文字(日本語)", strings.Repeat("あ", 101), "", true},
		{"空白を除いて100文字", "  " + strings.Repeat("あ", 100) + "  ", strings.Repeat("あ", 100), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeTitle(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
