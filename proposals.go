package main

import (
	"context"
	"fmt"
	"strings"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/gogo/protobuf/proto"

	tb "gopkg.in/tucnak/telebot.v2"
)

func getProposalsInfo(message *tb.Message) {
	log.Debug().Msg("getProposalsInfo")

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
	title := ""

	switch proposal.Content.TypeUrl {
	case "/cosmos.upgrade.v1beta1.SoftwareUpgradeProposal":
		var parsedMessage upgradetypes.SoftwareUpgradeProposal
		if err := proto.Unmarshal(proposal.Content.Value, &parsedMessage); err != nil {
			log.Error().Err(err).Msg("Could not parse SoftwareUpgradeProposal")
			return "Could not parse proposal"
		} else {
			title = parsedMessage.Title
		}
	case "/cosmos.gov.v1beta1.TextProposal":
		var parsedMessage govtypes.TextProposal
		if err := proto.Unmarshal(proposal.Content.Value, &parsedMessage); err != nil {
			log.Error().Err(err).Msg("Could not parse TextProposal")
			return "Could not parse proposal"
		} else {
			title = parsedMessage.Title
		}
	case "/cosmos.params.v1beta1.ParameterChangeProposal":
		var parsedMessage paramstypes.ParameterChangeProposal
		if err := proto.Unmarshal(proposal.Content.Value, &parsedMessage); err != nil {
			log.Error().Err(err).Msg("Could not parse ParameterChangeProposal")
			return "Could not parse proposal"
		} else {
			title = parsedMessage.Title
		}
	default:
		log.Error().Str("type", proposal.Content.TypeUrl).Msg("Unknown proposal type!")
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<strong>Proposal #%d</strong>\n", proposal.ProposalId))
	if title != "" {
		sb.WriteString(fmt.Sprintf("<code>%s</code>\n", title))
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
