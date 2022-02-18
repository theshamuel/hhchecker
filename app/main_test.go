package main

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	"log"
	"os"
	"strings"
	"testing"
)

func TestMain(m *testing.M)  {
	goleak.VerifyTestMain(m,
		goleak.IgnoreTopFunction("github.com/theshamuel/hhchecker/app.init.0.func1"))
}

func TestGetStackTrace(t *testing.T) {
	stackTrace := getStackTrace()
	assert.True(t, strings.Contains(stackTrace, "goroutine"))
	assert.True(t, strings.Contains(stackTrace, "[running]"))
	//assert.True(t, strings.Contains(stackTrace, "medregistry20/app/main.go"))
	assert.True(t, strings.Contains(stackTrace, "hhchecker/app.getStackTrace"))
	t.Logf("\n STACKTRACE: %s", stackTrace)
}

func captureStdout(f func()) string {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	f()
	return buf.String()
}