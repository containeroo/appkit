package logging

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestSetup(t *testing.T) {
	var output bytes.Buffer
	logger := Setup(LogFormatJSON, true, &output)
	logger.Debug("test", slog.String("category", "appkit"))
	if !strings.Contains(output.String(), `"level":"DEBUG"`) {
		t.Fatal(output.String())
	}
}
