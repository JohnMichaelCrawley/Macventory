package collector

import (
	"context"
	"os/exec"

	"macventory/components/command"
	"macventory/components/model"
)

// collectHomebrew inventories software managed by Homebrew.
func collectHomebrew(
	ctx context.Context,
	cfg model.Config,
) model.Section {
	if _, err := exec.LookPath("brew"); err != nil {
		return unavailableSection(
			30,
			"Homebrew",
			"Homebrew is not installed or is not on PATH.",
		)
	}

	runner := command.Runner{Timeout: cfg.CommandTimeout}

	return commandSection(
		30,
		"Homebrew",
		"Formulae, casks, top-level packages, taps, services and restorable Brewfile",
		[]command.Result{
			runner.Run(ctx, "brew", "--version"),
			runner.Run(ctx, "brew", "list", "--formula", "--versions"),
			runner.Run(ctx, "brew", "list", "--cask", "--versions"),
			runner.Run(ctx, "brew", "leaves"),
			runner.Run(ctx, "brew", "tap"),
			runner.Run(ctx, "brew", "services", "list"),
			runner.Run(
				ctx,
				"brew",
				"bundle",
				"dump",
				"--file=/dev/stdout",
				"--force",
			),
		},
	)
}

// collectMAS inventories applications reported by the mas CLI.
func collectMAS(
	ctx context.Context,
	cfg model.Config,
) model.Section {
	runner := command.Runner{Timeout: cfg.CommandTimeout}

	return commandSection(
		40,
		"Mac App Store",
		"Applications associated with the current App Store account",
		[]command.Result{
			runner.Run(ctx, "mas", "list"),
		},
	)
}

// collectLanguagePackages inventories globally installed language packages.
func collectLanguagePackages(
	ctx context.Context,
	cfg model.Config,
) model.Section {
	runner := command.Runner{Timeout: cfg.CommandTimeout}

	commands := []struct {
		name string
		args []string
	}{
		{name: "npm", args: []string{"list", "--global", "--depth=0"}},
		{name: "pnpm", args: []string{"list", "--global", "--depth=0"}},
		{name: "yarn", args: []string{"global", "list", "--depth=0"}},
		{name: "pipx", args: []string{"list"}},
		{name: "uv", args: []string{"tool", "list"}},
		{name: "gem", args: []string{"list", "--local"}},
		{name: "cargo", args: []string{"install", "--list"}},
		{name: "go", args: []string{"version"}},
	}

	results := make([]command.Result, 0, len(commands))

	for _, item := range commands {
		results = append(
			results,
			runner.Run(ctx, item.name, item.args...),
		)
	}

	return commandSection(
		50,
		"Language package managers",
		"Globally installed packages; project-local dependencies are intentionally excluded",
		results,
	)
}
