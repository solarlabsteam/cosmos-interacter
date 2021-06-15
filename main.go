package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"google.golang.org/grpc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	tb "gopkg.in/tucnak/telebot.v2"
)

var (
	ConfigPath     string
	NodeAddress    string
	LogLevel       string
	MintscanPrefix string

	TelegramToken string
	TelegramChat  int

	Prefix                    string
	AccountPrefix             string
	AccountPubkeyPrefix       string
	ValidatorPrefix           string
	ValidatorPubkeyPrefix     string
	ConsensusNodePrefix       string
	ConsensusNodePubkeyPrefix string

	Denom            string
	DenomCoefficient float64

	grpcConn *grpc.ClientConn

	log = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()

	bot *tb.Bot

	Printer = message.NewPrinter(language.English)
)

var rootCmd = &cobra.Command{
	Use:  "missed-blocks-checker",
	Long: "Tool to monitor missed blocks for Cosmos-chain validators",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if ConfigPath == "" {
			log.Trace().Msg("No config file provided, skipping")
			setBechPrefixes(cmd)
			return nil
		}

		log.Trace().Msg("Config file provided")

		viper.SetConfigFile(ConfigPath)
		if err := viper.ReadInConfig(); err != nil {
			log.Info().Err(err).Msg("Error reading config file")
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return err
			}
		}

		// Credits to https://carolynvanslyck.com/blog/2020/08/sting-of-the-viper/
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			if !f.Changed && viper.IsSet(f.Name) {
				val := viper.Get(f.Name)
				if err := cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val)); err != nil {
					log.Fatal().Err(err).Msg("Could not set flag")
				}
			}
		})

		setBechPrefixes(cmd)

		return nil
	},
	Run: Execute,
}

func setBechPrefixes(cmd *cobra.Command) {
	if flag, err := cmd.Flags().GetString("bech-account-prefix"); flag != "" && err == nil {
		AccountPrefix = flag
	} else {
		AccountPrefix = Prefix
	}

	if flag, err := cmd.Flags().GetString("bech-account-pubkey-prefix"); flag != "" && err == nil {
		AccountPubkeyPrefix = flag
	} else {
		AccountPubkeyPrefix = Prefix + "pub"
	}

	if flag, err := cmd.Flags().GetString("bech-validator-prefix"); flag != "" && err == nil {
		ValidatorPrefix = flag
	} else {
		ValidatorPrefix = Prefix + "valoper"
	}

	if flag, err := cmd.Flags().GetString("bech-validator-pubkey-prefix"); flag != "" && err == nil {
		ValidatorPubkeyPrefix = flag
	} else {
		ValidatorPubkeyPrefix = Prefix + "valoperpub"
	}

	if flag, err := cmd.Flags().GetString("bech-consensus-node-prefix"); flag != "" && err == nil {
		ConsensusNodePrefix = flag
	} else {
		ConsensusNodePrefix = Prefix + "valcons"
	}

	if flag, err := cmd.Flags().GetString("bech-consensus-node-pubkey-prefix"); flag != "" && err == nil {
		ConsensusNodePubkeyPrefix = flag
	} else {
		ConsensusNodePubkeyPrefix = Prefix + "valconspub"
	}
}

func setDenom() {
	bankClient := banktypes.NewQueryClient(grpcConn)
	denoms, err := bankClient.DenomsMetadata(
		context.Background(),
		&banktypes.QueryDenomsMetadataRequest{},
	)

	if err != nil {
		log.Fatal().Err(err).Msg("Error querying denom")
	}

	metadata := denoms.Metadatas[0] // always using the first one
	if Denom == "" {                // using display currency
		Denom = metadata.Display
	}

	for _, unit := range metadata.DenomUnits {
		log.Debug().
			Str("denom", unit.Denom).
			Uint32("exponent", unit.Exponent).
			Msg("Denom info")
		if unit.Denom == Denom {
			DenomCoefficient = math.Pow10(int(unit.Exponent))
			log.Info().
				Str("denom", Denom).
				Float64("coefficient", DenomCoefficient).
				Msg("Got denom info")
			return
		}
	}

	log.Fatal().Msg("Could not find the denom info")
}

