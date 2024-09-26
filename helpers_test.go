package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"testing"

	"github.com/tuxdude/zzzlog"
	"github.com/tuxdude/zzzlogi"
)

var (
	// TODO: Remove this after configEnv gets actually used in the code.
	_ = newConfigEnv(newFakeHostInfo())
)

type testContextInfo struct {
	logger          zzzlogi.Logger
	dockerHost      dockerAPIClient
	useRealHostInfo bool
}

type testOSEnvMap map[string]string

type wrappedReader func(p []byte) (int, error)

func (actual wrappedReader) Read(p []byte) (int, error) {
	return actual(p)
}

func newVanillaTestContext() context.Context {
	return newTestContext(
		&testContextInfo{
			dockerHost: newEmptyFakeDockerHost(),
		})
}

func newTestContext(info *testContextInfo) context.Context {
	ctx := context.Background()
	if info.logger != nil {
		ctx = withLogger(ctx, info.logger)
	} else {
		ctx = withLogger(ctx, newTestLogger())
	}
	if info.dockerHost != nil {
		ctx = withDockerAPIClient(ctx, info.dockerHost)
	}
	if !info.useRealHostInfo {
		ctx = withHostInfo(ctx, newFakeHostInfo())
	}
	return ctx
}

func newTestLogger() zzzlogi.Logger {
	config := zzzlog.NewConsoleLoggerConfig()
	config.SkipCallerInfo = true
	config.PanicInFatal = true
	return zzzlog.NewLogger(config)
}

func newCapturingTestLogger(lvl zzzlog.Level, w io.Writer) zzzlogi.Logger {
	config := zzzlog.NewConsoleLoggerConfig()
	config.MaxLevel = lvl
	config.SkipCallerInfo = true
	config.PanicInFatal = true
	config.Dest = w
	return zzzlog.NewLogger(config)
}

func newCapturingVanillaTestLogger(lvl zzzlog.Level, w io.Writer) zzzlogi.Logger {
	config := zzzlog.NewVanillaLoggerConfig()
	config.MaxLevel = lvl
	config.Dest = w
	config.PanicInFatal = true
	return zzzlog.NewLogger(config)
}

func setTestEnv(envs testOSEnvMap) {
	for k, v := range envs {
		err := os.Setenv(k, v)
		if err != nil {
			panic(err)
		}
	}
}

func clearTestEnv(envs testOSEnvMap) {
	for k := range envs {
		err := os.Unsetenv(k)
		if err != nil {
			panic(err)
		}
	}
}

func testPanicWithOutput(t *testing.T, methodUnderTest string, testCase string, out fmt.Stringer, wantPanic string) {
	gotR := recover()
	if gotR == nil {
		t.Errorf(
			"%s\nTest Case: %q\nFailure: panic expected but did not encounter one\n\nOut:\n%s\nReason: want panic = %q",
			methodUnderTest, testCase, out.String(), wantPanic)
		return
	}

	gotPanic, ok := gotR.(string)
	if !ok {
		t.Errorf(
			"%s\nTest Case: %q\nFailure: recovered interface from panic is not of type string\n\nOut:\n%s\nReason: want = %q",
			methodUnderTest, testCase, out.String(), wantPanic)
		return
	}

	match, err := regexp.MatchString(fmt.Sprintf("^%s$", wantPanic), gotPanic)
	if err != nil {
		t.Errorf(
			"%s\nTest Case: %q\nFailure: unexpected exception while matching against gotPanic\n\nOut:\n%s\nReason: error = %v",
			methodUnderTest, testCase, out.String(), err)
		return
	}

	if !match {
		t.Errorf(
			"%s\nTest Case: %q\nFailure: gotPanic did not match the wantPanic regex\n\nOut:\n%s\nReason:\n\ngotPanic = %q\n\twant = %q",
			methodUnderTest, testCase, out.String(), gotPanic, wantPanic)
	}
}

func pwd() string {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return pwd
}

func newInt(i int) *int {
	return &i
}

func newLogLevel(lvl zzzlog.Level) *zzzlog.Level {
	return &lvl
}
