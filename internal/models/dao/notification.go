package dao

// Notification is one message to deliver, before it's split across channels and recipients.
type Notification struct {
	App       string
	Category  string
	Title     string
	Body      string
	Link      string
	ImageURL  string
	Data      map[string]string
	DedupKey  string // empty: never deduplicated
	CreatedAt int64  // unix millis
}

// Recipient is who a notification is for; each channel uses the field it needs.
type Recipient struct {
	ProfileID int64
	Phone     string
	Email     string
}

// InAppNotification is a notification as stored in one profile's inbox.
type InAppNotification struct {
	ID        string
	App       string
	ProfileID int64
	Category  string
	Title     string
	Body      string
	Link      string
	ImageURL  string
	Data      map[string]string
	DedupKey  string
	CreatedAt int64 // unix millis
	ReadAt    int64 // unix millis; 0 while unread
}
