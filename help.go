package main

import (
	"fmt"
	"strings"

	tb "gopkg.in/tucnak/telebot.v2"
)

func getHelp(message *tb.Message) {
	var sb strings.Builder
	sb.WriteString("<strong>cosmos-interacter</strong>\n\n")
	sb.WriteString(fmt.Sprintf("Query for the %s network info.\n", NetworkName))
	sb.WriteString("Can understand the following commands:\n")
	sb.WriteString("- /wallet &lt;wallet address&gt; - get the wallet info (balance, delegated amount, rewards etc.)\n")
	sb.WriteString("- /validator &lt;validator address or name&gt; - get validator info\n")
	sb.WriteString("- /rate - get the Coingecko exchange rate to USD\n")
	sb.WriteString("- /proposal &lt;proposal ID&gt; - get the proposal info\n")
	sb.WriteString("- /proposals - proposals list\n")
	sb.WriteString("- /help - display this message\n")
	sb.WriteString("- /about - get info about this bot and its creators\n\n")
	sb.WriteString("<strong>Useful links:</strong>\n")
	sb.WriteString(fmt.Sprintf("<a href=\"https://mintscan.io/%s\">Mintscan</a> - the network explorer powered by Cosmostation\n", MintscanPrefix))
	sb.WriteString(fmt.Sprintf("<a href=\"https://www.coingecko.com/en/coins/%s\">Coingecko</a> - Coingecko exchange rate\n", CoingeckoCurrency))
	sb.WriteString("<a href=\"https://play.google.com/store/apps/details?id=wannabit.io.cosmostaion\">Cosmostation Wallet for Android</a>\n")
	sb.WriteString("<a href=\"https://apps.apple.com/us/app/cosmostation/id1459830339\">Cosmostation Wallet for iOS</a>\n")

	sendMessage(message, sb.String())
	log.Info().
		Str("user", message.Sender.Username).
		Msg("Successfully returned help info")
}
