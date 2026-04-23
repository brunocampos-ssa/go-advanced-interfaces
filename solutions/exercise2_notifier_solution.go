package solutions

import "fmt"

// Notifier defines a notification channel.
type Notifier interface {
	Notify(message string)
}

// EmailNotifier sends notifications via email.
type EmailNotifier struct {
	Address string
}

func (e EmailNotifier) Notify(message string) {
	fmt.Printf("[Email to %s] %s\n", e.Address, message)
}

// SMSNotifier sends notifications via SMS.
type SMSNotifier struct {
	Phone string
}

func (s SMSNotifier) Notify(message string) {
	fmt.Printf("[SMS to %s] %s\n", s.Phone, message)
}

// PushNotifier sends push notifications to a device.
type PushNotifier struct {
	DeviceID string
}

func (p PushNotifier) Notify(message string) {
	fmt.Printf("[Push to %s] %s\n", p.DeviceID, message)
}

// SendNotification sends a message through any Notifier.
func SendNotification(n Notifier, message string) {
	n.Notify(message)
}

// RunNotifierSolution demonstrates the solution for Exercise 2.
func RunNotifierSolution() {
	fmt.Println("=== Exercise 2 Solution: Notification System ===")
	fmt.Println()

	notifiers := []Notifier{
		EmailNotifier{Address: "alice@example.com"},
		SMSNotifier{Phone: "+55 11 99999-0000"},
		PushNotifier{DeviceID: "device-abc-123"},
	}

	for _, n := range notifiers {
		SendNotification(n, "Server is back online!")
	}

	fmt.Println()
}