func Execute(cmd *cobra.Command, args []string) {
	logLevel, err := zerolog.ParseLevel(LogLevel)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not parse log level")
	}

	zerolog.SetGlobalLevel(logLevel)

	config := sdk.GetConfig()
	config.SetBech32PrefixForValidator(ValidatorPrefix, ValidatorPubkeyPrefix)
	config.SetBech32PrefixForConsensusNode(ConsensusNodePrefix, ConsensusNodePubkeyPrefix)
	config.Seal()

	grpcConn, err = grpc.Dial(
		NodeAddress,
		grpc.WithInsecure(),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not connect to gRPC node")
	}

	defer grpcConn.Close()

	setDenom()

	bot, err = tb.NewBot(tb.Settings{
		Token:   TelegramToken,
		Poller:  &tb.LongPoller{Timeout: 10 * time.Second},
		Verbose: true,
	})

	if err != nil {
		log.Fatal().Err(err).Msg("Could not create bot")
	}

	bot.Handle("/wallet", getWalletInfo)
	bot.Start()
}

func getWalletInfo(message *tb.Message) {
	args := strings.Split(message.Text, " ")
	if len(args) < 2 {
		log.Info().Msg("getWalletInfo: args length < 2")
		return
	}

	address := args[1]
	log.Info().Str("address", address).Msg("getWalletInfo: address")

	bankClient := banktypes.NewQueryClient(grpcConn)

	// --------------------------------
	balancesResponse, err := bankClient.AllBalances(
		context.Background(),
		&banktypes.QueryAllBalancesRequest{Address: address},
	)

	if err != nil {
		log.Error().
			Str("address", address).
			Err(err).
			Msg("Could not get balance")
		return
	}

	delegationsTotal, err := getTotalDelegations(address)
	if err != nil {
		log.Error().
			Str("address", address).
			Err(err).
			Msg("Could not get delegations")
		return
	}

	unbondingsTotal, err := getTotalUnbondings(address)
	if err != nil {
		log.Error().
			Str("address", address).
			Err(err).
			Msg("Could not get unbondings")
		return
	}

	rewardsTotal, err := getTotalRewards(address)
	if err != nil {
		log.Error().
			Str("address", address).
			Err(err).
			Msg("Could not get rewards")
		return
	}

	// --------------------------------

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<code>%s</code>\n", address))
	sb.WriteString(fmt.Sprintf("<a href=\"https://mintscan.io/%s/account/%s\">Mintscan</a>\n\n", MintscanPrefix, address))

	sb.WriteString(fmt.Sprintf("<strong>Balance:        </strong>"))

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

	bot.Send(
		message.Chat,
		sb.String(),
		&tb.SendOptions{
			ParseMode: tb.ModeHTML,
		},
	)
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

func main() {
	rootCmd.PersistentFlags().StringVar(&ConfigPath, "config", "", "Config file path")
	rootCmd.PersistentFlags().StringVar(&LogLevel, "log-level", "info", "Logging level")
	rootCmd.PersistentFlags().StringVar(&NodeAddress, "node", "localhost:9090", "RPC node address")
	rootCmd.PersistentFlags().StringVar(&MintscanPrefix, "mintscan-prefix", "persistence", "Prefix for mintscan links like https://mintscan.io/{prefix}")

	rootCmd.PersistentFlags().StringVar(&TelegramToken, "telegram-token", "", "Telegram bot token")
	rootCmd.PersistentFlags().IntVar(&TelegramChat, "telegram-chat", 0, "Telegram chat or user ID")

	// some networks, like Iris, have the different prefixes for address, validator and consensus node
	rootCmd.PersistentFlags().StringVar(&Prefix, "bech-prefix", "persistence", "Bech32 global prefix")
	rootCmd.PersistentFlags().StringVar(&ValidatorPrefix, "bech-validator-prefix", "", "Bech32 validator prefix")
	rootCmd.PersistentFlags().StringVar(&ValidatorPubkeyPrefix, "bech-validator-pubkey-prefix", "", "Bech32 pubkey validator prefix")
	rootCmd.PersistentFlags().StringVar(&ConsensusNodePrefix, "bech-consensus-node-prefix", "", "Bech32 consensus node prefix")
	rootCmd.PersistentFlags().StringVar(&ConsensusNodePubkeyPrefix, "bech-consensus-node-pubkey-prefix", "", "Bech32 pubkey consensus node prefix")

	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Could not start application")
	}
}
