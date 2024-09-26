package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
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

func testLogErrorNotNil(t *testing.T, methodUnderTest string, testCase string, gotErr error) {
	t.Errorf(
		"%s\nTest Case: %q\nFailure: gotErr != nil\nReason: %v",
		methodUnderTest, testCase, gotErr)
}

func testLogErrorNotNilWithOutput(t *testing.T, methodUnderTest string, testCase string, out fmt.Stringer, gotErr error) {
	t.Errorf(
		"%s\nTest Case: %q\nFailure: gotErr != nil\n\nOut:\n%s\nReason: %v",
		methodUnderTest, testCase, out.String(), gotErr)
}

func testLogErrorNil(t *testing.T, methodUnderTest string, testCase string, want string) {
	t.Errorf(
		"%s\nTest Case: %q\nFailure: gotErr == nil\nReason: want = %q",
		methodUnderTest, testCase, want)
}

func testLogErrorNilWithOutput(t *testing.T, methodUnderTest string, testCase string, out fmt.Stringer, want string) {
	t.Errorf(
		"%s\nTest Case: %q\nFailure: gotErr == nil\n\nOut:\n%s\nReason: want = %q",
		methodUnderTest, testCase, out.String(), want)
}

func testLogCustomWithOutput(t *testing.T, methodUnderTest string, testCase string, out fmt.Stringer, custom string) {
	t.Errorf(
		"%s\nTest Case: %q\n\nOut:\n%s\nReason: %s",
		methodUnderTest, testCase, out.String(), custom)
}

func testRegexMatch(t *testing.T, methodUnderTest string, testCase string, desc string, want string, got string) bool {
	match, err := regexp.MatchString(fmt.Sprintf("^%s$", want), got)
	if err != nil {
		t.Errorf(
			"%s\nTest Case: %q\nFailure: unexpected exception while matching against %s\nReason: error = %v",
			methodUnderTest, testCase, desc, err)
		return false
	}

	if !match {
		t.Errorf(
			"%s\nTest Case: %q\nFailure: %s did not match the want regex\nReason:\n\nGot:\n%s\nWant:\n%s",
			methodUnderTest, testCase, desc, got, want)
		return false
	}

	return true
}

func testRegexMatchWithOutput(t *testing.T, methodUnderTest string, testCase string, out fmt.Stringer, desc string, want string, got string) bool {
	match, err := regexp.MatchString(fmt.Sprintf("^%s$", want), got)
	if err != nil {
		t.Errorf(
			"%s\nTest Case: %q\nFailure: unexpected exception while matching against %s\n\nOut:\n%s\nReason: error = %v",
			methodUnderTest, testCase, desc, out.String(), err)
		return false
	}

	if !match {
		t.Errorf(
			"%s\nTest Case: %q\nFailure: %s did not match the want regex\n\nOut:\n%s\nReason:\n\nGot:\n%s\nWant:\n%s",
			methodUnderTest, testCase, desc, out.String(), got, want)
		return false
	}

	return true
}

func testRegexMatchJoinNewLines(t *testing.T, methodUnderTest string, testCase string, desc string, want string, got string) bool {
	multiNewLineRegex, err := regexp.Compile(`\n+`)
	if err != nil {
		panic(err)
	}
	got = strings.TrimSpace(multiNewLineRegex.ReplaceAllString(got, "\n"))
	return testRegexMatch(t, methodUnderTest, testCase, desc, want, got)
}

func testCmpDiff(t *testing.T, methodUnderTest string, testCase string, desc string, want interface{}, got interface{}) bool {
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf(
			"%s\nTest Case: %q\nFailure: %s - got did not match the want\nDiff(-want +got): %s",
			methodUnderTest, testCase, desc, diff)
		return false
	}
	return true
}

func testExpectPanicWithOutput(t *testing.T, methodUnderTest string, testCase string, out fmt.Stringer, wantPanic string) {
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
