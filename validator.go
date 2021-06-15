package main

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	tb "gopkg.in/tucnak/telebot.v2"
)

func getValidatorInfo(message *tb.Message) {
	args := strings.Split(message.Text, " ")
	if len(args) < 2 {
		log.Info().Msg("getWalletInfo: args length < 2")
		sendMessage(message, "Usage: validator &lt;validator operator address or name&gt;")
		return
	}

	address := args[1]
	log.Info().Str("address", address).Msg("getValidatorInfo: address")

	// --------------------------------
	validator, err := getValidator(address)
	if err != nil {
		log.Error().Err(err).Msg("Could not get validator")
		sendMessage(message, "Could not find validator")
		return
	}

	rank, err := getValidatorRank(validator)
	if err != nil {
		log.Error().Err(err).Msg("Could not get validator rank")
		sendMessage(message, "Could not find validator rank")
		return
	}

	// --------------------------------

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<code>%s</code>\n", validator.Description.Moniker))
	sb.WriteString(fmt.Sprintf("<a href=\"https://mintscan.io/%s/validators/%s\">Mintscan</a>\n\n", MintscanPrefix, validator.OperatorAddress))

	sb.WriteString(fmt.Sprintf("<strong>Moniker: </strong><code>%s</code>\n", validator.Description.Moniker))
	sb.WriteString(fmt.Sprintf("<strong>Operator address: </strong><code>%s</code>\n", validator.OperatorAddress))
	sb.WriteString(fmt.Sprintf("<strong>Description: </strong><code>%s</code>\n", validator.Description.Details))
	sb.WriteString(fmt.Sprintf("<strong>Website: </strong><code>%s</code>\n", validator.Description.Website))
	sb.WriteString(fmt.Sprintf("<strong>Security contact: </strong><code>%s</code>\n", validator.Description.SecurityContact))

	// because cosmos's dec doesn't have .toFloat64() method or whatever and returns everything as int
	if value, err := strconv.ParseFloat(validator.Commission.CommissionRates.Rate.String(), 64); err != nil {
		log.Error().
			Str("address", address).
			Err(err).
			Msg("Could not parse balance")
		sendMessage(message, "Could not parse balance")
		return
	} else {
		sb.WriteString(fmt.Sprintf("<strong>Commission rate: </strong><code>%.1f%%</code>\n", value*100))
	}

	if value, err := strconv.ParseFloat(validator.DelegatorShares.String(), 64); err != nil {
		log.Error().
			Str("address", address).
			Err(err).
			Msg("Could not parse delegator shares")
		sendMessage(message, "Could not parse delegator shares")
		return
	} else {
		sb.WriteString(fmt.Sprintf(
			"\n<strong>Total tokens delegated: </strong><code>%.1f %s</code>\n",
			value/DenomCoefficient,
			Denom,
		))
	}

	if validator.Jailed {
		sb.WriteString("<strong>Rank: </strong>JAILED\n")
	} else {
		sb.WriteString(fmt.Sprintf("<strong>Rank: </strong>%d\n", rank))
	}

	sendMessage(message, sb.String())
	log.Info().
		Str("query", address).
		Str("validator", validator.OperatorAddress).
		Msg("Successfully returned validator info")
}

func getValidator(address string) (stakingtypes.Validator, error) {
	if strings.HasPrefix(address, ValidatorPrefix) {
		log.Debug().Str("address", address).Msg("Searching validator by address")

		stakingClient := stakingtypes.NewQueryClient(grpcConn)
		validatorResponse, err := stakingClient.Validator(
			context.Background(),
			&stakingtypes.QueryValidatorRequest{ValidatorAddr: address},
		)

		if err != nil {
			log.Error().
				Str("address", address).
				Err(err).
				Msg("Could not get validator")
			return stakingtypes.Validator{}, err
		}

		return validatorResponse.Validator, nil
	}

	log.Debug().Str("address", address).Msg("Searching validator by name")

	stakingClient := stakingtypes.NewQueryClient(grpcConn)
	validatorsResponse, err := stakingClient.Validators(
		context.Background(),
		&stakingtypes.QueryValidatorsRequest{},
	)

	if err != nil {
		log.Error().
			Str("address", address).
			Err(err).
			Msg("Could not get validators")
		return stakingtypes.Validator{}, err
	}

	for _, validator := range validatorsResponse.Validators {
		if strings.Contains(
			strings.ToLower(validator.Description.Moniker),
			strings.ToLower(address),
		) {
			log.Info().Str("address", address).Str("moniker", validator.Description.Moniker).Msg("Found validator")
			return validator, nil
		}
	}

	return stakingtypes.Validator{}, fmt.Errorf("validator is not found")
}

func getValidatorRank(validator stakingtypes.Validator) (int, error) {
	stakingClient := stakingtypes.NewQueryClient(grpcConn)
	validatorsResponse, err := stakingClient.Validators(
		context.Background(),
		&stakingtypes.QueryValidatorsRequest{},
	)

	if err != nil {
		log.Error().
			Str("address", validator.OperatorAddress).
			Err(err).
			Msg("Could not get validators")
		return 0, err
	}

	validators := validatorsResponse.Validators

	sort.Slice(validators[:], func(i, j int) bool {
		return validators[i].DelegatorShares.RoundInt64() > validators[j].DelegatorShares.RoundInt64()
	})

	for index, iteratedValidator := range validators {
		if validator.OperatorAddress == iteratedValidator.OperatorAddress {
			return index + 1, nil
		}
	}

	return 0, fmt.Errorf("could not find validator rank")
}
