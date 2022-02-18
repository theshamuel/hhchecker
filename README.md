# hhchecker
The http/https healthchecker with notification by: 
 1. Email - Mailgun as provider
 2. Telegram public/private channel

### Application options
```
      --url=                  the URL what you need to healthcheck [$URL]
      --timeout=              the timeout for health probe in seconds (default: 300s) [$TIMEOUT]
      --max-alerts=           the max count of alerts in sequence (default: 3) [$MAX_ALERTS]
      --debug                 debug mode [$DEBUG]

email:
      --email.enabled         enable email mailgun provider [$EMAIL_ENABLED]
      --email.from=           the source email address [$EMAIL_FROM]
      --email.to=             the target email address [$EMAIL_TO]
      --email.cc=             the cc email address [$EMAIL_CC]
      --email.subject=        the subject of email [$EMAIL_SUBJECT]
      --email.text=           the text of email not more 255 letters [$EMAIL_TEXT]
      --email.mailgunApiUrl=  the mailgun API URL for sending notification [$EMAIL_MAILGUN_API_URL]
      --email.mailgunApiKey=  the token for mailgun api [$EMAIL_MAILGUN_API_KEY]

telegram:
      --telegram.enabled      enable telegram provider [$TELEGRAM_ENABLED]
      --telegram.botApiKey=   the telegram bot api key [$TELEGRAM_BOT_API_KEY]
      --telegram.channelName= the channel name without leading symbol @ for public channel only [$TELEGRAM_CHANNEL_NAME]
      --telegram.channelId=   the channel id for private channel only [$TELEGRAM_CHANNEL_ID]
      --telegram.message=     the text message not more 255 letters [$TELEGRAM_MESSAGE]

config:
      --config.enabled        enable getting parameters from config. In that case all parameters will be read only form config [$CONFIG_ENABLED]
      --config.file-name=     config file name (default: hhchecker.yml) [$CONFIG_FILE_NAME]

Help Options:
  -h, --help                  Show this help message
```

