package main

import (
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/gogo/protobuf/proto"
)

type ProposalInfo struct {
	Title       string
	Description string
}

func getProposalInfoAsStruct(proposal govtypes.Proposal) (ProposalInfo, error) {
	switch proposal.Content.TypeUrl {
	case "/cosmos.upgrade.v1beta1.SoftwareUpgradeProposal":
		var parsedMessage upgradetypes.SoftwareUpgradeProposal
		if err := proto.Unmarshal(proposal.Content.Value, &parsedMessage); err != nil {
			log.Error().Err(err).Msg("Could not parse SoftwareUpgradeProposal")
			return ProposalInfo{}, err
		} else {
			return ProposalInfo{Title: parsedMessage.Title, Description: parsedMessage.Description}, nil
		}
	case "/cosmos.gov.v1beta1.TextProposal":
		var parsedMessage govtypes.TextProposal
		if err := proto.Unmarshal(proposal.Content.Value, &parsedMessage); err != nil {
			log.Error().Err(err).Msg("Could not parse TextProposal")
			return ProposalInfo{}, err
		} else {
			return ProposalInfo{Title: parsedMessage.Title, Description: parsedMessage.Description}, nil

		}
	case "/cosmos.params.v1beta1.ParameterChangeProposal":
		var parsedMessage paramstypes.ParameterChangeProposal
		if err := proto.Unmarshal(proposal.Content.Value, &parsedMessage); err != nil {
			log.Error().Err(err).Msg("Could not parse ParameterChangeProposal")
			return ProposalInfo{}, err
		} else {
			return ProposalInfo{Title: parsedMessage.Title, Description: parsedMessage.Description}, nil
		}
	case "/cosmos.distribution.v1beta1.CommunityPoolSpendProposal":
		var parsedMessage distributiontypes.CommunityPoolSpendProposal
		if err := proto.Unmarshal(proposal.Content.Value, &parsedMessage); err != nil {
			log.Error().Err(err).Msg("Could not parse CommunityPoolSpendProposal")
			return ProposalInfo{}, err
		} else {
			return ProposalInfo{Title: parsedMessage.Title, Description: parsedMessage.Description}, nil

		}
	}

	return ProposalInfo{
		Title:       "Unsupported proposal type",
		Description: "Unsupported proposal type",
	}, nil
}
