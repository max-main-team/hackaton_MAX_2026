package handler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractPDFTextLocalFiles(t *testing.T) {
	pdfs, _ := filepath.Glob("../../../cv*.pdf")
	pdfs2, _ := filepath.Glob("../../../big-resume.pdf")
	pdfs = append(pdfs, pdfs2...)
	if len(pdfs) == 0 {
		t.Skip("no local sample PDFs")
	}
	for _, p := range pdfs {
		t.Run(filepath.Base(p), func(t *testing.T) {
			if _, err := os.Stat(p); err != nil {
				t.Skip("missing file")
			}
			text, err := extractPDFText(p)
			require.NoError(t, err)
			assert.GreaterOrEqual(t, len(text), parseTextMinLen, "too little text")
			assert.NotContains(t, text, "\r")
			assert.NotContains(t, text, "  ")
			assert.True(t, strings.TrimSpace(text) != "")
			t.Logf("%s: %d chars", filepath.Base(p), len(text))
		})
	}
}
