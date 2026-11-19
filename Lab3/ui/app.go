package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"Lab3/peer"
	"Lab3/protocol"
)

type UserInterface struct {
	application       *tview.Application
	networkNode       *peer.Node
	chatTextView      *tview.TextView
	userListTextView  *tview.TextView
	messageInputField *tview.InputField
	userNickname      string
}

func NewUserInterface(networkNode *peer.Node, userNickname string) *UserInterface {
	var userInterface = &UserInterface{
		application:       tview.NewApplication(),
		networkNode:       networkNode,
		chatTextView:      tview.NewTextView(),
		userListTextView:  tview.NewTextView(),
		messageInputField: tview.NewInputField(),
		userNickname:      userNickname,
	}

	userInterface.setupComponents()
	userInterface.setupCallbacks()

	return userInterface
}

func (userInterface *UserInterface) setupComponents() {
	userInterface.application.EnableMouse(true)

	userInterface.chatTextView.
		SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(true).
		SetChangedFunc(func() {
			userInterface.chatTextView.ScrollToEnd()
		})
	userInterface.chatTextView.SetBorder(true).SetTitle(" Czat ")

	userInterface.userListTextView.
		SetDynamicColors(true).
		SetScrollable(true)
	userInterface.userListTextView.SetBorder(true).SetTitle(" Uzytkownicy ")

	userInterface.messageInputField.
		SetLabel(" Wpisz wiadomosc: ").
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetFieldTextColor(tcell.ColorWhite)
	userInterface.messageInputField.SetBorder(true).SetTitle(" Wyslij: Enter ")

	userInterface.messageInputField.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			var messageContent = strings.TrimSpace(userInterface.messageInputField.GetText())
			if messageContent == "" {
				return
			}

			sendingError := userInterface.networkNode.SendChatMessage(messageContent)
			if sendingError != nil {
				userInterface.AppendMessage("", fmt.Sprintf("Blad wysylania: %v", sendingError))
				return
			}

			userInterface.AppendMessage(userInterface.userNickname, messageContent)
			userInterface.messageInputField.SetText("")
		}
	})

	var headerTextView = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	var headerContent = fmt.Sprintf(
		"[yellow]Aplikacja czatu[-] | Zalogowany jako: [green]%s[-] | Wyjscie: [red]ESC / Ctrl+C[-]",
		userInterface.userNickname,
	)
	headerTextView.SetText(headerContent)

	var centralColumnsFlex = tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(userInterface.chatTextView, 0, 3, false).
		AddItem(userInterface.userListTextView, 0, 1, false)

	var mainLayoutFlex = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(headerTextView, 1, 0, false).
		AddItem(centralColumnsFlex, 0, 1, false).
		AddItem(userInterface.messageInputField, 3, 0, true)

	userInterface.application.SetRoot(mainLayoutFlex, true)
	userInterface.application.SetFocus(userInterface.messageInputField)

	userInterface.application.SetInputCapture(func(eventKey *tcell.EventKey) *tcell.EventKey {
		if eventKey.Key() == tcell.KeyEscape || eventKey.Key() == tcell.KeyCtrlC {
			userInterface.networkNode.LeaveAndClose()
			userInterface.application.Stop()
			return nil
		}
		return eventKey
	})
}

func (userInterface *UserInterface) setupCallbacks() {
	userInterface.networkNode.OnChatMessageReceived = func(chatMessage protocol.Message) {
		userInterface.application.QueueUpdateDraw(func() {
			userInterface.AppendMessage(chatMessage.Nickname, chatMessage.Content)
		})
	}

	userInterface.networkNode.OnSystemEventReceived = func(systemNotification string) {
		userInterface.application.QueueUpdateDraw(func() {
			userInterface.AppendMessage("", systemNotification)
		})
	}

	userInterface.networkNode.OnUserListChanged = func(activeUserNicknames []string) {
		userInterface.application.QueueUpdateDraw(func() {
			userInterface.RenderUserList(activeUserNicknames)
		})
	}
}

func (userInterface *UserInterface) AppendMessage(senderNickname string, messageContent string) {
	var currentTimeFormatted = time.Now().Format("15:04:05")
	if senderNickname == "" {
		fmt.Fprintf(userInterface.chatTextView, "[grey]%s[-] [yellow]* %s*[-]\n", currentTimeFormatted, messageContent)
		return
	}

	var senderColor = "cyan"
	if senderNickname == userInterface.userNickname {
		senderColor = "green"
	}
	fmt.Fprintf(userInterface.chatTextView, "[grey]%s[-] [%s]%s>[-] %s\n", currentTimeFormatted, senderColor, senderNickname, messageContent)
}

func (userInterface *UserInterface) RenderUserList(activeUserNicknames []string) {
	userInterface.userListTextView.Clear()
	var header = fmt.Sprintf("[yellow]Aktywni (%d):[-]\n\n", len(activeUserNicknames))
	fmt.Fprint(userInterface.userListTextView, header)

	for _, activeNickname := range activeUserNicknames {
		if activeNickname == userInterface.userNickname {
			fmt.Fprintf(userInterface.userListTextView, " [green]* %s (Ty)[-]\n", activeNickname)
		} else {
			fmt.Fprintf(userInterface.userListTextView, "   %s\n", activeNickname)
		}
	}
}

func (userInterface *UserInterface) Run() error {
	userInterface.RenderUserList(userInterface.networkNode.GetActiveUsers())
	userInterface.AppendMessage("", "Polaczono z grupa multicast. Witamy na czacie!")

	return userInterface.application.Run()
}
