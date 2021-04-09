package main

import (
	"context"
	"strings"

	esive_grpc "github.com/code-cell/esive/grpc"
	tcell "github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ChatView struct {
	*tview.Flex

	client   esive_grpc.EsiveClient
	backView tview.Primitive
	app      *tview.Application

	textView *tview.TextView
	input    *tview.InputField

	lines []string
}

func NewChatView(client esive_grpc.EsiveClient, app *tview.Application) *ChatView {
	textView := tview.NewTextView()
	textView.SetBorder(true)

	input := tview.NewInputField().SetLabel("Say: ")
	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(textView, 0, 1, true).
		AddItem(input, 2, 0, true)

	chatView := &ChatView{
		Flex:     flex,
		textView: textView,
		input:    input,
		app:      app,
	}

	input.
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				text := input.GetText()
				go func() {
					_, err := client.Say(context.Background(), &esive_grpc.SayReq{Text: text})
					if err != nil {
						panic(err)
					}
				}()
			}
			input.SetText("")
			app.SetFocus(chatView.backView)
		})

	return chatView
}

func (r *ChatView) SetBackView(backView tview.Primitive) {
	r.backView = backView
}

func (r *ChatView) Append(line string) {
	r.lines = append(r.lines, line)

	r.textView.SetText(strings.Join(r.lines, "\n"))
	go r.app.Draw()
}
