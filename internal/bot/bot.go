package bot

import (
	"fmt"
	"log"
	"time"
	"whisp/internal/config"
	"whisp/internal/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

//Start initializes and run tg bot
func Start(cfg *config.Config, repo *storage.Repository) {
	deepseekAPIKey := cfg.DeepSeekAPIKey
	bot, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		log.Fatalf("failed to create bot: %v", err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	//Configure updates
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	//Process incoming messages
	for update := range updates {
		if update.Message == nil {
			continue
		}
		handleMessage(bot, update.Message, repo, deepseekAPIKey)
	}
}

func handleMessage(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, repo *storage.Repository, apiKey string){

	// Check for /usage command
	if msg.Text == "/usage" {
		checkUsageCommand(bot, msg, repo)
		return
	}else if msg.Text == "/help"{
		sendHelpMessage(bot, msg)
		return
	}

	//Rate limit check
	currentDate := time.Now().UTC()

	//Get usage stats
	tokensUsed, err := repo.GetDailyUsage(msg.Chat.ID, currentDate)
	if err != nil {
		log.Printf("Failed to get daily usage: %v", err)
	}

	promptsUsed, err := repo.GetDailyPrompts(msg.Chat.ID, currentDate)
	if err != nil {
		log.Printf("Failed to g et daily prompts: %v", err)
	}

	// Check limits
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Failed to load config: %v", err)
		return
	}
	if tokensUsed >= cfg.DailyTokenLimit || promptsUsed >= cfg.DailyPromptLimit {
        hoursRemaining := 24 - currentDate.Hour()
        response := fmt.Sprintf("⚠️ Daily limits reached:\n- Used: %d/%d tokens\n- Used: %d/%d prompts\nReset in ~%d hours",
            tokensUsed, cfg.DailyTokenLimit,
            promptsUsed, cfg.DailyPromptLimit,
            hoursRemaining)
        
        reply := tgbotapi.NewMessage(msg.Chat.ID, response)
        bot.Send(reply)
        return
    }

	//Save the incoming message
	if err := repo.SaveMessage(msg.Chat.ID, msg.Text); err != nil {
		log.Printf("Failed to save message: %v", err)
	}

	//Get conversation context (last 5 messages)
	context, err := repo.GetLastMessages(msg.Chat.ID, 5)
	if err != nil {
		log.Printf("Failed to get context: %v", err)
	}

	//Generate responspe using DeepSeek
	response, tokens, err := callDeepSeek(msg.Text, context, apiKey)
	if err !=nil {
		log.Printf("DeepSeek API Error: %v", err)
		response = "⚠️ Sorry, I’m having trouble connecting to the AI."
	}

	cleanResponse := sanitizeForTelegram(response)

	//Save and send the response
	if err := repo.SaveMessage(msg.Chat.ID, cleanResponse); err != nil {
		log.Printf("failed to save bot response: %v", err)
	}
	reply := tgbotapi.NewMessage(msg.Chat.ID, cleanResponse)
	if _, err := bot.Send(reply); err != nil {
		log.Printf("Failed to send message: %v", err)
	}

	if err == nil {
		err = repo.RecordUsage(msg.Chat.ID, tokens)
		if err != nil {
			log.Printf("failed to record usage: %v", err)
		}
	}
}

func checkUsageCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, repo *storage.Repository){
	currentDate := time.Now().UTC()

	//Get Usage Stats
	tokensUsed, err:= repo.GetDailyUsage(msg.Chat.ID, currentDate)
	if err != nil {
		log.Printf("Error getting usage: %v", err)
		tokensUsed =0
	}
	promptsUsed, err := repo.GetDailyPrompts(msg.Chat.ID, currentDate)
	if err != nil {
		log.Printf("Error getting prompts: %v", err)
		promptsUsed = 0
	}

	//Get limits from cfg
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Error loading config: %v", err)
		return
	}

	// Calculate remaining limits
	tokensRemaining := cfg.DailyTokenLimit - tokensUsed
	promptsRemaining := cfg.DailyPromptLimit - promptsUsed
	hoursRemaining := 24 - currentDate.Hour()

	response := fmt.Sprintf(`📊 Daily Usage:
    
• Tokens Used: %d/%d
• Prompts Used: %d/%d
• Remaining Tokens: %d
• Remaining Prompts: %d
⏳ Resets in ~%d hours`,
        tokensUsed, cfg.DailyTokenLimit,
        promptsUsed, cfg.DailyPromptLimit,
        tokensRemaining,
        promptsRemaining,
		hoursRemaining,
	)
	reply := tgbotapi.NewMessage(msg.Chat.ID, response)
	if _, err := bot.Send(reply); err != nil {
		log.Printf("Failed to send message: %v", err)
	}
}

func sendHelpMessage(bot *tgbotapi.BotAPI, msg *tgbotapi.Message){
	helpText := `🤖 Bot Commands:
/help - Show this help
/usage - Check your daily usage limits
     
Ask me anything else to start chatting!`
    
    reply := tgbotapi.NewMessage(msg.Chat.ID, helpText)
    bot.Send(reply)

}

