package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tb "gopkg.in/tucnak/telebot.v2"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func getWalletInfo(message *tb.Message) {
	args := strings.Split(message.Text, " ")
	if len(args) < 2 {
		log.Info().
			Str("user", message.Sender.Username).
			Msg("getWalletInfo: args length < 2")
		sendMessage(message, "Usage: wallet &lt;wallet&gt;")
		return
	}

	address := args[1]
	log.Debug().Str("address", address).Msg("getWalletInfo: address")

	// --------------------------------
	bankClient := banktypes.NewQueryClient(grpcConn)
	balancesResponse, err := bankClient.AllBalances(
		context.Background(),
		&banktypes.QueryAllBalancesRequest{Address: address},
	)

	if err != nil {
		log.Error().
			Str("address", address).
			Err(err).
			Msg("Could not get balance")
		sendMessage(message, "Could not get wallet balance")
		return
	}

	delegationsTotal, err := getTotalDelegations(address)
	if err != nil {
		log.Error().
			Str("address", address).
			Err(err).
			Msg("Could not get delegations")
		delegationsTotal = 0
	}

	unbondingsTotal, err := getTotalUnbondings(address)
	if err != nil {
		log.Error().
			Str("address", address).
			Err(err).
			Msg("Could not get unbondings")
		unbondingsTotal = 0
	}

	rewardsTotal, err := getTotalRewards(address)
	if err != nil {
		log.Error().
			Str("address", address).
			Err(err).
			Msg("Could not get rewards")
		rewardsTotal = 0
	}

	// --------------------------------

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<code>%s</code>\n", address))
	sb.WriteString(fmt.Sprintf("<a href=\"https://mintscan.io/%s/account/%s\">Mintscan</a>\n\n", MintscanPrefix, address))

	sb.WriteString("<strong>Balance:        </strong>")

	for _, balance := range balancesResponse.Balances {
		// because cosmos's dec doesn't have .toFloat64() method or whatever and returns everything as int
		if value, err := strconv.ParseFloat(balance.Amount.String(), 64); err != nil {
			log.Error().
				Str("address", address).
				Err(err).
				Msg("Could not parse balance")
		} else {
			sb.WriteString(Printer.Sprintf("<code>%.2f %s</code> ", value/DenomCoefficient, Denom))
		}
	}

	sb.WriteString(Printer.Sprintf(
		"\n<strong>Total delegated: </strong><code>%.2f %s</code>",
		delegationsTotal/DenomCoefficient,
		Denom,
	))

	sb.WriteString(Printer.Sprintf(
		"\n<strong>Total unbonded: </strong><code>%.2f %s</code>",
		unbondingsTotal/DenomCoefficient,
		Denom,
	))

	sb.WriteString(Printer.Sprintf(
		"\n<strong>Total rewards:  </strong><code>%.2f %s</code>",
		rewardsTotal/DenomCoefficient,
		Denom,
	))

	sendMessage(message, sb.String())
	log.Info().
		Str("query", address).
		Str("user", message.Sender.Username).
		Msg("Successfully returned wallet info")
}

func getTotalDelegations(address string) (float64, error) {
	stakingClient := stakingtypes.NewQueryClient(grpcConn)
	delegationsResponse, err := stakingClient.DelegatorDelegations(
		context.Background(),
		&stakingtypes.QueryDelegatorDelegationsRequest{DelegatorAddr: address},
	)

	if err != nil {
		log.Error().
			Str("address", address).
			Err(err).
			Msg("Could not get balance")
		return 0, err
	}

	delegationsTotal := float64(0)
	for _, delegation := range delegationsResponse.DelegationResponses {
		if value, err := strconv.ParseFloat(delegation.Balance.Amount.String(), 64); err != nil {
			log.Error().
				Str("address", address).
				Err(err).
				Msg("Could not parse balance")
			return 0, err
		} else {
			delegationsTotal += value
		}
	}

	return delegationsTotal, nil
}

func getTotalUnbondings(address string) (float64, error) {
	stakingClient := stakingtypes.NewQueryClient(grpcConn)
	unbondingsResponse, err := stakingClient.DelegatorUnbondingDelegations(
		context.Background(),
		&stakingtypes.QueryDelegatorUnbondingDelegationsRequest{DelegatorAddr: address},
	)

	if err != nil {
		log.Error().
			Str("address", address).
			Err(err).
			Msg("Could not get balance")
		return 0, err
	}

	unbondingsTotal := float64(0)
	for _, unbonding := range unbondingsResponse.UnbondingResponses {
		for _, entry := range unbonding.Entries {
			unbondingsTotal += float64(entry.Balance.Int64())
		}
	}

	return unbondingsTotal, nil
}

func getTotalRewards(address string) (float64, error) {
	distributionClient := distributiontypes.NewQueryClient(grpcConn)
	rewardsResponse, err := distributionClient.DelegationTotalRewards(
		context.Background(),
		&distributiontypes.QueryDelegationTotalRewardsRequest{DelegatorAddress: address},
	)

	if err != nil {
		log.Error().
			Str("address", address).
			Err(err).
			Msg("Could not get rewards")
		return 0, err
	}

	rewardsTotal := float64(0)
	for _, reward := range rewardsResponse.Total {
		if value, err := strconv.ParseFloat(reward.Amount.String(), 64); err != nil {
			log.Error().
				Str("address", address).
				Err(err).
				Msg("Could not parse reward")
			return 0, err
		} else {
			rewardsTotal += value
		}
	}

	return rewardsTotal, nil
}
