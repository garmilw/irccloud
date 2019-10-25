package ui

import (
	"fmt"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/termoose/irccloud/requests"
)

type channel struct {
	layout *tview.Grid

	name   string
	chat   *tview.TextView
	users  *tview.List
	input  *tview.InputField
	info   *tview.TextView
	cid    int
}

type View struct {
	pages         *tview.Pages
	app           *tview.Application
	activeChannel int
	channels      []channel
	websocket     *requests.Connection
}

func NewView(socket *requests.Connection) (*View) {
	view := &View{
		pages: tview.NewPages(),
		websocket: socket,
	}

	return view
}

func (v *View) Start() {
	v.app = tview.NewApplication()

	v.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		total_active := v.pages.GetPageCount()

		if event.Key() == tcell.KeyLeft {
			v.activeChannel = (v.activeChannel - 1 + total_active) % total_active
			page := v.channels[v.activeChannel]
			v.pages.SwitchToPage(page.name)
			v.app.SetFocus(page.input)
		}

		if event.Key() == tcell.KeyRight {
			v.activeChannel = (v.activeChannel + 1) % total_active
			page := v.channels[v.activeChannel]
			v.pages.SwitchToPage(page.name)
			v.app.SetFocus(page.input)
		}

		return event
	})

	if err := v.app.SetRoot(v.pages, true).SetFocus(v.pages).Run(); err != nil {
		panic(err)
	}
}

func (v *View) AddChannel(name string, cid int, user_list []string) {
	new_chan := channel{
		layout: tview.NewGrid().SetRows(1, 0, 1).SetColumns(20, 0, 20).SetBorders(true),
		name: name,
		chat: newTextView(""),
		users: newListView(),
		input: newTextInput(),
		info: newTextView(name),
		cid: cid,
	}

	// Set callback for handling message sending
	new_chan.input.SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				v.sendToBuffer(cid, name, new_chan.input.GetText())
				new_chan.input.SetText("")
			}
	})

	for _, user := range user_list {
		new_chan.users.AddItem(user, user, 0, nil)
	}

	// Layout
	new_chan.layout.AddItem(new_chan.users, 0, 2, 3, 1, 0, 0, false)
	new_chan.layout.AddItem(new_chan.chat,  1, 0, 1, 2, 0, 0, false)
	new_chan.layout.AddItem(new_chan.input, 2, 0, 1, 2, 0, 0, false)
	new_chan.layout.AddItem(new_chan.info,  0, 0, 1, 2, 0, 0, false)

	v.pages.AddPage(name, new_chan.layout, true, true)

	// FIXME: Might not be the best idea to keep this counter
	v.activeChannel = v.pages.GetPageCount()

	v.channels = append(v.channels, new_chan)

	v.app.SetFocus(new_chan.input)
	v.app.Draw()
}

func (v *View) sendToBuffer(cid int, channel, message string) {
	v.websocket.SendMessage(cid, channel, message)
}

func (v *View) getChannel(name string) *channel {
	for i, c := range v.channels {
		if c.name == name {
			return &v.channels[i]
		}
	}

	return nil
}

func (v *View) AddUser(channel, nick string) {
	c := v.getChannel(channel)

	c.users.AddItem(nick, nick, 0, nil)
}

func (v *View) RemoveUser(channel, nick string) {
	c := v.getChannel(channel)

	list := c.users.FindItems(nick, nick, true, false)
	if len(list) > 0 {
		// Remove the first occurence
		c.users.RemoveItem(list[0])
	}
}

func (v *View) AddPartEvent(channel, nick, hostmask string) {
	c := v.getChannel(channel)

	if c != nil {
		line := fmt.Sprintf("  <- %s left (%s)\n", nick, hostmask)
		c.chat.Write([]byte(line))
		c.chat.ScrollToEnd()
	}
}

func (v *View) AddJoinEvent(channel, nick, hostmask string) {
	c := v.getChannel(channel)

	if c != nil {
		line := fmt.Sprintf("  -> %s joined (%s)\n", nick, hostmask)
		c.chat.Write([]byte(line))
		c.chat.ScrollToEnd()
	}
}

func (v *View) AddBufferMsg(channel, from, msg string) {
	c := v.getChannel(channel)

	if c != nil {
		line := fmt.Sprintf("<%s> %s\n", from, msg)
		c.chat.Write([]byte(line))
		c.chat.ScrollToEnd()
	}

	v.app.Draw()
}

func newTextInput() *tview.InputField {
	return tview.NewInputField().
		SetFieldBackgroundColor(tcell.ColorBlack).
			SetPlaceholder("type here...")
}

func newListView() *tview.List {
	return tview.NewList().ShowSecondaryText(false)
}

func newTextView(text string) *tview.TextView {
	return tview.NewTextView().SetText(text)
}
