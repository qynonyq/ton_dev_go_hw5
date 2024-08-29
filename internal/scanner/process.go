package scanner

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"golang.org/x/sync/errgroup"
	"gopkg.in/tomb.v2"

	"github.com/qynonyq/ton_dev_go_hw5/internal/app"
	"github.com/qynonyq/ton_dev_go_hw5/internal/storage"
)

func (s *Scanner) processBlocks(ctx context.Context) {
	const (
		delayBase = 2 * time.Second
		delayMax  = 8 * time.Second
	)
	delay := delayBase

	for {
		master, err := s.api.LookupBlock(
			ctx,
			s.lastBlock.Workchain,
			s.lastBlock.Shard,
			s.lastBlock.SeqNo,
		)
		if err == nil {
			delay = delayBase
		}
		if err != nil {
			if !errors.Is(err, ton.ErrBlockNotFound) {
				logrus.Errorf("[SCN] failed to lookup master block %d: %s", s.lastBlock.SeqNo, err)
			}

			time.Sleep(delay)
			delay *= 2
			if delay > delayMax {
				delay = delayMax
			}

			continue
		}

		err = s.processMcBlock(ctx, master)
		if err == nil {
			delay = delayBase
		}
		if err != nil {
			if !strings.Contains(err.Error(), "is not in db") {
				logrus.Errorf("[SCN] failed to process MC block [seqno=%d] [shard=%d]: %s",
					master.SeqNo, master.Shard, err)
				continue
			}

			time.Sleep(delay)
			delay *= 2
			if delay > delayMax {
				delay = delayMax
			}
		}
	}
}

func (s *Scanner) processMcBlock(ctx context.Context, master *ton.BlockIDExt) error {
	start := time.Now()

	currentShards, err := s.api.GetBlockShardsInfo(ctx, master)
	if err != nil {
		return err
	}

	shards := make(map[string]*ton.BlockIDExt, len(currentShards))

	for _, shard := range currentShards {
		// unique key
		key := fmt.Sprintf("%d:%d:%d", shard.Workchain, shard.Shard, shard.SeqNo)
		shards[key] = shard

		if err := s.fillWithNotSeenShards(ctx, shards, shard); err != nil {
			return err
		}
		s.lastShardsSeqNo[s.getShardID(shard)] = shard.SeqNo
	}

	txs := make([]*tlb.Transaction, 0, len(shards))
	for _, shard := range shards {
		shardTxs, err := s.getTxsFromShard(ctx, shard)
		if err != nil {
			return err
		}
		txs = append(txs, shardTxs...)
	}

	var (
		tmb tomb.Tomb
		wg  sync.WaitGroup
		mu  sync.Mutex
		//events = make([]storage.DedustSwap, 0, len(txs))
		//events = make([]storage.DedustDeposit, 0, len(txs))
		//events = make([]storage.DedustWithdrawal, 0, len(txs))
		events = make([]storage.StonfiSwap, 0, len(txs))
	)
	// process transactions
	tmb.Go(func() error {
		for _, tx := range txs {
			// break loop if there was transaction processing error
			select {
			case <-tmb.Dying():
				break
			default:
			}

			wg.Add(1)
			go func() {
				defer wg.Done()

				//ee, err := processTxDedustSwap(tx)
				//ee, err := processTxDedustDeposit(tx)
				//ee, err := processTxDedustWithdrawal(tx)
				e, err := s.processTxStonfiSwap(tx)
				if err != nil {
					tmb.Kill(err)
					return
				}

				if e != nil {
					mu.Lock()
					//events = append(events, ee...)
					events = append(events, *e)
					mu.Unlock()
				}
			}()
		}
		wg.Wait()

		return nil
	})
	if err := tmb.Wait(); err != nil {
		logrus.Errorf("[SCN] failed to process transactions: %s", err)
		// start with next block, otherwise process will get stuck
		s.lastBlock.SeqNo++
		return err
	}

	txDB := app.DB.Begin()

	for _, e := range events {
		if err := txDB.Create(&e).Error; err != nil {
			txDB.Rollback()
			return err
		}
	}

	if err := s.addBlock(master, txDB); err != nil {
		txDB.Rollback()
		return err
	}

	if err := txDB.Commit().Error; err != nil {
		logrus.Errorf("[SCN] failed to commit txDB: %s", err)
		return err
	}

	lastSeqno, err := s.getLastBlockSeqno(ctx)
	if err != nil {
		logrus.Infof("[SCN] block [%d] processed in [%.2fs] with [%d] transactions",
			master.SeqNo,
			time.Since(start).Seconds(),
			len(txs),
		)
	} else {
		logrus.Infof("[SCN] block [%d|%d] processed in [%.2fs] with [%d] transactions",
			master.SeqNo,
			lastSeqno,
			time.Since(start).Seconds(),
			len(txs),
		)
	}

	return nil
}
func (s *Scanner) getTxsFromShard(ctx context.Context, shard *ton.BlockIDExt) ([]*tlb.Transaction, error) {
	var (
		after    *ton.TransactionID3
		more     = true
		err      error
		eg       errgroup.Group
		txsShort []ton.TransactionShortInfo
		mu       sync.Mutex
		txs      []*tlb.Transaction
	)

	for more {
		txsShort, more, err = s.api.GetBlockTransactionsV2(
			ctx,
			shard,
			100,
			after,
		)
		if err != nil {
			return nil, err
		}

		if more {
			after = txsShort[len(txsShort)-1].ID3()
		}

		for _, txShort := range txsShort {
			eg.Go(func() error {
				tx, err := s.api.GetTransaction(
					ctx,
					shard,
					address.NewAddress(0, 0, txShort.Account),
					txShort.LT,
				)
				if err != nil {
					if strings.Contains(err.Error(), "is not in db") ||
						strings.Contains(err.Error(), "is not applied") {
						return nil
					}

					logrus.Errorf("[SCN] failed to load tx: %s", err)

					return err
				}

				mu.Lock()
				defer mu.Unlock()
				txs = append(txs, tx)

				return nil
			})
		}
	}

	if err := eg.Wait(); err != nil {
		return nil, fmt.Errorf("[SCN] failed to get transactions: %w", err)
	}

	return txs, nil
}
