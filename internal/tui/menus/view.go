package menus

import (
	"fmt"
	"sort"
	"twist/internal/api"
	"twist/internal/log"
	"twist/internal/tui/analysis"
)

func NewViewMenu() Menu {
	return Menu{
		Name:     "View",
		Shortcut: "Alt+V",
		Items: []MenuItem{
			{
				Name:         "Panels",
				IsEnabled:    isConnectedCheck,
				HandleAction: handlePanels,
			},
			{
				Name:         "Bubbles",
				CreatesModal: true,
				IsEnabled:    isConnectedCheck,
				HandleAction: handleBubbles,
			},
			{
				Name:         "Paired Ports",
				CreatesModal: true,
				IsEnabled:    isConnectedCheck,
				HandleAction: handlePairedPorts,
			},
		},
	}
}

func handlePanels(app AppInterface) error {
	if app.GetPanelsVisible() {
		app.HidePanels()
		log.Info("ViewMenu: Hiding panels")
	} else {
		app.ShowPanels()
		log.Info("ViewMenu: Showing panels")
	}
	return nil
}

func handleBubbles(app AppInterface) error {
	proxyAPI := app.GetProxyAPI()
	if proxyAPI == nil {
		app.ShowModal("Bubbles", "Not connected.", []string{"Close"},
			func(buttonIndex int, buttonLabel string) { app.CloseModal() })
		return nil
	}

	sectors, err := proxyAPI.GetSectorAnalysisData()
	if err != nil {
		app.ShowModal("Bubbles", fmt.Sprintf("Error loading sector data: %v", err), []string{"Close"},
			func(buttonIndex int, buttonLabel string) { app.CloseModal() })
		return nil
	}

	bubbles := analysis.FindBubbles(sectors)

	// Filter to size >= 2 and sort largest first
	var filtered []api.BubbleInfo
	for _, b := range bubbles {
		if b.Size >= 2 {
			filtered = append(filtered, b)
		}
	}
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Size > filtered[j].Size
	})

	var text string
	if len(filtered) == 0 {
		text = "No bubbles (size ≥ 2) found in explored sectors."
	} else {
		text = fmt.Sprintf(" Found %d bubble(s)\n\n", len(filtered))
		text += " ┌─────────┬──────┐\n"
		text += " │ Gateway │ Size │\n"
		text += " ├─────────┼──────┤\n"
		for _, b := range filtered {
			text += fmt.Sprintf(" │ %7d │ %4d │\n", b.GatewaySector, b.Size)
		}
		text += " └─────────┴──────┘\n"
	}

	// Content height: header (2) + table header (3) + rows + table footer (1) + padding (2)
	contentHeight := len(filtered) + 8
	termHeight := app.GetTerminalHeight()
	maxHeight := termHeight - 6 // Leave room for top/bottom margins
	if contentHeight > maxHeight {
		contentHeight = maxHeight
	}
	if contentHeight < 8 {
		contentHeight = 8
	}

	app.ShowScrollableModal("Bubbles", text, contentHeight)
	return nil
}

func handlePairedPorts(app AppInterface) error {
	app.ShowModal("Paired Ports", "Paired Ports\n\npair 1", []string{"Close"},
		func(buttonIndex int, buttonLabel string) {
			app.CloseModal()
		})
	return nil
}
