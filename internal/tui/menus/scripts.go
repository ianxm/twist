package menus

import (
	"fmt"
	"strings"
	"twist/internal/api"
	"twist/internal/log"
)

func NewScriptsMenu() Menu {
	return Menu{
		Name:     "Scripts",
		Shortcut: "Alt+R",
		Items: []MenuItem{
			{
				Name:         "List",
				CreatesModal: true,
				IsEnabled:    isConnectedCheck,
				HandleAction: handleList,
			},
			{
				Name:         "Burst",
				CreatesModal: true,
				IsEnabled:    isConnectedCheck,
				HandleAction: handleBurst,
			},
			{
				Name:         "Stop All Scripts",
				Shortcut:     "Esc",
				CreatesModal: true,
				IsEnabled:    isConnectedCheck,
				HandleAction: handleStopAllScripts,
			},
		},
	}
}

func handleList(app AppInterface) error {
	defer func() {
		if r := recover(); r != nil {
			log.Error("PANIC in handleList", "error", r)
		}
	}()

	proxyAPI := app.GetProxyAPI()
	if proxyAPI == nil {
		app.ShowModal("Scripts List",
			"Not connected to proxy. Please connect first.",
			[]string{"OK"},
			func(buttonIndex int, buttonLabel string) {
				app.CloseModal()
			})
		return nil
	}

	scripts, err := proxyAPI.GetScriptList()
	if err != nil {
		app.ShowModal("Scripts List Error",
			fmt.Sprintf("Error getting script list: %v", err),
			[]string{"OK"},
			func(buttonIndex int, buttonLabel string) {
				app.CloseModal()
			})
		return nil
	}

	var listText strings.Builder
	if len(scripts) == 0 {
		listText.WriteString("No scripts loaded.\n\n")
	} else {
		listText.WriteString(fmt.Sprintf("Loaded Scripts (%d):\n\n", len(scripts)))
		listText.WriteString(buildReasonableTable(scripts))
	}

	app.ShowModal("Scripts List", listText.String(), []string{"Close"},
		func(buttonIndex int, buttonLabel string) {
			app.CloseModal()
		})

	return nil
}

func handleBurst(app AppInterface) error {
	proxyAPI := app.GetProxyAPI()
	if proxyAPI == nil {
		app.ShowModal("Burst Command",
			"Not connected to proxy. Please connect first.",
			[]string{"OK"},
			func(buttonIndex int, buttonLabel string) {
				app.CloseModal()
			})
		return nil
	}

	app.ShowBurstDialog(
		func(burstText string) {
			err := proxyAPI.SendBurstCommand(burstText)
			if err != nil {
				app.ShowModal("Burst Command Error",
					fmt.Sprintf("Error sending burst command: %v", err),
					[]string{"OK"},
					func(buttonIndex int, buttonLabel string) {
						app.CloseModal()
					})
			} else {
				app.CloseModal()
			}
		},
		func() {
			app.CloseModal()
		},
	)
	return nil
}

func buildReasonableTable(scripts []api.ScriptInfo) string {
	var table strings.Builder

	table.WriteString("┌────┬─────────────────┬─────────────────┬────────┐\n")
	table.WriteString("│ ID │ Name            │ File            │ Active │\n")
	table.WriteString("├────┼─────────────────┼─────────────────┼────────┤\n")

	for i, script := range scripts {
		status := ""
		if script.IsActive {
			status = "✓"
		}

		name := script.Name
		if len(name) > 15 {
			name = name[:12] + "..."
		}

		filename := script.Filename
		if strings.Contains(filename, "/") {
			parts := strings.Split(filename, "/")
			filename = parts[len(parts)-1]
		}
		if len(filename) > 15 {
			filename = filename[:12] + "..."
		}

		table.WriteString(fmt.Sprintf("│ %2d │ %-15s │ %-15s │   %s    │\n",
			i+1, name, filename, status))
	}

	table.WriteString("└────┴─────────────────┴─────────────────┴────────┘\n")
	return table.String()
}

func handleStopAllScripts(app AppInterface) error {
	defer func() {
		if r := recover(); r != nil {
			log.Error("PANIC in handleStopAllScripts", "error", r)
		}
	}()

	proxyAPI := app.GetProxyAPI()
	if proxyAPI == nil {
		app.ShowModal("Stop All Scripts",
			"Not connected to proxy. Please connect first.",
			[]string{"OK"},
			func(buttonIndex int, buttonLabel string) {
				app.CloseModal()
			})
		return nil
	}

	err := proxyAPI.StopAllScripts()
	if err != nil {
		app.ShowModal("Stop All Scripts",
			fmt.Sprintf("Error stopping scripts: %v", err),
			[]string{"OK"},
			func(buttonIndex int, buttonLabel string) {
				app.CloseModal()
			})
		return nil
	}

	app.ShowModal("Stop All Scripts",
		"Stopping all scripts...",
		[]string{"OK"},
		func(buttonIndex int, buttonLabel string) {
			app.CloseModal()
		})

	return nil
}
