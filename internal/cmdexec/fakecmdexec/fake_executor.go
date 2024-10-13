package fakecmdexec

import (
	"context"
	"fmt"
	"strings"

	"github.com/tuxdudehomelab/homelab/internal/cmdexec"
)

type FakeExecutor struct {
	validCmds cmdOutputMap
	errorCmds cmdErrorMap
}

type cmdOutputMap map[string]argsOutputMap

type argsOutputMap map[string]string

type cmdErrorMap map[string]argsErrorMap

type argsErrorMap map[string]error

func newValidCmdsMap(cmds []FakeValidCmdInfo) cmdOutputMap {
	res := cmdOutputMap{}
	for _, cmd := range cmds {
		bin := cmd.Cmd[0]
		args := cmd.Cmd[1:]
		out := cmd.Output
		var a argsOutputMap

		if val, found := res[bin]; found {
			a = val
		} else {
			a = argsOutputMap{}
			res[bin] = a
		}
		a[argsToStr(args...)] = out
	}
	return res
}

func newErrorCmdsMap(cmds []FakeErrorCmdInfo) cmdErrorMap {
	res := cmdErrorMap{}
	for _, cmd := range cmds {
		bin := cmd.Cmd[0]
		args := cmd.Cmd[1:]
		err := cmd.Err
		var a argsErrorMap

		if val, found := res[bin]; found {
			a = val
		} else {
			a = argsErrorMap{}
			res[bin] = a
		}
		a[argsToStr(args...)] = err
	}
	return res
}

func (c cmdOutputMap) output(bin string, args ...string) (string, bool) {
	if argsMap, found := c[bin]; found {
		return argsMap.output(args...)
	}
	return "", false
}

func (a argsOutputMap) output(args ...string) (string, bool) {
	res, found := a[argsToStr(args...)]
	return res, found
}

func (c cmdErrorMap) err(bin string, args ...string) (error, bool) {
	if argsMap, found := c[bin]; found {
		return argsMap.err(args...)
	}
	return nil, false
}

func (a argsErrorMap) err(args ...string) (error, bool) {
	res, found := a[argsToStr(args...)]
	return res, found
}

type FakeExecutorInitInfo struct {
	ValidCmds []FakeValidCmdInfo
	ErrorCmds []FakeErrorCmdInfo
}

type FakeValidCmdInfo struct {
	Cmd    []string
	Output string
}

type FakeErrorCmdInfo struct {
	Cmd []string
	Err error
}

func FakeExecutorFromContext(ctx context.Context) *FakeExecutor {
	if f, ok := cmdexec.MustExecutor(ctx).(*FakeExecutor); ok {
		return f
	}
	panic("unable to convert the retrieved executor to fake executor")
}

func NewEmptyFakeExecutor() *FakeExecutor {
	return NewFakeExecutor(&FakeExecutorInitInfo{})
}

func NewFakeExecutor(initInfo *FakeExecutorInitInfo) *FakeExecutor {
	return &FakeExecutor{
		validCmds: newValidCmdsMap(initInfo.ValidCmds),
		errorCmds: newErrorCmdsMap(initInfo.ErrorCmds),
	}
}

func (f *FakeExecutor) Run(bin string, args ...string) (string, error) {
	if err, found := f.errorCmds.err(bin, args...); found {
		return "", err
	}
	if out, found := f.validCmds.output(bin, args...); found {
		return out, nil
	}
	return "", fmt.Errorf("invalid fake executor command %s %q", bin, args)
}

func argsToStr(args ...string) string {
	argsStr := strings.Join(args, "__@@__")
	if len(args) > 0 {
		argsStr = "__@@__" + argsStr
	}
	return argsStr
}
