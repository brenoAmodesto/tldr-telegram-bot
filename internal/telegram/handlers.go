package telegram

import (
	"fmt"
	"log"
	"os"
	"strings"

	"tldr-telegram-bot/internal/config"
	"tldr-telegram-bot/internal/db"
	"tldr-telegram-bot/internal/llm"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var triggerWords = []string{"resuma", "resume", "tldr", "summary", "toguro por favor", "toguro please", "professor toguro", "professor toguro por favor", "professor toguro please", "toguro", "toguro por favor", "toguro please", "toguro professor", "toguro professor por favor", "toguro professor please"}

func HandleMessage(update tgbotapi.Update) {
	if update.Message == nil || update.Message.ReplyToMessage == nil {
		return
	}

	if !isAuthorizedGroup(update.Message.Chat.ID) {
		logUnauthorizedAttempt(update.Message.Chat.ID)
		return
	}

	if isTriggerWord(update.Message.Text) {
		log.Printf("Trigger word detected in group %d", update.Message.Chat.ID)
		collectAndSummarizeMessages(update)
	}
}

func isAuthorizedGroup(groupID int64) bool {
	config, err := config.LoadConfig()
	if err != nil {
		log.Printf("Error loading config: %v", err)
		return false
	}

	// Check if the group ID is in the list of authorized groups
	for _, authorizedGroup := range config.AuthorizedGroups {
		if authorizedGroup == groupID {
			return true
		}
	}
	log.Printf("Unauthorized group access attempt: %d", groupID)
	return false
}

func logUnauthorizedAttempt(groupID int64) {
	log.Printf("Unauthorized access attempt in group: %d", groupID)
}

func isTriggerWord(message string) bool {
	for _, word := range triggerWords {
		if strings.Contains(strings.ToLower(message), word) {
			return true
		}
	}
	return false
}

func collectAndSummarizeMessages(update tgbotapi.Update) {
	myDb := db.GetDB()

	myConfig, err := config.LoadConfig()
	if err != nil {
		log.Printf("Error loading config: %v", err)
		return
	}

	// func que ta no utils faze a logica do 30
	messages, err := GetMessagesAfterReply(
		myDb,
		int64(update.Message.ReplyToMessage.MessageID),
		update.Message.Chat.ID,
	)
	if err != nil {
		log.Printf("Error collecting messages: %v", err)
		return
	}

	if len(messages) == 0 {
		log.Println("No messages found for summarization.")
		return
	}
	
	// que estava se repetindo jogadao na função cleantxt que está em utils.go
	concatenatedText := CleanText(formatMessages(messages))
	fmt.Println("Concatenated text for summarization:", concatenatedText)

	if os.Getenv("LOCAL_MODEL") != "true" {
		summary, err := llm.SummarizeGemini(concatenatedText, myConfig.Lang)
		if err != nil {
			log.Printf("Error summarizing messages: %v", err)
			return
		}
		sendSummary(update.Message.Chat.ID, summary)
		return
	}
	summary, err := llm.Summarize(concatenatedText, myConfig.Lang)
	if err != nil {
		log.Printf("Error summarizing messages: %v", err)
		return
	}

	sendSummary(update.Message.Chat.ID, summary)
}

func formatMessages(messages []db.Message) string {
	var sb strings.Builder
	for _, msg := range messages {
		sender := msg.Username
		switch {
		case msg.Name != "" && msg.LastName != "":
			sender = fmt.Sprintf("%s %s", msg.Name, msg.LastName)
		case msg.Name != "":
			sender = msg.Name
		case msg.LastName != "":
			sender = msg.LastName
		}
		sb.WriteString(fmt.Sprintf("%s: %s\n", sender, msg.Content))
	}
	return sb.String()
}

func sendSummary(chatID int64, summary string) {
	bot, err := NewBot()
	if err != nil {
		log.Printf("Error creating bot instance: %v", err)
		return
	}

	msg := tgbotapi.NewMessage(chatID, summary)
	_, err = bot.api.Send(msg)
	if err != nil {
		log.Printf("Error sending summary: %v", err)
	}
}