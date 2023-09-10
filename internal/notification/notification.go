package notification

import (
	"fmt"
)

type KeyStanding struct {
	key      string
	standing int
}

type Notification interface {
	createMessage() string
	pushMessage(msg string)
}

func PushNotificationMessage(n Notification) {
	fmt.Println("Sending notification...")
	msg := n.createMessage()
	fmt.Println(msg)
	n.pushMessage(msg)
}
