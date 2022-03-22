package downloader

import (
	"context"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"

	"xyzc.dev/expirements/downloader/eth"
)

type SimpleDownloader struct {
	Client    *ethclient.Client
	RPCClient *rpc.Client
}

func (dl SimpleDownloader) Fetch(ctx context.Context, id eth.BlockID) (*types.Block, []*types.Receipt, error) {
	block, err := dl.Client.BlockByHash(ctx, id.Hash)
	if err != nil {
		return nil, nil, err
	}
	txs := block.Transactions()
	receipts := make([]*types.Receipt, len(txs))

	semaphoreChan := make(chan struct{}, 100)
	defer close(semaphoreChan)
	var wg sync.WaitGroup
	for idx, tx := range txs {
		wg.Add(1)
		i := idx
		hash := tx.Hash()
		go func() {
			semaphoreChan <- struct{}{}
			maxReceiptRetry := 3
			for j := 0; j < maxReceiptRetry; j++ {
				receipt, err := dl.Client.TransactionReceipt(ctx, hash)
				if err != nil && j == maxReceiptRetry-1 {
					panic(err)
					// return nil, nil, err
				} else if err == nil {
					receipts[i] = receipt
					break
				} else {
					time.Sleep(20 * time.Millisecond)
				}
			}
			wg.Done()
			<-semaphoreChan
		}()
	}
	wg.Wait()
	return block, receipts, nil
}

func (dl SimpleDownloader) FetchBatch(ctx context.Context, id eth.BlockID) (*types.Block, []*types.Receipt, error) {
	block, err := dl.Client.BlockByHash(ctx, id.Hash)
	if err != nil {
		return nil, nil, err
	}
	receipts, err := BatchTransactionReceipts(ctx, *dl.RPCClient, block.Transactions())
	if err != nil {
		return nil, nil, err
	}

	return block, receipts, nil
}

func BatchTransactionReceipts(ctx context.Context, client rpc.Client, transcations []*types.Transaction) ([]*types.Receipt, error) {
	if len(transcations) == 0 {
		return nil, nil
	}
	receipts := make([]*types.Receipt, len(transcations))
	reqs := make([]rpc.BatchElem, len(transcations))
	for i := range reqs {
		reqs[i] = rpc.BatchElem{
			Method: "eth_getTransactionReceipt",
			Args:   []interface{}{transcations[i].Hash()},
			Result: &receipts[i],
		}
	}

	err := client.BatchCallContext(ctx, reqs)
	if err != nil {
		return nil, err
	}
	for _, req := range reqs {
		if req.Error != nil {
			return nil, req.Error
		}
	}

	return receipts, nil
}
