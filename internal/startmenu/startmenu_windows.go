package startmenu

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// EnsureShortcut creates a per-user Start Menu shortcut for the running
// executable if one does not already exist. No UAC elevation is required
// because the target is inside %APPDATA%.
func EnsureShortcut(appName string) error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable path: %w", err)
	}

	appData := os.Getenv("APPDATA")
	if appData == "" {
		return fmt.Errorf("APPDATA environment variable not set")
	}

	linkPath := filepath.Join(appData, `Microsoft\Windows\Start Menu\Programs`, appName+".lnk")

	// Already exists — nothing to do.
	if _, err := os.Stat(linkPath); err == nil {
		return nil
	}

	exeEscaped := strings.ReplaceAll(exePath, `'`, `''`)
	linkEscaped := strings.ReplaceAll(linkPath, `'`, `''`)
	nameEscaped := strings.ReplaceAll(appName, `'`, `''`)

	ps := fmt.Sprintf(
		`$ws = New-Object -ComObject WScript.Shell; `+
			`$sc = $ws.CreateShortcut('%s'); `+
			`$sc.TargetPath = '%s'; `+
			`$sc.Description = '%s'; `+
			`$sc.IconLocation = '%s,0'; `+
			`$sc.Save()`,
		linkEscaped, exeEscaped, nameEscaped, exeEscaped,
	)

	out, err := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", ps).CombinedOutput()
	if err != nil {
		return fmt.Errorf("create shortcut: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
