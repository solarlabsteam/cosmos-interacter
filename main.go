package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	sdk "github.com/cosmos/cosmos-sdk/types"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
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

	grpcConn *grpc.ClientConn

	log = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()

	bot tgbotapi.BotAPI
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

	grpcConn, err := grpc.Dial(
		NodeAddress,
		grpc.WithInsecure(),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not connect to gRPC node")
	}

	defer grpcConn.Close()

	bot, err := tgbotapi.NewBotAPI(TelegramToken)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not connect to Telegram")
	}

	bot.Debug = true

	log.Info().Str("username", bot.Self.UserName).Msg("Authorized on account")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not get updates")
	}

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		log.Info().
			Str("from", update.Message.From.UserName).
			Str("text", update.Message.Text).
			Msg("Got message")

		switch {
		case strings.HasPrefix(update.Message.Text, "/wallet"):
			getWalletInfo(update.Message.Text)
			return
		}

		// msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		// msg.ReplyToMessageID = update.Message.MessageID

		// bot.Send(msg)
	}
}

func getWalletInfo(text string) {
	args := strings.Split(text, " ")
	if len(args) < 2 {
		log.Info().Msg("getWalletInfo: args length < 2")
		return
	}

	address := args[1]
	log.Info().Str("address", address).Msg("getWalletInfo: address")
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
