package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var platforms = map[string]string{
	"darwin/amd64":  "darwin_amd64",
	"darwin/arm64":  "darwin_arm64",
	"linux/amd64":   "linux_amd64",
	"linux/arm64":   "linux_arm64",
	"windows/amd64": "windows_amd64",
	"windows/arm64": "windows_arm64",
}

func newUpdateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Self-update to the latest version",
		Args:  noPositionalArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get current version
			current := Version
			if current == "dev" {
				return writeError(cmd, "update_error", "dev build — use git pull && go build", "")
			}

			// Fetch latest from GitHub
			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Get("https://api.github.com/repos/izzzzzi/agent-asearch/releases/latest")
			if err != nil {
				return writeError(cmd, "update_error", fmt.Sprintf("failed to fetch latest release: %v", err), "")
			}
			defer resp.Body.Close()

			var release struct {
				TagName string `json:"tag_name"`
				Assets  []struct {
					Name               string `json:"name"`
					BrowserDownloadURL string `json:"browser_download_url"`
				} `json:"assets"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
				return writeError(cmd, "update_error", fmt.Sprintf("failed to parse release: %v", err), "")
			}

			latest := strings.TrimPrefix(release.TagName, "v")
			if latest == current {
				return writeJSON(cmd, map[string]any{
					"ok":      true,
					"current": current,
					"latest":  latest,
					"message": "already up to date",
				})
			}

			// Find matching asset
			plat := platforms[runtime.GOOS+"/"+runtime.GOARCH]
			if plat == "" {
				return writeError(cmd, "update_error", "unsupported platform: "+runtime.GOOS+"/"+runtime.GOARCH, "")
			}

			archiveName := fmt.Sprintf("asearch_%s_%s.tar.gz", latest, plat)
			var downloadURL string
			for _, a := range release.Assets {
				if a.Name == archiveName {
					downloadURL = a.BrowserDownloadURL
					break
				}
			}
			if downloadURL == "" {
				return writeError(cmd, "update_error", fmt.Sprintf("no asset found for %s", archiveName), "")
			}

			// Download
			cmd.Printf("Downloading %s...\n", archiveName)
			binResp, err := client.Get(downloadURL)
			if err != nil {
				return writeError(cmd, "update_error", fmt.Sprintf("download failed: %v", err), "")
			}
			defer binResp.Body.Close()

			// Get current binary path
			binPath, err := os.Executable()
			if err != nil {
				return writeError(cmd, "update_error", fmt.Sprintf("can't find self: %v", err), "")
			}

			// Extract to temp dir
			tmpDir, err := os.MkdirTemp("", "asearch-update")
			if err != nil {
				return writeError(cmd, "update_error", fmt.Sprintf("temp dir: %v", err), "")
			}
			defer os.RemoveAll(tmpDir)

			// Read archive body
			data, err := io.ReadAll(binResp.Body)
			if err != nil {
				return writeError(cmd, "update_error", fmt.Sprintf("read: %v", err), "")
			}

			// Save archive
			tarPath := filepath.Join(tmpDir, "update.tar.gz")
			if err := os.WriteFile(tarPath, data, 0600); err != nil {
				return writeError(cmd, "update_error", fmt.Sprintf("write: %v", err), "")
			}

			// Extract — use shell tar or archive/zip for gzip+tar
			// Simple approach: pipe through tar
			newBin := filepath.Join(tmpDir, "asearch")
			if err := extractTarball(tarPath, newBin); err != nil {
				return writeError(cmd, "update_error", fmt.Sprintf("extract: %v", err), "")
			}

			// Replace current binary
			backupPath := binPath + ".bak"
			os.Rename(binPath, backupPath) // backup old binary
			if err := os.Rename(newBin, binPath); err != nil {
				// Restore backup
				os.Rename(backupPath, binPath)
				return writeError(cmd, "update_error", fmt.Sprintf("replace: %v", err), "")
			}
			os.Chmod(binPath, 0755)
			os.Remove(backupPath) // cleanup backup

			return writeJSON(cmd, map[string]any{
				"ok":      true,
				"message": fmt.Sprintf("updated from %s to %s", current, latest),
				"current": current,
				"latest":  latest,
			})
		},
	}
}

func extractTarball(tarPath, binPath string) error {
	// Pipe through system tar
	cmd := exec.Command("tar", "xzf", tarPath, "-C", filepath.Dir(binPath))
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("tar: %s: %w", string(out), err)
	}
	// The binary in the archive is at the root with name "asearch"
	extracted := filepath.Join(filepath.Dir(binPath), "asearch")
	if _, err := os.Stat(extracted); err != nil {
		return fmt.Errorf("binary not found in archive: %w", err)
	}
	return os.Rename(extracted, binPath)
}
