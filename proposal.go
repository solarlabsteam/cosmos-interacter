package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/gogo/protobuf/proto"

	tb "gopkg.in/tucnak/telebot.v2"
)

func getProposalInfo(message *tb.Message) {
	args := strings.SplitAfterN(message.Text, " ", 2)
	if len(args) < 2 {
		log.Info().Msg("getProposalInfo: args length < 2")
		sendMessage(message, "Usage: proposal &lt;proposal ID&gt;")
		return
	}

	id, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		log.Error().Err(err).Msg("Could not parse proposal ID")
		sendMessage(message, "Proposal ID should be a number")
		return
	}

	log.Debug().Uint64("id", id).Msg("getProposalInfo: id")

	// --------------------------------
	proposal, err := getProposal(id)
	if err != nil {
		log.Error().Err(err).Msg("Could not get proposal")
		sendMessage(message, "Could not find proposal")
		return
	}

	// --------------------------------

	sendMessage(message, serializeProposal(proposal))
	log.Info().
		Uint64("id", proposal.ProposalId).
		Str("user", message.Sender.Username).
		Msg("Successfully returned proposal info")
}

func serializeProposal(proposal govtypes.Proposal) string {
	title := ""
	description := ""

	switch proposal.Content.TypeUrl {
	case "/cosmos.upgrade.v1beta1.SoftwareUpgradeProposal":
		var parsedMessage upgradetypes.SoftwareUpgradeProposal
		if err := proto.Unmarshal(proposal.Content.Value, &parsedMessage); err != nil {
			log.Error().Err(err).Msg("Could not parse SoftwareUpgradeProposal")
			return "Could not parse proposal"
		} else {
			title = parsedMessage.Title
			description = parsedMessage.Description
		}
	case "/cosmos.gov.v1beta1.TextProposal":
		var parsedMessage govtypes.TextProposal
		if err := proto.Unmarshal(proposal.Content.Value, &parsedMessage); err != nil {
			log.Error().Err(err).Msg("Could not parse TextProposal")
			return "Could not parse proposal"
		} else {
			title = parsedMessage.Title
			description = parsedMessage.Description
		}
	case "/cosmos.params.v1beta1.ParameterChangeProposal":
		var parsedMessage paramstypes.ParameterChangeProposal
		if err := proto.Unmarshal(proposal.Content.Value, &parsedMessage); err != nil {
			log.Error().Err(err).Msg("Could not parse ParameterChangeProposal")
			return "Could not parse proposal"
		} else {
			title = parsedMessage.Title
			description = parsedMessage.Description
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<string>Proposal #%d</string>\n", proposal.ProposalId))
	sb.WriteString(fmt.Sprintf("<code>%s</code>\n", title))
	sb.WriteString(fmt.Sprintf("Submit time: <code>%s</code>\n", proposal.SubmitTime))
	sb.WriteString(fmt.Sprintf("Deposit time: <code>%s</code>\n", proposal.DepositEndTime))
	sb.WriteString(fmt.Sprintf("Voting starts: <code>%s</code>\n", proposal.VotingStartTime))
	sb.WriteString(fmt.Sprintf("Voting ends: <code>%s</code>\n", proposal.VotingEndTime))
	sb.WriteString(fmt.Sprintf("Stattus: <code>%s</code>\n", proposal.Status))
	sb.WriteString(fmt.Sprintf("<a href=\"https://mintscan.io/%s/proposals/%d\">Mintscan</a>\n\n", MintscanPrefix, proposal.ProposalId))

	sb.WriteString(fmt.Sprintf("<pre>%s</pre>", description))

	return sb.String()
}

func getProposal(id uint64) (govtypes.Proposal, error) {
	govClient := govtypes.NewQueryClient(grpcConn)
	proposalResponse, err := govClient.Proposal(
		context.Background(),
		&govtypes.QueryProposalRequest{ProposalId: id},
	)

	if err != nil {
		log.Error().
			Uint64("id", id).
			Err(err).
			Msg("Could not get proposal")
		return govtypes.Proposal{}, err
	}

	return proposalResponse.Proposal, nil
}
