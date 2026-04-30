package fw

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
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

			// Clone the tagged release into a temp dir.
			tmpDir, err := os.MkdirTemp("", "fw-update-*")
			if err != nil {
				return err
			}
			defer os.RemoveAll(tmpDir)

			fmt.Printf("cloning %s...\n", latest)
			if err := runGitCmd(ctx, "", "clone", "-c", "advice.detachedHead=false",
				"--depth=1", "--branch", latest, repoURL, tmpDir); err != nil {
				return fmt.Errorf("git clone: %w", err)
			}

			// Build with the version injected.
			newBin := filepath.Join(tmpDir, "fw")
			ver := strings.TrimPrefix(latest, "v")
			ldflags := fmt.Sprintf("-X github.com/mahi160/flowd/internal/fw.Version=%s", ver)

			fmt.Println("building...")
			c := exec.CommandContext(ctx, "go", "build", "-ldflags", ldflags, "-o", newBin, "./cmd/fw")
			c.Dir = tmpDir
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			if err := c.Run(); err != nil {
				return fmt.Errorf("go build: %w", err)
			}

			// Replace the binary — try direct rename first, then fall back to sudo.
			if err := replaceBinary(newBin, exe); err != nil {
				// Stage the built binary in a temp location.
				fallback := filepath.Join(os.TempDir(), "fw-new")
				if renErr := os.Rename(newBin, fallback); renErr != nil {
					// Cross-device rename (tmpDir on different fs) — copy instead.
					if cpErr := copyFile(newBin, fallback, 0755); cpErr != nil {
						return fmt.Errorf("could not stage binary: %w", cpErr)
					}
				}
				fmt.Printf("\n%s is owned by root — running sudo to replace it\n", exe)
				sudo := exec.CommandContext(ctx, "sudo", "mv", fallback, exe)
				sudo.Stdin = os.Stdin
				sudo.Stdout = os.Stdout
				sudo.Stderr = os.Stderr
				if sudoErr := sudo.Run(); sudoErr != nil {
					os.Remove(fallback)
					fmt.Printf("sudo failed: %v\nrun manually:\n  sudo mv %s %s\n", sudoErr, fallback, exe)
					return nil
				}
				os.Remove(fallback) // clean up if sudo cp was used instead of mv
			}

			fmt.Printf("\nupdated to %s ✓\n", latest)
			return nil
		},
	}
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

// replaceBinary atomically replaces dst with src on the same filesystem.
func replaceBinary(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	// Write to a sibling temp file, then rename (atomic swap).
	tmp := dst + ".new"
	out, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		os.Remove(tmp)
		return err
	}
	if err := out.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, dst)
}

// runGitCmd runs git with the given args, optionally in a working directory.
func runGitCmd(ctx context.Context, dir string, args ...string) error {
	c := exec.CommandContext(ctx, "git", args...)
	if dir != "" {
		c.Dir = dir
	}
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}
