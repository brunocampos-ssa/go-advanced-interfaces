package exercises

// Exercise 2 — Notification System
//
// 1. The Notifier interface is defined below.
//
// 2. Implement three types that satisfy Notifier:
//    - EmailNotifier  (has a field: Address string)
//    - SMSNotifier    (has a field: Phone string)
//    - PushNotifier   (has a field: DeviceID string)
//
// 3. Each Notify method should print the message with the channel info.
//    Examples:
//      EmailNotifier{Address: "a@b.com"}.Notify("Hi")
//        → [Email to a@b.com] Hi
//
//      SMSNotifier{Phone: "+5511999"}.Notify("Hi")
//        → [SMS to +5511999] Hi
//
//      PushNotifier{DeviceID: "abc123"}.Notify("Hi")
//        → [Push to abc123] Hi
//
// 4. Implement the function SendNotification that accepts a Notifier
//    and a message, and calls Notify.

// Notifier defines a notification channel.
type Notifier interface {
	Notify(message string)
}

// TODO: Implement EmailNotifier

// TODO: Implement SMSNotifier

// TODO: Implement PushNotifier

// SendNotification sends a message through any Notifier.
func SendNotification(n Notifier, message string) {
	// TODO: Implement.
}
