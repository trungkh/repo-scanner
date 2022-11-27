package serror

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"repo-scanner/internal/constants"
	"repo-scanner/internal/utils/utstring"
)

func isLocal() bool {
	return strings.ToLower(utstring.Env(constants.AppEnv, constants.EnvLocal)) == constants.EnvLocal
}

func printErr(m string) {
	fmt.Fprintln(os.Stderr, m)
}

func exit() {
	err := syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	if err != nil {
		os.Exit(1)
	}
}

func getPath(val string) string {
	for _, v := range rootPaths {
		if strings.HasPrefix(val, v) {
			val = utstring.Sub(val, len(v), 0)
			return val
		}
	}

	return val
}
