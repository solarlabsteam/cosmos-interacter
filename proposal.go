package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

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

	serializedProposal, err := serializeProposal(proposal)
	if err != nil {
		sendMessage(message, "Error getting proposal")
		log.Info().
			Uint64("id", proposal.ProposalId).
			Str("user", message.Sender.Username).
			Msg("Successfully returned proposal info")
	} else {
		sendMessage(message, serializedProposal)

	}
}

func serializeProposal(proposal govtypes.Proposal) (string, error) {
	proposalInfo, err := getProposalInfoAsStruct(proposal)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<strong>Proposal #%d</strong>\n", proposal.ProposalId))
	sb.WriteString(fmt.Sprintf("<code>%s</code>\n", proposalInfo.Title))
	sb.WriteString(fmt.Sprintf("Submit time:   <code>%s</code>\n", proposal.SubmitTime.Format(time.RFC822)))
	sb.WriteString(fmt.Sprintf("Deposit time:  <code>%s</code>\n", proposal.DepositEndTime.Format(time.RFC822)))
	sb.WriteString(fmt.Sprintf("Voting starts: <code>%s</code>\n", proposal.VotingStartTime.Format(time.RFC822)))
	sb.WriteString(fmt.Sprintf("Voting ends:   <code>%s</code>\n", proposal.VotingEndTime.Format(time.RFC822)))
	sb.WriteString(fmt.Sprintf("Status: <code>%s</code>\n", proposal.Status))
	sb.WriteString(fmt.Sprintf("<a href=\"https://mintscan.io/%s/proposals/%d\">Mintscan</a>\n\n", MintscanPrefix, proposal.ProposalId))

	sb.WriteString(fmt.Sprintf("<pre>%s</pre>", proposalInfo.Description))

	return sb.String(), nil
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
