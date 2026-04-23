package util

import (
	"net/http"
	"testing"
)

func TestGetFileNameWithMalformedFilenameStar(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		Header: http.Header{
			"Content-Disposition": []string{
				`attachment;filename=".docx";filename*=UTF-8''%E4%B8%80%E5%9C%BA%E5%AE%89%E9%9D%99%E8%80%8C%E5%85%8B%E5%88%B6%E8%B6%8A%E9%87%8E%E8%B5%9B:%E4%B8%9C%E6%B5%B7%E4%BA%91%E9%A1%B6%E8%B7%91%E5%B1%B1%E8%B5%9B.docx`,
			},
		},
	}

	fileName, err := getFileName(resp)
	if err != nil {
		t.Fatalf("getFileName() error = %v", err)
	}

	want := "一场安静而克制越野赛:东海云顶跑山赛.docx"
	if fileName != want {
		t.Fatalf("getFileName() = %q, want %q", fileName, want)
	}
}

func TestGetFileNameFallbackToFilename(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		Header: http.Header{
			"Content-Disposition": []string{
				`attachment; filename="simple.docx"`,
			},
		},
	}

	fileName, err := getFileName(resp)
	if err != nil {
		t.Fatalf("getFileName() error = %v", err)
	}

	if fileName != "simple.docx" {
		t.Fatalf("getFileName() = %q, want %q", fileName, "simple.docx")
	}
}
