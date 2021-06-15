package main

import (
	"context"
	"fmt"
	"strings"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	tb "gopkg.in/tucnak/telebot.v2"
)

func getValidatorInfo(message *tb.Message) {
	args := strings.Split(message.Text, " ")
	if len(args) < 2 {
		log.Info().Msg("getWalletInfo: args length < 2")
		return
	}

	address := args[1]
	log.Info().Str("address", address).Msg("getValidatorInfo: address")

	// --------------------------------
	validator, err := getValidator(address)
	if err != nil {
		log.Error().Err(err).Msg("Could not get validator")
		return
	}

	// --------------------------------

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<code>%s</code>\n", validator.Description.Moniker))
	sb.WriteString(fmt.Sprintf("<a href=\"https://mintscan.io/%s/validators/%s\">Mintscan</a>\n\n", MintscanPrefix, validator.OperatorAddress))

	sb.WriteString(fmt.Sprintf("<strong>Moniker: </strong>%s\n", validator.Description.Moniker))
	sb.WriteString(fmt.Sprintf("<strong>Operator address: </strong>%s\n", validator.OperatorAddress))
	sb.WriteString(fmt.Sprintf("<strong>Description: </strong>%s\n", validator.Description.Details))
	sb.WriteString(fmt.Sprintf("<strong>Website: </strong>%s\n", validator.Description.Website))
	sb.WriteString(fmt.Sprintf("<strong>Security contact: </strong>%s\n", validator.Description.SecurityContact))

	bot.Send(
		message.Chat,
		sb.String(),
		&tb.SendOptions{
			ParseMode: tb.ModeHTML,
		},
	)
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

	return stakingtypes.Validator{}, fmt.Errorf("Validator is not found")
}
