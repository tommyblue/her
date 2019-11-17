# Her

Her is a home assistant.

## Features

* Connect to a MQTT server
* Connect to a Telegram bot
* Subscribe to MQTT topics and send notifications to Telegram when the value changes

## Config

Create a toml file with your configuration and use it with her.
You need a valid token to send messages to your [Telegram bot](https://core.telegram.org/bots) and  
its channel id. To obtain the channel id you can use the instructions provided in this
[Stackoverflow question](https://stackoverflow.com/questions/33858927/how-to-obtain-the-chat-id-of-a-private-telegram-channel)

The [config.example.toml](config.example.toml) file contains all possible configurations, so use
it as a template.

## Build

To build her, run `make build`.
If you want to build it for Raspberry Pi, run `make build-pi`.
