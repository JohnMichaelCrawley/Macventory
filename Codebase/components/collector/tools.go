package collector

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"macventory/components/command"
	"macventory/components/model"
)

// Function: Collect Editor Extensions
func collectEditorExtensions(ctx context.Context, cfg model.Config) model.Section {
	runner := command.Runner{Timeout: cfg.CommandTimeout}
	return commandSection(60, "Editor extensions", "Installed extensions exposed by supported CLIs", []command.Result{runner.Run(ctx, "code", "--list-extensions", "--show-versions"), runner.Run(ctx, "cursor", "--list-extensions", "--show-versions")})
}

// Function: Collect Containers
func collectContainers(ctx context.Context, cfg model.Config) model.Section {
	runner := command.Runner{Timeout: cfg.CommandTimeout}
	return commandSection(70, "Containers", "Docker client, images, containers and volume names (volume contents are not read)", []command.Result{
		runner.Run(ctx, "docker", "version", "--format", "{{.Client.Version}}"),
		runner.Run(ctx, "docker", "image", "ls", "--format", "{{.Repository}}:{{.Tag}} ({{.ID}})"),
		runner.Run(ctx, "docker", "container", "ls", "--all", "--format", "{{.Names}}| {{.Image}} | {{.Status}} "),
		runner.Run(ctx, "docker", "volume", "ls", "--format", "{{.Name}}"),
	})
}

// Function: Collect Developer Tools
func collectDeveloperTools(ctx context.Context, cfg model.Config) model.Section {
	runner := command.Runner{Timeout: cfg.CommandTimeout}
	return commandSection(80, "Developer Tools", "Selected toolchains and versions", []command.Result{
		runner.Run(ctx, "xcode-select", "-p"), runner.Run(ctx, "xcodebuild", "-version"),
		runner.Run(ctx, "xcrun", "metal", "--version"), runner.Run(ctx, "git", "--version"),
		runner.Run(ctx, "java", "-version"), runner.Run(ctx, "python3", "--version"),
		runner.Run(ctx, "terraform", "version"), runner.Run(ctx, "gcloud", "version"),
		runner.Run(ctx, "aws", "--version"),
	})
}

// Function: Collect Executables
func collectExecutables(_ context.Context, _ model.Config) model.Section {
	home, _ := os.UserHomeDir()
	dirs := []string{"/opt/homebrew/bin", "/opt/homebrew/sbin", "/usr/local/bin", "/usr/local/sbin", filepath.Join(home, ".local", "bin"), filepath.Join(home, ".cargo", "bin"), filepath.Join(home, "go", "bin")}
	seen := make(map[string]bool)
	var rows []string
	for _, dir := range dirs {
		if seen[dir] {
			continue
		}
		seen[dir] = true
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			info, err := entry.Info()
			if err != nil || entry.IsDir() || info.Mode()&0o111 == 0 {
				continue
			}
			path, target := filepath.Join(dir, entry.Name()), ""
			if info.Mode()&os.ModeSymlink != 0 {
				if resolved, err := filepath.EvalSymlinks(path); err == nil {
					target = resolved
				}
			}
			rows = append(rows, fmt.Sprintf("| `%s` | `%s` |", escapeTable(path), escapeTable(target)))
		}
	}
	sort.Strings(rows)
	return model.Section{Order: 90, Title: "User-installed executables", Summary: fmt.Sprintf("%d executables found in common non-system locations", len(rows)), Body: "| Executable | Symlink target |\n|---|---|\n" + strings.Join(rows, "\n"), Status: "ok"}
}
