//go:build !windows

package update

// SwapAndRestart has no implementation outside Windows. The check path still
// works everywhere; only automatic installation is refused.
func SwapAndRestart(exePath, staged string, restart bool) error {
	return errCode(CodeUnsupported, "automatic install is supported on Windows only")
}
