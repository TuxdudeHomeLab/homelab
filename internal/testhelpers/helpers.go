package testhelpers

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var (
	multiNewLineRegex = regexp.MustCompile(`\n+`)
)

type TestEnvMap map[string]string

func SetTestEnv(t *testing.T, envs TestEnvMap) {
	t.Helper()
	for k, v := range envs {
		t.Setenv(k, v)
	}
}

func LogErrorNotNil(t *testing.T, methodUnderTest string, testCase string, gotErr error) {
	t.Helper()
	t.Errorf(
		"%s\nTest Case: %q\nFailure: gotErr != nil\nReason: %v",
		methodUnderTest, testCase, gotErr)
}

func LogErrorNotNilWithOutput(t *testing.T, methodUnderTest string, testCase string, out fmt.Stringer, gotErr error) {
	t.Helper()
	t.Errorf(
		"%s\nTest Case: %q\nFailure: gotErr != nil\n\nOut:\n%s\nReason: %v",
		methodUnderTest, testCase, out.String(), gotErr)
}

func LogErrorNil(t *testing.T, methodUnderTest string, testCase string, want string) {
	t.Helper()
	t.Errorf(
		"%s\nTest Case: %q\nFailure: gotErr == nil\nReason: want = %q",
		methodUnderTest, testCase, want)
}

func LogErrorNilWithOutput(t *testing.T, methodUnderTest string, testCase string, out fmt.Stringer, want string) {
	t.Helper()
	t.Errorf(
		"%s\nTest Case: %q\nFailure: gotErr == nil\n\nOut:\n%s\nReason: want = %q",
		methodUnderTest, testCase, out.String(), want)
}

func LogCustom(t *testing.T, methodUnderTest string, testCase string, custom string) {
	t.Helper()
	t.Errorf(
		"%s\nTest Case: %q\nReason: %s",
		methodUnderTest, testCase, custom)
}

func LogCustomWithOutput(t *testing.T, methodUnderTest string, testCase string, out fmt.Stringer, custom string) {
	t.Helper()
	t.Errorf(
		"%s\nTest Case: %q\n\nOut:\n%s\nReason: %s",
		methodUnderTest, testCase, out.String(), custom)
}

func RegexMatch(t *testing.T, methodUnderTest string, testCase string, desc string, want string, got string) bool {
	t.Helper()
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

func RegexMatchWithOutput(t *testing.T, methodUnderTest string, testCase string, out fmt.Stringer, desc string, want string, got string) bool {
	t.Helper()
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

func RegexMatchJoinNewLines(t *testing.T, methodUnderTest string, testCase string, desc string, want string, got string) bool {
	t.Helper()
	got = strings.TrimSpace(multiNewLineRegex.ReplaceAllString(got, "\n"))
	return RegexMatch(t, methodUnderTest, testCase, desc, want, got)
}

func CmpDiff(t *testing.T, methodUnderTest string, testCase string, desc string, want interface{}, got interface{}) bool {
	t.Helper()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf(
			"%s\nTest Case: %q\nFailure: %s - got did not match the want\nDiff(-want +got): %s",
			methodUnderTest, testCase, desc, diff)
		return false
	}
	return true
}

func ExpectPanic(t *testing.T, methodUnderTest string, testCase string, wantPanic string) {
	t.Helper()
	gotR := recover()
	if gotR == nil {
		t.Errorf(
			"%s\nTest Case: %q\nFailure: panic expected but did not encounter one\nReason: want panic = %q",
			methodUnderTest, testCase, wantPanic)
		return
	}

	gotPanic, ok := gotR.(string)
	if !ok {
		t.Errorf(
			"%s\nTest Case: %q\nFailure: recovered interface from panic is not of type string\nReason: want = %q",
			methodUnderTest, testCase, wantPanic)
		return
	}

	match, err := regexp.MatchString(fmt.Sprintf("^%s$", wantPanic), gotPanic)
	if err != nil {
		t.Errorf(
			"%s\nTest Case: %q\nFailure: unexpected exception while matching against gotPanic\nReason: error = %v",
			methodUnderTest, testCase, err)
		return
	}

	if !match {
		t.Errorf(
			"%s\nTest Case: %q\nFailure: gotPanic did not match the wantPanic regex\nReason:\ngotPanic = %q\nwant = %q",
			methodUnderTest, testCase, gotPanic, wantPanic)
	}
}

func ExpectPanicWithOutput(t *testing.T, methodUnderTest string, testCase string, out fmt.Stringer, wantPanic string) {
	t.Helper()
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
			"%s\nTest Case: %q\nFailure: gotPanic did not match the wantPanic regex\n\nOut:\n%s\nReason:\ngotPanic = %q\nwant = %q",
			methodUnderTest, testCase, out.String(), gotPanic, wantPanic)
	}
}
