package fw

import (
	"context"
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

			tmpFile, err := os.CreateTemp("", "fw-update-*")
			if err != nil {
				return err
			}
			tmpPath := tmpFile.Name()
			defer os.Remove(tmpPath)

			if err := downloadFile(ctx, downloadURL, tmpFile); err != nil {
				tmpFile.Close()
				return fmt.Errorf("download binary: %w", err)
			}
			tmpFile.Close()

			if err := os.Chmod(tmpPath, 0755); err != nil {
				return fmt.Errorf("chmod: %w", err)
			}

			// Replace the running binary.
			if err := replaceBinary(tmpPath, exe); err != nil {
				fmt.Printf("\n%s is not writable — running sudo cp\n", exe)
				sudo := exec.CommandContext(ctx, "sudo", "cp", tmpPath, exe)
				sudo.Stdin = os.Stdin
				sudo.Stdout = os.Stdout
				sudo.Stderr = os.Stderr
				if sudoErr := sudo.Run(); sudoErr != nil {
					fmt.Printf("sudo failed: %v\nrun manually:\n  sudo cp %s %s\n", sudoErr, tmpPath, exe)
					return nil
				}
			}

			fmt.Printf("\nupdated to %s ✓\n", latest)
			return nil
		},
	}
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

// copyFile copies src to dst with the given permissions.
func copyFile(src, dst string, perm os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		os.Remove(dst)
		return err
	}
	return out.Close()
}

// replaceBinary overwrites dst with src, the same way `cp src dst` does.
func replaceBinary(src, dst string) error {
	return copyFile(src, dst, 0755)
}
