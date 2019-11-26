package ui

import (
	"errors"
	"fmt"
	_ "github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"time"
	"sort"
)

type activityBuffer struct {
	displayName  string
	lastActivity time.Time
}

type activityBar struct {
	bar     *tview.TextView
	buffers map[string]activityBuffer
	sorted  []activityBuffer
}

func NewActivityBar() *activityBar {
	return &activityBar{
		bar: tview.NewTextView().
			SetDynamicColors(true).
			SetWrap(false).
			SetScrollable(false),
		buffers: make(map[string]activityBuffer),
	}
}

func (b *activityBar) updateActivityBar() {
	list := []activityBuffer{}
	for _, b := range b.buffers {
		list = append(list, b)
	}

	sort.SliceStable(list, func(i, j int) bool {
		return list[i].lastActivity.After(list[j].lastActivity)
	})

	b.sorted = list

	b.bar.SetText(generateBar(list))
}

func (b *activityBar) GetLatestActiveChannel() (string, error) {
	if len(b.sorted) == 0 {
		return "", errors.New("No channels!")
	}

	return b.sorted[0].displayName, nil
}

func bufferToBarElement(buffer activityBuffer) string {
	return fmt.Sprintf("[-:blueviolet:-]%s[-:-:-] ", buffer.displayName)
}

func generateBar(buffers []activityBuffer) string {
	var result string
	for _, b := range buffers {
		result += bufferToBarElement(b)
	}

	return result
}

func (b *activityBar) RegisterActivity(buffer string, view *View) {
	view.app.QueueUpdate(func() {
		b.buffers[buffer] = activityBuffer{
			displayName:  buffer,
			lastActivity: time.Now(),
		}

		b.updateActivityBar()
	})
}