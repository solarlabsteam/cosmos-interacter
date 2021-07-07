package main

import (
	"fmt"
	"strings"

	tb "gopkg.in/tucnak/telebot.v2"

	gecko "github.com/superoo7/go-gecko/v3"
)

func getRate(message *tb.Message) {
	var sb strings.Builder

	if CoingeckoCurrency != "" {
		var cg = gecko.NewClient(nil)
		if result, err := cg.SimpleSinglePrice(CoingeckoCurrency, "usd"); err != nil {
			log.Error().Err(err).Str("currency", CoingeckoCurrency).Msg("Could not get Coingecko currency rate")
		} else {
			sb.WriteString(fmt.Sprintf("<code>$%.3f</code> ", result.MarketPrice))
			sb.WriteString(fmt.Sprintf("<a href=\"https://www.coingecko.com/en/coins/%s\">Coingecko</a>\n", CoingeckoCurrency))
		}
	}

	if AscendexCurrency != "" {
		if result, err := getAscendexRate(); err != nil {
			log.Error().Err(err).Str("currency", AscendexCurrency).Msg("Could not get Ascendex currency rate")
		} else {
			sb.WriteString(fmt.Sprintf("<code>$%.3f</code> ", result))
			sb.WriteString(fmt.Sprintf(
				"<a href=\"https://ascendex.com/en/basic/cashtrade-spottrading/%s/xprt\">Ascendex</a>\n",
				strings.ToLower(AscendexCurrency),
			))
		}
	}

	if text := sb.String(); text == "" {
		sendMessage(message, "Could not get currency rate")
		log.Error().
			Str("currency", CoingeckoCurrency).
			Str("user", message.Sender.Username).
			Msg("Could not get any currency rate")
	} else {
		sendMessage(message, sb.String())
		log.Info().
			Str("currency", CoingeckoCurrency).
			Str("user", message.Sender.Username).
			Msg("Successfully returned currency info")
	}
}
