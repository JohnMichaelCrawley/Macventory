package collector

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"macventory/components/model"
)

// Function: All
func All() []model.Collector {
	return []model.Collector{
		{Order: 10, Title: "System", Run: collectSystem},
		{Order: 20, Title: "Applications", Run: collectApplications},
		{Order: 30, Title: "Homebrew", Run: collectHomebrew},
		{Order: 40, Title: "Mac App Store", Run: collectMAS},
		{Order: 50, Title: "Language package managers", Run: collectLanguagePackages},
		{Order: 60, Title: "Editor extensions", Run: collectEditorExtensions},
		{Order: 70, Title: "Containers", Run: collectContainers},
		{Order: 80, Title: "Developer Tools", Run: collectDeveloperTools},
		{Order: 90, Title: "User-installed executables", Run: collectExecutables},
	}
}

// Function: Collect All
func CollectAll(
	ctx context.Context,
	cfg model.Config,
	collectors []model.Collector,
) []model.Section {
	sections := make([]model.Section, len(collectors))

	var wg sync.WaitGroup

	for i, item := range collectors {
		i, item := i, item
		wg.Add(1)

		go func() {
			defer wg.Done()

			defer func() {
				if recovered := recover(); recovered != nil {
					sections[i] = model.Section{
						Order:  item.Order,
						Title:  item.Title,
						Status: "error",
						Body: fmt.Sprintf(
							"Collector failed safely: %v",
							recovered,
						),
					}
				}
			}()

			sections[i] = item.Run(ctx, cfg)
		}()
	}

	wg.Wait()

	sort.Slice(sections, func(i, j int) bool {
		return sections[i].Order < sections[j].Order
	})

	return sections
}
