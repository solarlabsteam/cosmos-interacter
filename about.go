package main

import (
	"strings"

	tb "gopkg.in/tucnak/telebot.v2"
)

func getAbout(message *tb.Message) {
	var sb strings.Builder
	sb.WriteString("<strong>cosmos-interacter</strong>\n\n")
	sb.WriteString("A bot that can return info about Cosmos-based blockchain params. Check /help for more info.\n")
	sb.WriteString("Created by<a href=\"https://validator.solar\">SOLAR Labs</a> with ❤️.\n")
	sb.WriteString("This bot is open-sourced, you can get the source code at https://github.com/solarlabsteam/cosmos-interacter.\n\n")
	sb.WriteString("We also maintain the following tools for Cosmos ecosystem:")
	sb.WriteString("- <a href=\"https://github.com/solarlabsteam/missed-blocks-checker\">missed-blocks-checker</a> - monitor for validator' missing blocks\n")
	sb.WriteString("- <a href=\"https://github.com/solarlabsteam/cosmos-exporter\">cosmos-exporter</a> - scrape the blockchain data from the local node and export it to Prometheus\n")
	sb.WriteString("- <a href=\"https://github.com/solarlabsteam/coingecko-exporter\">coingecko-exporter</a> - scrape the Coingecko exchange rate and export it to Prometheus\n")
	sb.WriteString("- <a href=\"https://github.com/solarlabsteam/cosmos-transactions-bot\">cosmos-transactions-bot</a> - monitor the incoming transactions for a given filter\n\n")
	sb.WriteString("If you like what we're doing, consider staking with us!\n")
	sb.WriteString("- <a href=\"https://www.mintscan.io/sentinel/validators/sentvaloper1sazxkmhym0zcg9tmzvc4qxesqegs3q4u66tpmf\">Sentinel</a>\n")
	sb.WriteString("- <a href=\"https://www.mintscan.io/persistence/validators/persistencevaloper1kp2sype5n0ky3f8u50pe0jlfcgwva9y79qlpgy\">Persistence</a>\n")

	sendMessage(message, sb.String())
	log.Info().
		Str("user", message.Sender.Username).
		Msg("Successfully returned about info")
}
