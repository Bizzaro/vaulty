package secretDetails

import (
	"encoding/json"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type SecretDetailsView struct {
	*tview.TextView
	secretValue string
}

type azSecretResponse struct {
	Value string `json:"value"`
}

func CreateSecretsDetailView(title, content string) *SecretDetailsView {
	details := &SecretDetailsView{}
	details.TextView = tview.NewTextView()
	details.SetBorder(true)
	details.SetTitle(title)
	details.SetText(content)

	var resp azSecretResponse
	if err := json.Unmarshal([]byte(content), &resp); err == nil {
		// Replace literal \n sequences with real newlines (e.g. YAML stored with escaped newlines)
		details.secretValue = strings.ReplaceAll(resp.Value, `\n`, "\n")
	}

	return details
}

func (sdv *SecretDetailsView) UpdateContent(content string) {
	sdv.SetText(content)
	var resp azSecretResponse
	if err := json.Unmarshal([]byte(content), &resp); err == nil {
		sdv.secretValue = strings.ReplaceAll(resp.Value, `\n`, "\n")
	}
}

func (sdv *SecretDetailsView) AddControls(closer func(tv *tview.TextView), onCopy func(), onRefresh func()) {
	sdv.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'b' || event.Key() == tcell.KeyEscape {
			closer(sdv.TextView)
			return tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone)
		}
		if event.Rune() == 'y' {
			if sdv.secretValue != "" {
				_ = clipboard.WriteAll(sdv.secretValue)
				if onCopy != nil {
					onCopy()
				}
			}
			return nil
		}
		if event.Rune() == 'r' {
			if onRefresh != nil {
				onRefresh()
			}
			return nil
		}
		return event
	})
}
