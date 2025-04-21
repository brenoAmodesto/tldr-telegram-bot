package telegram

import (
    "log"
    "database/sql"
    "fmt"
    "strings"
    "time"
    "tldr-telegram-bot/internal/db"
)


// FormatMessage formats a message with the user's identifier.
func FormatMessage(name, lastName, username string, userID int64) string {
	if name != "" && lastName != "" {
		return fmt.Sprintf("%s %s: ", name, lastName)
	} else if name != "" {
		return fmt.Sprintf("%s: ", name)
	} else if username != "" {
		return fmt.Sprintf("@%s: ", username)
	}
	return fmt.Sprintf("%d: ", userID)
}

// IsTriggerWord checks if the message contains any of the trigger words.
func IsTriggerWord(message string) bool {
	triggerWords := []string{"resuma", "tldr", "summary", "toguro por favor", "toguro please", "toguro"}
	for _, word := range triggerWords {
		if strings.EqualFold(strings.TrimSpace(message), word) {
			return true
		}
	}
	return false
}

// CleanText removes newlines and trims the input string.
func CleanText(text string) string {
	text = strings.ReplaceAll(text, "\n", " ")
	return strings.TrimSpace(text)
}

// GetMessagesAfterReply 
func GetMessagesAfterReply(dbConn *sql.DB, messageID, groupID int64) ([]db.Message, error) {
    // 1) pegalo timestamp
    ts, err := db.GetMessageTimestamp(dbConn, messageID, groupID)
    if err != nil {
        return nil, err
    }
    if ts == nil {
        return nil, nil
    }

    // 2) 30 minuto de intervalo
    start := *ts
    end   := start.Add(30 * time.Minute)

    // 3) chama func
    return db.GetMessagesByTimeRange(dbConn, groupID, start, end)
}

// LogConfigError logs failures loading config or DB.
func LogConfigError(err error) {
    log.Printf("⚠️ Error loading config or DB: %v", err)
}

// LogLocationError logs failures loading the desired timezone.
func LogLocationError(err error) {
    log.Printf("⚠️ Error loading timezone, falling back to UTC: %v", err)
}

// LogCheckingInterval logs the start of a 3h‑block check.
func LogCheckingInterval(from, to time.Time) {
    log.Printf("🔍 Checking messages from %s to %s",
        from.UTC().Format(time.RFC3339),
        to.UTC().Format(time.RFC3339),
    )
}

// LogDBError logs DB query errors for a given interval.
func LogDBError(err error, from, to time.Time) {
    log.Printf("❌ Error fetching messages from %s to %s: %v",
        from.Format("15:04"), to.Format("15:04"), err,
    )
}

// LogRetrievedCount logs how many messages were retrieved.
func LogRetrievedCount(count int, from, to time.Time) {
    log.Printf("✅ %d messages found between %s and %s",
        count, from.Format("15:04"), to.Format("15:04"),
    )
}

// LogSummarizeError logs failures during the LLM summarization.
func LogSummarizeError(err error, from, to time.Time) {
    log.Printf("❌ Error summarizing block %s - %s: %v",
        from.Format("15:04"), to.Format("15:04"), err,
    )
}

// LogNoSummaryContent logs that there was no content to summarize.
func LogNoSummaryContent(chatID int64) {
    log.Printf("⚠️ No content to summarize for group %d", chatID)
}