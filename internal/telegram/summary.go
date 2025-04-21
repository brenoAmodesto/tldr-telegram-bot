package telegram

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"tldr-telegram-bot/internal/config"
	"tldr-telegram-bot/internal/db"
	"tldr-telegram-bot/internal/llm"
)

func summarizeDay(chatID int64) {
    myDb, myConfig, err := func() (*sql.DB, *config.Config, error) {
        cfg, err := config.LoadConfig()
        if err != nil {
            return nil, nil, err
        }
        return db.GetDB(), cfg, nil
    }()
    if err != nil {
        LogConfigError(err)
        return
    }

    // fallback 
    loc, err := time.LoadLocation("America/Sao_Paulo")
    if err != nil {
        LogLocationError(err)
        loc = time.UTC
    }

    // 24h padrão
    now := time.Now().In(loc)
    dayStart := time.Date(
        now.Year(), now.Month(), now.Day(),
        3, 0, 0, 0, loc,
    ).Add(-24 * time.Hour)

    var finalSummary strings.Builder
    finalSummary.WriteString("📊 *Relatório diário - barra pesada onlaine*\n\n")

    //Oito blocos de 3h
    for i := 0; i < 8; i++ {
        from := dayStart.Add(time.Duration(i*3) * time.Hour)
        to   := from.Add(3 * time.Hour)

        LogCheckingInterval(from, to)

        msgs, err := db.GetMessagesByTimeRange(myDb, chatID, from.UTC(), to.UTC())
        if err != nil {
            LogDBError(err, from, to)
            continue
        }
        LogRetrievedCount(len(msgs), from, to)

        if len(msgs) == 0 {
            finalSummary.WriteString(fmt.Sprintf(
                "🕒 %s - %s\nNenhuma mensagem nesse período.\n\n",
                from.Format("15:04"), to.Format("15:04"),
            ))
            continue
        }

        text := CleanText(formatMessages(msgs))
        if len(text) > 412 {
            text = text[:412]
        }

        var summary string
        if os.Getenv("LOCAL_MODEL") != "true" {
            summary, err = llm.SummarizeGeminiLimited(text, myConfig.Lang)
        } else {
            summary, err = llm.Summarize(text, myConfig.Lang)
        }
        if err != nil {
            LogSummarizeError(err, from, to)
            continue
        }

        finalSummary.WriteString(fmt.Sprintf(
            "🕒 %s - %s\n%s\n\n",
            from.Format("15:04"), to.Format("15:04"), summary,
        ))
    }

    //Envio
    if finalSummary.Len() > 0 {
        sendSummary(chatID, finalSummary.String())
    } else {
        LogNoSummaryContent(chatID)
    }
}



func RunDailySummary() {
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Printf("Failed to load config for daily summary: %v", err)
        return
    }

    for _, groupID := range cfg.AuthorizedGroups {
        log.Printf("🟡 Running daily summary for group: %d", groupID)
        summarizeDay(groupID)
    }
}
