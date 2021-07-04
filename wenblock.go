package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	tmrpc "github.com/tendermint/tendermint/rpc/client/http"
	ctypes "github.com/tendermint/tendermint/types"

	tb "gopkg.in/tucnak/telebot.v2"
)

var BlocksDiffInThePast int64 = 100

func getBlockApproximateDate(message *tb.Message) {
	args := strings.SplitAfterN(message.Text, " ", 2)
	if len(args) < 2 {
		log.Info().Msg("getBlockApproximateDate: args length < 2")
		sendMessage(message, "Usage: wenblock &lt;proposal ID&gt;")
		return
	}

	blockHeightProvided, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		log.Error().Err(err).Msg("getBlockApproximateDate: Could not parse block")
		sendMessage(message, "Block should be a number!")
		return
	}

	latestBlock, err := getBlock(nil)
	if err != nil {
		log.Error().Err(err).Msg("getBlockApproximateDate: Could not get latest block")
		sendMessage(message, "Could not get block info")
		return
	}

	if blockHeightProvided <= latestBlock.Height {
		log.Debug().Int64("height", latestBlock.Height).Msg("Block is in the past.")
		if block, err := getBlock(&blockHeightProvided); err != nil {
			log.Error().Err(err).Msg("getBlockApproximateDate: Could not get latest block")
			sendMessage(message, "Could not get block info")
		} else {
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("<strong>Block #%d</strong>\n", block.Height))
			sb.WriteString(fmt.Sprintf("<strong>Generation time: </strong><code>%s</code>\n", block.Time.Format(time.RFC822)))
			sb.WriteString(fmt.Sprintf("<code>%s</code> in the past.\n", time.Since(block.Time).String()))
			sb.WriteString(fmt.Sprintf("<a href=\"https://mintscan.io/%s/blocks/%d\">Mintscan</a>\n\n", MintscanPrefix, blockHeightProvided))

			sendMessage(message, sb.String())
		}

		return
	}

	latestHeight := latestBlock.Height
	beforeLatestBlockHeight := latestBlock.Height - BlocksDiffInThePast
	beforeLatestBlock, err := getBlock(&beforeLatestBlockHeight)

	if err != nil {
		log.Error().Err(err).Msg("getBlockApproximateDate: Could not parse before latest block")
		sendMessage(message, "Could not get block info")
		return
	}

	heightDiff := float64(latestHeight - beforeLatestBlockHeight)
	timeDiff := latestBlock.Time.Sub(beforeLatestBlock.Time).Seconds()

	avgBlockTime := timeDiff / heightDiff

	log.Debug().
		Float64("heightDiff", heightDiff).
		Float64("timeDiff", timeDiff).
		Float64("avgBlockTime", avgBlockTime).
		Msg("Average block time")

	blocksToCalculate := blockHeightProvided - latestHeight

	log.Debug().
		Int64("diff", blocksToCalculate).
		Msg("Blocks till the specified block")

	latestTime := latestBlock.Time
	timeToAddAsSeconds := int64(avgBlockTime * float64(blocksToCalculate))
	timeToAddAsDuration := time.Duration(timeToAddAsSeconds) * time.Second
	calculatedBlockTime := latestTime.Add(timeToAddAsDuration)

	log.Debug().
		Time("diff", calculatedBlockTime).
		Msg("Estimated block time")

	log.Debug().
		Str("diff", timeToAddAsDuration.String()).
		Msg("Time till block")

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<strong>Block #%d</strong>\n", blockHeightProvided))
	sb.WriteString(fmt.Sprintf("<strong>Generation time: </strong><code>%s</code>\n", calculatedBlockTime.Format(time.RFC822)))
	sb.WriteString(fmt.Sprintf("<code>%s</code> in the future.\n", timeToAddAsDuration.String()))

	sendMessage(message, sb.String())
}

func getBlock(height *int64) (*ctypes.Block, error) {
	client, err := tmrpc.New(TendermintRpc, "/websocket")
	if err != nil {
		log.Error().Err(err).Msg("Could not create Tendermint client")
		return &ctypes.Block{}, err
	}

	block, err := client.Block(context.Background(), height)
	if err != nil {
		log.Error().Err(err).Msg("Could not query Tendermint status")
		return &ctypes.Block{}, err
	}

	return block.Block, nil
}
