package keyvault

import (
    "github.com/declan-whiting/vaulty/internal/models"
    "github.com/gdamore/tcell/v2"
    "github.com/rivo/tview"
)

type CacheService interface {
    ReadKeyvaults() []models.KeyvaultModel
}

type ConfigrationService interface {
    GetConfiguration() models.ConfigurationList
}

type CurrentKeyvaultWatcher interface {
    CurrentKeyvaultUpdated(name, subscription string)
}

type Themer interface {
    GetColor(color string) tcell.Color
}

type KeyvaultView struct {
    *tview.List
    Conf                    ConfigrationService
    Cache                   CacheService
    QuitHandler             func()
    SearchHandler           func()
    SelectedHandler         func()
    CurrentKeyvaultWatchers []CurrentKeyvaultWatcher
}

func NewKeyvaultView(cache CacheService, conf ConfigrationService, quiter, searcher, selecter func(), theme Themer) *KeyvaultView {
    keyvaultView := &KeyvaultView{
        Cache:           cache,
        Conf:            conf,
        QuitHandler:     quiter,
        SearchHandler:   searcher,
        SelectedHandler: selecter,
    }
    keyvaultView.List = tview.NewList()
    keyvaultView.SetTitle("Keyvaults")
    keyvaultView.SetBorder(true)
    keyvaultView.ShowSecondaryText(false)
    keyvaultView.SetBorderPadding(0, 0, 1, 0)

    keyvaultView.SetSelectedBackgroundColor(theme.GetColor("background"))
    keyvaultView.SetSelectedTextColor(theme.GetColor("pink"))

    for i, v := range cache.ReadKeyvaults() {
        shortcut := rune(0)
        if i < 9 {
            shortcut = rune('1' + i)
        }
        keyvaultView.AddItem(v.Name, v.SubscriptionId, shortcut, nil)
    }

    keyvaultView.AddKeyvaultViewControls()
    keyvaultView.KeyvaultSelectionChanged()

    return keyvaultView
}

func (kv *KeyvaultView) AddKeyvaultViewControls() {
    kv.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
        if event.Rune() == 'q' {
            kv.QuitHandler()
            return tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone)
        }
        if event.Rune() == '/' {
            kv.SearchHandler()
        }
        if event.Key() == tcell.KeyEnter {
            kv.SelectedHandler()
            return nil
        }
        if event.Rune() >= '1' && event.Rune() <= '9' {
            idx := int(event.Rune() - '1')
            if idx < kv.GetItemCount() {
                kv.SetCurrentItem(idx)
                kv.SelectedHandler()
            }
            return nil
        }
        return event
    })

}

func (kv *KeyvaultView) SetInitialKeyvault() {
    vault := kv.Conf.GetConfiguration().Keyvaults[0]

    for _, v := range kv.CurrentKeyvaultWatchers {
        v.CurrentKeyvaultUpdated(vault.Name, vault.SubscriptionId)
    }
}

func (kv *KeyvaultView) KeyvaultSelectionChanged() {
    kv.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
        for _, v := range kv.CurrentKeyvaultWatchers {
            v.CurrentKeyvaultUpdated(mainText, secondaryText)
        }
    })

}

func (kv *KeyvaultView) AddCurrentKeyvaultWatcher(watcher CurrentKeyvaultWatcher) {
    kv.CurrentKeyvaultWatchers = append(kv.CurrentKeyvaultWatchers, watcher)
}

// Width returns the column width needed to display the longest vault name in full.
// Accounts for border (2), left padding (1), and shortcut prefix e.g. "1. " (3).
func (kv *KeyvaultView) Width() int {
    const overhead = 8 // left border + left padding + "N. " prefix + right border
    const minWidth = 20
    max := minWidth
    for i := 0; i < kv.GetItemCount(); i++ {
        main, _ := kv.GetItemText(i)
        if w := len(main) + overhead; w > max {
            max = w
        }
    }
    return max
}
