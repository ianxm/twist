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
				Name:         "Paired Ports",
				CreatesModal: true,
				IsEnabled:    isConnectedCheck,
				HandleAction: handlePairedPorts,
			},
			{
				Name:         "Nearby Ports",
				CreatesModal: true,
				IsEnabled:    isConnectedCheck,
				HandleAction: handleNearbyPorts,
			},
			{
				Name:         "Special Ports",
				CreatesModal: true,
				IsEnabled:    isConnectedCheck,
				HandleAction: handleSpecialPorts,
			},
			{
				Name:         "Bubbles",
				CreatesModal: true,
				IsEnabled:    isConnectedCheck,
				HandleAction: handleBubbles,
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
		text = " No bubbles (size ≥ 2) found in explored sectors.\n"
	} else {
		text = fmt.Sprintf(" Found %d bubble(s)\n\n", len(filtered))
		text += " ┌─────────┬──────┐\n"
		text += " │ Gateway │ Size │\n"
		text += " ├─────────┼──────┤\n"
		for _, b := range filtered {
			text += fmt.Sprintf(" │ %7d │ %4d │\n", b.GatewaySector, b.Size)
		}
		text += " └─────────┴──────┘"
	}

	contentHeight := len(filtered) + 11
	termHeight := app.GetTerminalHeight()
	maxHeight := termHeight - 6
	if contentHeight > maxHeight {
		contentHeight = maxHeight
	}
	if contentHeight < 8 {
		contentHeight = 8
	}

	app.ShowScrollableModal("Bubbles", text, 24, contentHeight)
	return nil
}

func handlePairedPorts(app AppInterface) error {
	proxyAPI := app.GetProxyAPI()
	if proxyAPI == nil {
		app.ShowModal("Paired Ports", "Not connected.", []string{"Close"},
			func(buttonIndex int, buttonLabel string) { app.CloseModal() })
		return nil
	}

	currentSector, err := proxyAPI.GetCurrentSector()
	if err != nil || currentSector == 0 {
		app.ShowModal("Paired Ports", "Current sector unknown. Move to a sector first.", []string{"Close"},
			func(buttonIndex int, buttonLabel string) { app.CloseModal() })
		return nil
	}

	sectors, err := proxyAPI.GetSectorAnalysisData()
	if err != nil {
		app.ShowModal("Paired Ports", fmt.Sprintf("Error loading sector data: %v", err), []string{"Close"},
			func(buttonIndex int, buttonLabel string) { app.CloseModal() })
		return nil
	}

	pairs := analysis.FindPairs(sectors, currentSector)

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Distance < pairs[j].Distance
	})

	var text string
	if len(pairs) == 0 {
		text = fmt.Sprintf(" No paired ports found within 10 hops of sector %d.\n", currentSector)
	} else {
		text = fmt.Sprintf(" %d pair(s) from sector %d\n\n", len(pairs), currentSector)
		text += " ┌─────────┬───────┬─────────┬───────┬──────┐\n"
		text += " │ Sector1 │ Class │ Sector2 │ Class │ Dist │\n"
		text += " ├─────────┼───────┼─────────┼───────┼──────┤\n"
		for _, p := range pairs {
			c1 := api.PortClass(p.Class1).String()
			c2 := api.PortClass(p.Class2).String()
			text += fmt.Sprintf(" │ %7d │  %s  │ %7d │  %s  │ %4d │\n",
				p.Sector1, c1, p.Sector2, c2, p.Distance)
		}
		text += " └─────────┴───────┴─────────┴───────┴──────┘"
	}

	contentHeight := len(pairs) + 11
	termHeight := app.GetTerminalHeight()
	maxHeight := termHeight - 6
	if contentHeight > maxHeight {
		contentHeight = maxHeight
	}
	if contentHeight < 8 {
		contentHeight = 8
	}

	app.ShowScrollableModal("Paired Ports", text, 52, contentHeight)
	return nil
}

func handleSpecialPorts(app AppInterface) error {
	proxyAPI := app.GetProxyAPI()
	if proxyAPI == nil {
		app.ShowModal("Special Ports", "Not connected.", []string{"Close"},
			func(buttonIndex int, buttonLabel string) { app.CloseModal() })
		return nil
	}

	ports, err := proxyAPI.GetSpecialPorts()
	if err != nil {
		app.ShowModal("Special Ports", fmt.Sprintf("Error: %v", err), []string{"Close"},
			func(buttonIndex int, buttonLabel string) { app.CloseModal() })
		return nil
	}

	if len(ports) == 0 {
		app.ShowModal("Special Ports", "No special ports found.", []string{"Close"},
			func(buttonIndex int, buttonLabel string) { app.CloseModal() })
		return nil
	}

	text := " ┌────────┬─────────────────────┐\n"
	text += " │ Sector │ Port                │\n"
	text += " ├────────┼─────────────────────┤\n"
	for _, p := range ports {
		text += fmt.Sprintf(" │ %6s │ %-19s │\n", p.Sector, p.Name)
	}
	text += " └────────┴─────────────────────┘"

	app.ShowModal("Special Ports", text, []string{"Close"},
		func(buttonIndex int, buttonLabel string) { app.CloseModal() })
	return nil
}

func handleNearbyPorts(app AppInterface) error {
	proxyAPI := app.GetProxyAPI()
	if proxyAPI == nil {
		app.ShowModal("Nearby Ports", "Not connected.", []string{"Close"},
			func(buttonIndex int, buttonLabel string) { app.CloseModal() })
		return nil
	}

	currentSector, err := proxyAPI.GetCurrentSector()
	if err != nil || currentSector == 0 {
		app.ShowModal("Nearby Ports", "Current sector unknown.", []string{"Close"},
			func(buttonIndex int, buttonLabel string) { app.CloseModal() })
		return nil
	}

	sectors, err := proxyAPI.GetSectorAnalysisData()
	if err != nil {
		app.ShowModal("Nearby Ports", fmt.Sprintf("Error: %v", err), []string{"Close"},
			func(buttonIndex int, buttonLabel string) { app.CloseModal() })
		return nil
	}

	ports := analysis.FindNearbyPorts(sectors, currentSector)

	var text string
	if len(ports) == 0 {
		text = fmt.Sprintf(" No ports found near sector %d.\n", currentSector)
	} else {
		text = fmt.Sprintf(" Ports near sector %d\n\n", currentSector)
		text += " ┌────────┬───────┬──────┐\n"
		text += " │ Sector │ Class │ Dist │\n"
		text += " ├────────┼───────┼──────┤\n"
		for _, p := range ports {
			c := api.PortClass(p.Class).String()
			text += fmt.Sprintf(" │ %6d │  %s  │ %4d │\n", p.Sector, c, p.Distance)
		}
		text += " └────────┴───────┴──────┘"
	}

	contentHeight := len(ports) + 11
	termHeight := app.GetTerminalHeight()
	maxHeight := termHeight - 6
	if contentHeight > maxHeight {
		contentHeight = maxHeight
	}
	if contentHeight < 8 {
		contentHeight = 8
	}

	app.ShowScrollableModal("Nearby Ports", text, 30, contentHeight)
	return nil
}
