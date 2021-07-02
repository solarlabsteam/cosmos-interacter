module main

go 1.16

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1

require (
	github.com/cosmos/cosmos-sdk v0.42.5
	github.com/gogo/protobuf v1.3.3
	github.com/rs/zerolog v1.23.0
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/superoo7/go-gecko v1.0.0
	golang.org/x/text v0.3.3
	google.golang.org/grpc v1.38.0
	gopkg.in/tucnak/telebot.v2 v2.3.5
)
