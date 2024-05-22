package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/urfave/cli/v2"
	"github.com/waku-org/go-waku/waku/v2/dnsdisc"
	"github.com/waku-org/go-waku/waku/v2/node"
	"github.com/waku-org/go-waku/waku/v2/protocol"
	"go.uber.org/zap/zapcore"
)

func discoverNodes(cCtx *cli.Context, wakuNode *node.WakuNode) {
	enr := cCtx.String(BootstrapNodeFlag)
	nodes, err := dnsdisc.RetrieveNodes(cCtx.Context, enr)
	if err != nil {
		fmt.Println("Error retrieving nodes", err)
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(len(nodes))
	for _, node := range nodes {
		go func(ctx context.Context, info peer.AddrInfo) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(ctx, time.Duration(30)*time.Second)
			defer cancel()

			err = wakuNode.DialPeerWithInfo(ctx, info)
			if err != nil {
				fmt.Println("Error dialing peer", err)
				return
			}
		}(cCtx.Context, node.PeerInfo)
	}
	wg.Wait()
}

func startNode(cCtx *cli.Context) (*node.WakuNode, error) {
	contentTopic, err := protocol.NewContentTopic("universal", "1", "message-rate", "proto")
	if err != nil {
		fmt.Println("Invalid Content Topic")
		panic(err)
	}

	fmt.Println(contentTopic)

	hostAddr, _ := net.ResolveTCPAddr("tcp", "0.0.0.0:0")
	key, err := randomHex(32)
	if err != nil {
		fmt.Println("Could not generate random key")
		panic(err)
	}
	prvKey, err := crypto.HexToECDSA(key)
	if err != nil {
		fmt.Println("Could not generate private key")
		panic(err)
	}

	wakuNode, err := node.New(
		node.WithPrivateKey(prvKey),
		node.WithHostAddress(hostAddr),
		node.WithNTP(),
		node.WithWakuRelay(),
		node.WithWakuRelayAndMinPeers(1),
		node.WithClusterID(16),
		node.WithLogLevel(zapcore.DebugLevel),
	)
	if err != nil {
		fmt.Println("Error creating Waku node: ", err)
	}

	if err := wakuNode.Start(cCtx.Context); err != nil {
		fmt.Println("Could not start waku node")
		panic(err)
	}

	fmt.Println("Waku node started")

	discoverNodes(cCtx, wakuNode)

	return wakuNode, nil
}

func randomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
