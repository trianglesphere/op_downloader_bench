package main

import (
	"context"
	"fmt"
	"math/big"
	"time"

	dl "xyzc.dev/expirements/downloader/downloader"
	"xyzc.dev/expirements/downloader/eth"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

/*
	l1Sources := make([]eth.L1Source, 0, len(c.L1NodeAddrs))
	for i, addr := range c.L1NodeAddrs {
		// L1 exec engine: read-only, to update L2 consensus with
		l1Node, err := rpc.DialContext(ctx, addr)
		if err != nil {
			// HTTP or WS RPC may create a disconnected client, RPC over IPC may fail directly
			if l1Node == nil {
				return fmt.Errorf("failed to dial L1 address %d (%s): %v", i, addr, err)
			}
			c.log.Warn("failed to dial L1 address, but may connect later", "i", i, "addr", addr, "err", err)
		}
		// TODO: we may need to authenticate the connection with L1
		// l1Node.SetHeader()
		cl := ethclient.NewClient(l1Node)
		l1Sources = append(l1Sources, cl)
	}
	if len(l1Sources) == 0 {
		return fmt.Errorf("need at least one L1 source endpoint, see --l1")
	}

	// Combine L1 sources, so work can be balanced between them
	c.l1Source = eth.NewCombinedL1Source(l1Sources)
	l1CanonicalChain := eth.CanonicalChain(c.l1Source)

	c.l1Downloader = l1.NewDownloader(c.l1Source)
*/

func main() {
	httpAddr := "http://127.0.0.1:8545"
	// wsAddr := "ws://127.0.0.1:8546"
	// ipcAddr := "/Users/joshua/Library/Ethereum/geth.ipc"
	ctx := context.Background()
	node, err := rpc.DialContext(ctx, httpAddr)
	if err != nil {
		panic(err)
	}
	client := ethclient.NewClient(node)

	// source := eth.NewCombinedL1Source([]eth.L1Source{client})
	// downloader := dl.NewDownloader(source)
	// downloader.AddReceiptWorkers(4)
	downloader := dl.SimpleDownloader{Client: client, RPCClient: node}

	// 14027000

	startNumber := 14027000
	for i := 1000; i >= 0; i-- {
		header, err := client.HeaderByNumber(ctx, big.NewInt(int64(startNumber-i)))
		if err != nil {
			panic(err)
		}
		start := time.Now()
		_, _, err = downloader.Fetch(ctx, eth.BlockID{Hash: header.Hash(), Number: uint64(startNumber - i)})
		elapsed := time.Since(start)
		fmt.Println(elapsed)
		if err != nil {
			panic(err)
		}
	}

	n, err := client.BlockNumber(ctx)
	fmt.Println(n, err)
}
