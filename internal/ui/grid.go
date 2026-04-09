package ui

import (
	"fmt"
	"sync"
	"time"

	"github.com/declan-whiting/vaulty/internal/events"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (ui *Ui) CreateGrid() *Ui {
	w := ui.KeyvaultView.Width()
	ui.Grid = tview.NewGrid().SetRows(5, 3, 3).SetColumns(w, 0)
	ui.Grid.AddItem(ui.ControlsView, 0, 0, 1, 2, 0, 0, false)
	ui.Grid.AddItem(ui.StatusView, 0, 2, 1, 1, 0, 0, false)
	ui.Grid.AddItem(ui.KeyvaultView, 1, 0, 3, 1, 0, 0, false)
	ui.Grid.AddItem(ui.SecretsView, 1, 1, 3, 2, 0, 0, false)
	ui.AddStatusControls()
	return ui
}

func (ui *Ui) HideSearch() {
	ui.Grid.RemoveItem(ui.SearchView)
	ui.Grid.RemoveItem(ui.KeyvaultView)
	ui.Grid.RemoveItem(ui.SecretsView)
	ui.Grid.AddItem(ui.KeyvaultView, 1, 0, 3, 1, 0, 0, false)
	ui.Grid.AddItem(ui.SecretsView, 1, 1, 3, 2, 0, 0, false)
}

func (ui *Ui) ShowSearch() {
	ui.App.SetFocus(ui.SearchView)
	ui.Grid.RemoveItem(ui.KeyvaultView)
	ui.Grid.RemoveItem(ui.SecretsView)
	ui.Grid.AddItem(ui.SearchView, 1, 0, 1, 3, 0, 0, false)
	ui.Grid.AddItem(ui.KeyvaultView, 2, 0, 2, 1, 0, 0, false)
	ui.Grid.AddItem(ui.SecretsView, 2, 1, 2, 2, 0, 0, false)
}

func (ui *Ui) AddStatusControls() *Ui {
	ui.Grid.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlR {
			start := time.Now()
			ui.Events.NewEvent("\U0001F510 Synchronise Started", "")

			go func() {
				config := ui.Services.ConfigrationService.GetConfiguration()
				var wg sync.WaitGroup

				for _, v := range config.Keyvaults {
					wg.Add(1)
					go func(name, subscriptionID string) {
						defer wg.Done()
						ui.Services.AzureService.AzGetSecrets(name, subscriptionID)
					}(v.Name, v.SubscriptionId)
				}

				wg.Wait()

				ui.Services.CacheService.WriteLastSync([]byte(fmt.Sprintf("Last Sync: %s", time.Now().Format(time.ANSIC))))

				ui.App.QueueUpdateDraw(func() {
					if ui.SecretsView.CurrentKeyvaultName != "" {
						ui.SecretsView.NotifyUpdate(ui.SearchView.GetText())
					}
					events.TimedEventLog(start, "\U0001F510 Synchronise Finished", *ui.Events)
				})
			}()

			return nil

		}

		return event
	})

	return ui
}
