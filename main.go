package main

import (
	"os"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

const BootstrapNode = "enrtree://AMOJVZX4V6EXP7NTJPMAYJYST2QP6AJXYW76IU6VGJS7UVSNDYZG4@boot.test.shards.nodes.status.im"
const ContentTopic = "/universal/1/message-rate/proto"
const PubSubTopic = "/waku/2/rs/16/32"

const BootstrapNodeFlag = "bootstrap"
const ContentTopicFlag = "contentTopic"
const PubSubTopicFlag = "pubsubTopic"

var CommonFlags = []cli.Flag{
	&cli.StringFlag{
		Name:  "bootstrap",
		Value: BootstrapNode,
		Usage: "Bootstramp enrtre record",
	},
	&cli.StringFlag{
		Name:  "contentTopic",
		Value: ContentTopic,
		Usage: "The content topic where messages are sent",
	},
	&cli.StringFlag{
		Name:  "pubsubTopic",
		Value: PubSubTopic,
		Usage: "The pubsub topic for relay messages",
	},
}

var WriteFlags = append([]cli.Flag{}, CommonFlags...)

var ReadFlags = append([]cli.Flag{}, CommonFlags...)

var logger *zap.SugaredLogger

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "write",
				Aliases: []string{"w"},
				Usage:   "Write to content topic",
				Flags:   WriteFlags,
				Action: func(cCtx *cli.Context) error {
					return startWriter(cCtx)
				},
			},
			{
				Name:    "read",
				Aliases: []string{"r"},
				Usage:   "Read messages of content topic",
				Flags:   ReadFlags,
				Action: func(cCtx *cli.Context) error {
					return read(cCtx)
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger.Fatal(err)
	}
}
