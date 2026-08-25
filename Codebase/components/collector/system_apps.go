package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"macventory/components/command"
	"macventory/components/model"
)

func collectSystem(ctx context.Context, cfg model.Config) model.Section {
	runner := command.Runner{Timeout: cfg.CommandTimeout}
	results := []command.Result{
		runner.Run(ctx, "sw_vers"),
		runner.Run(ctx, "uname", "-m"),
		runner.Run(ctx, "sysctl", "-n", "machdep.cpu.brand_string"),
		runner.Run(ctx, "sysctl", "-n", "hw.memsize"),
		runner.Run(ctx, "system_profiler", "SPHardwareDataType", "-detailLevel", "mini"),
	}
	for i := range results {
		results[i].Output = redactLines(results[i].Output, []string{"Serial Number", "Hardware UUID", "Provisioning UDID"})
	}
	return commandSection(10, "System", "macOS and hardware overview (unique device identifiers redacted)", results)
}

type profilerPayload struct {
	Applications []profilerApp `json:"SPApplicationsDataType"`
}

type profilerApp struct {
	Name    string `json:"_name"`
	Version string `json:"version"`
	Path    string `json:"path"`
	Source  string `json:"obtained_from"`
	Arch    string `json:"arch_kind"`
}

func collectApplications(ctx context.Context, cfg model.Config) model.Section {
	runner := command.Runner{Timeout: cfg.CommandTimeout * 3}
	result := runner.Run(ctx, "system_profiler", "SPApplicationsDataType", "-json", "-detailLevel", "full")
	section := model.Section{Order: 20, Title: "Applications", Status: "ok"}
	if !result.Available {
		section.Status, section.Body = "unavailable", "`system_profiler` is unavailable."
		return section
	}
	if result.Err != nil {
		section.Status, section.Body = "error", fmt.Sprintf("Application scan failed: `%s`", escapeInline(result.Err.Error()))
		return section
	}
	var payload profilerPayload
	if err := json.Unmarshal([]byte(result.Output), &payload); err != nil {
		section.Status, section.Body = "error", fmt.Sprintf("Application data could not be parsed: `%s`", escapeInline(err.Error()))
		return section
	}

	apps := make([]profilerApp, 0, len(payload.Applications))
	for _, app := range payload.Applications {
		if !cfg.IncludeSystemApps && strings.HasPrefix(app.Path, "/System/") {
			continue
		}
		apps = append(apps, app)
	}
	sort.Slice(apps, func(i, j int) bool {
		if strings.EqualFold(apps[i].Name, apps[j].Name) {
			return apps[i].Path < apps[j].Path
		}
		return strings.ToLower(apps[i].Name) < strings.ToLower(apps[j].Name)
	})

	var body strings.Builder
	body.WriteString("| Application | Version | Architecture | Source | Path |\n|---|---|---|---|---|\n")
	for _, app := range apps {
		fmt.Fprintf(&body, "| %s | %s | %s | %s | `%s` |\n", cell(app.Name), cell(app.Version), cell(app.Arch), cell(app.Source), strings.ReplaceAll(app.Path, "`", "\\`"))
	}
	section.Summary = fmt.Sprintf("%d applications detected", len(apps))
	section.Body = body.String()
	return section
}
