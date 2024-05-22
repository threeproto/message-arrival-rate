package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"github.com/urfave/cli/v2"
	"github.com/waku-org/go-waku/logging"
	"github.com/waku-org/go-waku/waku/v2/node"
	"github.com/waku-org/go-waku/waku/v2/payload"
	"github.com/waku-org/go-waku/waku/v2/protocol/pb"
	"github.com/waku-org/go-waku/waku/v2/protocol/relay"
	"github.com/waku-org/go-waku/waku/v2/utils"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

var log = utils.Logger().Named("message-writer")

func startWriter(cCtx *cli.Context) error {
	wakuNode, err := startNode(cCtx)
	if err != nil {
		return err
	}

	senderFile, err := os.OpenFile("sender.csv", os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	defer senderFile.Close()

	writeLoop(cCtx, wakuNode, senderFile)

	return nil
}

func writeLoop(cCtx *cli.Context, wakuNode *node.WakuNode, senderFile *os.File) {

	sendWriter := csv.NewWriter(senderFile)
	defer sendWriter.Flush()

	for {
		time.Sleep(10 * time.Second)
		write(cCtx, wakuNode, sendWriter)
	}
}

func write(cCtx *cli.Context, wakuNode *node.WakuNode, sendWriter *csv.Writer) {
	var version uint32 = 0

	p := new(payload.Payload)
	p.Data = []byte(wakuNode.ID() + ": " + "hello world")
	p.Key = &payload.KeyInfo{Kind: payload.None}

	payload, err := p.Encode(version)
	if err != nil {
		fmt.Println("Error encoding the payload", zap.Error(err))
		return
	}

	msg := &pb.WakuMessage{
		Payload:      payload,
		Version:      proto.Uint32(version),
		ContentTopic: cCtx.String(ContentTopicFlag),
		Timestamp:    utils.GetUnixEpoch(wakuNode.Timesource()),
	}

	hash, err := wakuNode.Relay().Publish(cCtx.Context, msg, relay.WithPubSubTopic(cCtx.String(PubSubTopicFlag)))
	if err != nil {
		log.Error("Error sending a message", zap.Error(err))
		return
	}

	log.Info("Published msg,", zap.String("data", string(msg.Payload)), logging.HexBytes("hash", hash.Bytes()))

	err = sendWriter.Write([]string{hash.String()})
	if err != nil {
		log.Error("Error writing to csv", zap.Error(err))
	}
	sendWriter.Flush()
}
