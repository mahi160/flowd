package fw

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

const repoURL = "https://github.com/mahi160/flowd"

func cmdUpdate() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update fw to the latest version from GitHub",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			fmt.Println("checking for updates...")

			latest, err := latestRemoteTag(ctx)
			if err != nil {
				return fmt.Errorf("fetch latest tag: %w", err)
			}

			// Normalise: Version may be "dev", "0.9.0", or "v0.9.0[-extra]".
			current := Version
			if current != "dev" && !strings.HasPrefix(current, "v") {
				current = "v" + current
			}
			// Strip any git-describe suffix (v0.9.0-1-gabcdef → v0.9.0)
			if idx := strings.Index(current, "-"); idx != -1 && strings.HasPrefix(current, "v") {
				current = current[:idx]
			}

			fmt.Printf("  current : %s\n  latest  : %s\n\n", current, latest)

			if current == latest {
				fmt.Println("already up to date.")
				return nil
			}

			// Resolve the path of the running binary.
			exe, err := os.Executable()
			if err != nil {
				return fmt.Errorf("resolve executable: %w", err)
			}
			exe, err = filepath.EvalSymlinks(exe)
			if err != nil {
				return fmt.Errorf("eval symlinks: %w", err)
			}

			// Download the pre-built binary for this platform from GitHub Releases.
			assetName := fmt.Sprintf("fw-%s-%s", runtime.GOOS, runtime.GOARCH)
			downloadURL := fmt.Sprintf("%s/releases/download/%s/%s", repoURL, latest, assetName)
			fmt.Printf("downloading %s...\n", downloadURL)

			// Download into the same directory as the target binary so the
			// final os.Rename is atomic (same filesystem) and never leaves a
			// half-written executable. Overwriting a running binary in place
			// can crash the process (notably on macOS with code signing);
			// rename swaps the inode instead.
			tmpFile, err := os.CreateTemp(filepath.Dir(exe), ".fw-update-*")
			if err != nil {
				// Target dir not writable — fall back to /tmp; the sudo path
				// below will move it into place.
				tmpFile, err = os.CreateTemp("", "fw-update-*")
				if err != nil {
					return err
				}
			}
			tmpPath := tmpFile.Name()
			defer os.Remove(tmpPath)

			if err := downloadFile(ctx, downloadURL, tmpFile); err != nil {
				tmpFile.Close()
				return fmt.Errorf("download binary: %w", err)
			}
			tmpFile.Close()

			if err := verifyChecksum(ctx, tmpPath, downloadURL+".sha256"); err != nil {
				return fmt.Errorf("checksum: %w", err)
			}

			if err := os.Chmod(tmpPath, 0755); err != nil {
				return fmt.Errorf("chmod: %w", err)
			}

			// Replace the running binary atomically.
			if err := os.Rename(tmpPath, exe); err != nil {
				fmt.Printf("\n%s is not writable — running sudo mv\n", exe)
				sudo := exec.CommandContext(ctx, "sudo", "mv", tmpPath, exe)
				sudo.Stdin = os.Stdin
				sudo.Stdout = os.Stdout
				sudo.Stderr = os.Stderr
				if sudoErr := sudo.Run(); sudoErr != nil {
					fmt.Printf("sudo failed: %v\nrun manually:\n  sudo mv %s %s\n", sudoErr, tmpPath, exe)
					return nil
				}
			}

			fmt.Printf("\nupdated to %s ✓\n", latest)
			return nil
		},
	}
}

// verifyChecksum fetches the .sha256 sidecar for a release asset and checks
// the downloaded file against it. Releases published before checksums were
// introduced have no sidecar; that case is reported and skipped rather than
// failing the update.
func verifyChecksum(ctx context.Context, path, sumURL string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sumURL, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		fmt.Println("no checksum published for this release — skipping verification")
		return nil
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch checksum: server returned %s", resp.Status)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024))
	if err != nil {
		return err
	}
	want := strings.Fields(string(body)) // "hash  filename"
	if len(want) == 0 || len(want[0]) != 64 {
		return fmt.Errorf("malformed checksum file")
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	got := hex.EncodeToString(h.Sum(nil))
	if got != want[0] {
		return fmt.Errorf("mismatch: got %s, want %s", got, want[0])
	}
	fmt.Println("checksum verified ✓")
	return nil
}

// downloadFile fetches url and streams the body into dst.
func downloadFile(ctx context.Context, url string, dst *os.File) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned %s for %s", resp.Status, url)
	}

	_, err = io.Copy(dst, resp.Body)
	return err
}

// latestRemoteTag returns the highest semver tag (vX.Y.Z) from the remote.
func latestRemoteTag(ctx context.Context) (string, error) {
	out, err := exec.CommandContext(ctx,
		"git", "ls-remote", "--tags", "--sort=-v:refname", repoURL,
	).Output()
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		ref := parts[1]
		if strings.HasSuffix(ref, "^{}") {
			continue // skip peeled tag objects
		}
		tag := strings.TrimPrefix(ref, "refs/tags/")
		if !strings.HasPrefix(tag, "v") {
			continue
		}
		// Must be exactly vX.Y.Z
		seg := strings.SplitN(strings.TrimPrefix(tag, "v"), ".", 3)
		if len(seg) == 3 {
			return tag, nil
		}
	}
	return "", fmt.Errorf("no semver tags found at %s", repoURL)
}
