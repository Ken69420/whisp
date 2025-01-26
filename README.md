# Whisp

Whisp is a Telegram bot that leverages the DeepSeek API to provide intelligent responses to user queries. It also tracks usage metrics and enforces daily limits on token and prompt usage.

## Features

- **Telegram Integration**: Communicate with users via Telegram.
- **DeepSeek API**: Generate responses using the DeepSeek API.
- **Usage Tracking**: Track daily token and prompt usage.
- **Rate Limiting**: Enforce daily limits on token and prompt usage.

## Installation

1. Clone the repository:

   ```sh
   git clone https://github.com/yourusername/whisp.git
   cd whisp
   ```

2. Install dependencies:

   ```sh
   go mod tidy
   ```

3. Create a `.env` file in the root directory with the following content:

   ```env
   TELEGRAM_TOKEN=your_telegram_token
   DATABASE_PATH="./data/chat.db"
   DEEPSEEK_API_KEY=your_deepseek_api_key
   ```

4. Build the project:

   ```sh
   go build -o bin/bot cmd/bot/main.go
   ```

5. Run the bot:
   ```sh
   ./bin/bot
   ```

## Usage

- **/help**: Show help message.
- **/usage**: Check your daily usage limits.
- **Chat**: Ask anything to start chatting with the bot.

## Project Structure

- `cmd/bot/main.go`: Entry point of the application.
- `internal/bot/`: Contains the bot logic and handlers.
- `internal/config/`: Configuration loading.
- `internal/storage/`: Database interactions and repository pattern.
- `.env`: Environment variables for configuration.

## Dependencies

- [go-telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api)
- [godotenv](https://github.com/joho/godotenv)
- [go-sqlite3](https://github.com/mattn/go-sqlite3)
