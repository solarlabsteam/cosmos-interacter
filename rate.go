package main

import (
	"fmt"
	"strings"

	tb "gopkg.in/tucnak/telebot.v2"

	gecko "github.com/superoo7/go-gecko/v3"
)

func getRate(message *tb.Message) {
	var cg = gecko.NewClient(nil)
	result, err := cg.SimpleSinglePrice(CoingeckoCurrency, "usd")
	if err != nil {
		sendMessage(message, "Could not get currency rate")
		return
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<code>$%.3f</code> ", result.MarketPrice))
	sb.WriteString(fmt.Sprintf("<a href=\"https://www.coingecko.com/en/coins/%s\">Coingecko</a>", CoingeckoCurrency))

	sendMessage(message, sb.String())
	log.Info().
		Str("currency", CoingeckoCurrency).
		Msg("Successfully returned currency info")
}
