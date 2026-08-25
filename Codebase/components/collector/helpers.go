package collector

import (
	"strings"

	"macventory/components/command"
	"macventory/components/model"
)

// Function Command Section
func commandSection(order int, title, summary string, results []command.Result) model.Section {
	var body strings.Builder
	available, errorsCount := 0, 0
	for _, result := range results {
		body.WriteString("### `" + strings.ReplaceAll(result.Command, "`", "\\`") + "`\n\n")
		switch {
		case !result.Available:
			body.WriteString("_Not installed or not on PATH._\n\n")
		case result.Err != nil:
			available++
			errorsCount++
			body.WriteString("_Command did not complete successfully: " + escapeMarkdown(result.Err.Error()) + "_\n\n")
			if result.Output != "" {
				body.WriteString(fenced(result.Output))
			}
		default:
			available++
			if result.Output == "" {
				body.WriteString("_(No output)_\n\n")
			} else {
				body.WriteString(fenced(result.Output))
			}
		}
	}
	status := "ok"
	if available == 0 {
		status = "unavailable"
	} else if errorsCount > 0 {
		status = "partial"
	}
	return model.Section{Order: order, Title: title, Summary: summary, Body: body.String(), Status: status}
}

// Function: Unavailable Section
func unavailableSection(order int, title, message string) model.Section {
	return model.Section{Order: order, Title: title, Body: message, Status: "unavailable"}
}

// Function: Fenced
func fenced(value string) string {
	value = strings.TrimSpace(value)
	fence := "```"
	if strings.Contains(value, fence) {
		fence = "````"
	}
	return fence + "text\n" + value + "\n" + fence + "\n\n"
}

// Function: Redact Lines
func redactLines(value string, keys []string) string {
	lines := strings.Split(value, "\n")
	for i, line := range lines {
		for _, key := range keys {
			if strings.Contains(line, key+":") {
				lines[i] = line[:strings.Index(line, key)] +
					key +
					": [redacted]"
				break
			}

		}
	}
	return strings.Join(lines, "\n")
}

// Function: Cell
func cell(value string) string {
	if value == "" {
		return "—"
	}
	return strings.NewReplacer("|", "\\|", "\n", " ", "\r", "").Replace(value)
}

// Function: Escape Table
func escapeTable(value string) string { return strings.ReplaceAll(value, "|", "\\|") }

// Function:  Escape In Line
func escapeInline(value string) string { return strings.ReplaceAll(value, "`", "'") }

// Function: Escape MD
func escapeMarkdown(value string) string {
	return strings.NewReplacer("*", "\\*", "_", "\\_", "`", "'").Replace(value)
}
