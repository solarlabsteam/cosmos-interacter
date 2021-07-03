package main

import (
	"context"
	"fmt"
	"strings"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	tb "gopkg.in/tucnak/telebot.v2"
)

func getProposalsInfo(message *tb.Message) {
	// --------------------------------
	proposals, err := getProposals()
	if err != nil {
		log.Error().Err(err).Msg("Could not get proposals")
		sendMessage(message, "Could not get proposals")
		return
	}

	// --------------------------------

	var sb strings.Builder
	for _, proposal := range proposals {
		sb.WriteString(serializeProposalShort(proposal) + "\n\n")
	}

	sendMessage(message, sb.String())
	log.Info().
		Str("user", message.Sender.Username).
		Msg("Successfully returned proposals info")
}

func serializeProposalShort(proposal govtypes.Proposal) string {
	proposalInfo, err := getProposalInfoAsStruct(proposal)
	if err != nil {
		log.Error().Err(err).Msg("Could not parse proposal")
		return ""
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<strong>Proposal #%d</strong>\n", proposal.ProposalId))
	if proposalInfo.Title != "" {
		sb.WriteString(fmt.Sprintf("<code>%s</code>\n", proposalInfo.Title))
	}
	sb.WriteString(fmt.Sprintf("Status: <code>%s</code>\n", proposal.Status))
	sb.WriteString(fmt.Sprintf("<a href=\"https://mintscan.io/%s/proposals/%d\">Mintscan</a>\n", MintscanPrefix, proposal.ProposalId))
	sb.WriteString(fmt.Sprintf("More info: <code>/proposal %d</code>", proposal.ProposalId))

	return sb.String()
}

func getProposals() (govtypes.Proposals, error) {
	govClient := govtypes.NewQueryClient(grpcConn)
	proposalResponse, err := govClient.Proposals(
		context.Background(),
		&govtypes.QueryProposalsRequest{},
	)

	if err != nil {
		log.Error().
			Err(err).
			Msg("Could not get proposals")
		return govtypes.Proposals{}, err
	}

	return proposalResponse.Proposals, nil
}
