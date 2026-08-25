package report

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"macventory/components/model"
)

// Function: Latest
func Latest(directory string) (string, os.FileInfo, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return "", nil, err
	}

	var latestPath string
	var latestInfo os.FileInfo
	for _, entry := range entries {
		if entry.IsDir() ||
			!strings.HasPrefix(entry.Name(), "macventory-") ||
			filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		info, err := entry.Info()
		if err == nil && (latestInfo == nil || info.ModTime().After(latestInfo.ModTime())) {
			latestPath, latestInfo = filepath.Join(directory, entry.Name()), info
		}
	}
	if latestInfo == nil {
		return "", nil, fmt.Errorf("no macventory reports found in %s", directory)
	}
	return latestPath, latestInfo, nil
}

// Function: Render
func Render(sections []model.Section, version string) string {
	generated := time.Now().Format(time.RFC3339)
	host, _ := os.Hostname()
	var body strings.Builder
	body.WriteString("# Mac Inventory\n\n")
	body.WriteString("> Generated locally by Macventory " + version + " on " + generated + ".\n\n")
	body.WriteString("## Scan Summary\n\n| Field | Value | \n|---|---|\n")
	fmt.Fprintf(&body, "| Computer name | %s |\n", cell(host))
	fmt.Fprintf(&body, "| Generated | %s | \n", cell(generated))
	fmt.Fprintf(&body, "| Macventory | %s |\n", version)
	body.WriteString("\n| Section | Status | Summary |\n")
	body.WriteString("|---|---|---|\n")

	for _, section := range sections {
		fmt.Fprintf(
			&body,
			"| [%s](#%s) | %s | %s |\n",
			cell(section.Title),
			anchor(section.Title),
			cell(section.Status),
			cell(section.Summary),
		)
	}

	for _, section := range sections {
		fmt.Fprintf(&body, "\n## %s\n\n", section.Title)

		if section.Summary != "" {
			body.WriteString(section.Summary + ".\n\n")
		}

		body.WriteString(strings.TrimSpace(section.Body))
		body.WriteString("\n")
	}
	body.WriteString(`
## Restore notes

- Reinstall Homebrew entries using the embedded Brewfile as a reference.
- Reinstall Mac App Store applications after signing in with the same Apple Account.
- Download manually installed applications from their official publishers.
- Restore personal data and application settings from a separately verified backup.
- Review the report before sharing it.
`)

	return body.String()

}

// Function: Cell
func cell(value string) string {
	if value == "" {
		return "-"
	}
	return strings.NewReplacer("|", "\\|", "\n", " ", "\r", "").Replace(value)
}

// Function: Anchor
func anchor(value string) string {
	value = strings.ToLower(value)

	var body strings.Builder

	for _, char := range value {
		if (char >= 'a' && char <= 'z') ||
			(char >= '0' && char <= '9') ||
			char == ' ' ||
			char == '-' {
			body.WriteRune(char)
		}
	}

	return strings.ReplaceAll(
		strings.TrimSpace(body.String()),
		" ",
		"-",
	)
}
