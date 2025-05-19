// SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

// RunCommandWithTimeout runs a command with the provided timeout duration and returns its
// combined output.
func RunCommandWithTimeout(ctx context.Context, timeout time.Duration, name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return exec.CommandContext(ctx, name, args...).CombinedOutput()
}

// CmdArgsToStr returns a space separated string for cmdArgs.
func CmdArgsToStr(cmdArgs []string) string {
	return strings.Join(cmdArgs, " ")
}
