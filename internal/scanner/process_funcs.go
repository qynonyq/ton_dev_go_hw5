package scanner

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"

	"github.com/qynonyq/ton_dev_go_hw5/internal/storage"
	"github.com/qynonyq/ton_dev_go_hw5/internal/structures"
)

func processTxDedustSwap(tx *tlb.Transaction) ([]storage.DedustSwap, error) {
	if tx.IO.Out == nil {
		return nil, nil
	}

	mmOut, err := tx.IO.Out.ToSlice()
	if err != nil {
		return nil, nil
	}

	events := make([]storage.DedustSwap, 0, len(mmOut))

	for _, m := range mmOut {
		if m.MsgType != tlb.MsgTypeExternalOut {
			continue
		}

		extOut := m.AsExternalOut()
		if extOut.Body == nil {
			continue
		}

		var des structures.DedustEventSwap
		if err := tlb.LoadFromCell(&des, extOut.Body.BeginParse()); err != nil {
			continue
		}

		var (
			amountIn  string
			amountOut string
		)

		// assetIn
		if des.AssetIn.Type() == "native" {
			amountIn = des.AmountIn.String() + " TON"
		} else {
			jettonAddr := des.AssetIn.AsJetton()
			amountIn = fmt.Sprintf("%s JETTON root [%s]",
				des.AmountIn,
				address.NewAddress(0, byte(jettonAddr.WorkchainID), jettonAddr.AddressData))
		}

		// assetOut
		if des.AssetOut.Type() == "native" {
			amountOut = des.AmountOut.String() + " TON"
		} else {
			jettonAddr := des.AssetOut.AsJetton()
			amountOut = fmt.Sprintf("%s JETTON [%s]",
				des.AmountOut,
				address.NewAddress(0, byte(jettonAddr.WorkchainID), jettonAddr.AddressData))
		}

		logrus.Infof("[DDST] new swap")
		logrus.Infof("[DDST] from: %s", des.ExtraInfo.SenderAddr.String())
		logrus.Infof("[DDST] amount input: %s", amountIn)
		logrus.Infof("[DDST] amount output: %s\n\n", amountOut)

		swap := storage.DedustSwap{
			PoolAddress:   extOut.SrcAddr.String(),
			AssetIn:       des.AssetIn.Type(),
			AmountIn:      decimal.NewFromBigInt(des.AmountIn.Nano(), 0),
			AssetOut:      des.AssetOut.Type(),
			AmountOut:     decimal.NewFromBigInt(des.AmountOut.Nano(), 0),
			SenderAddress: des.ExtraInfo.SenderAddr.String(),
			Reserve0:      decimal.NewFromBigInt(des.ExtraInfo.Reserve0.Nano(), 0),
			Reserve1:      decimal.NewFromBigInt(des.ExtraInfo.Reserve1.Nano(), 0),
			CreatedAt:     time.Unix(int64(extOut.CreatedAt), 0),
			ProcessedAt:   time.Now(),
		}

		events = append(events, swap)
	}

	return events, nil
}

func processTxDedustDeposit(tx *tlb.Transaction) ([]storage.DedustDeposit, error) {
	if tx.IO.Out == nil {
		return nil, nil
	}

	mmOut, err := tx.IO.Out.ToSlice()
	if err != nil {
		return nil, nil
	}

	events := make([]storage.DedustDeposit, 0, len(mmOut))

	for _, m := range mmOut {
		if m.MsgType != tlb.MsgTypeExternalOut {
			continue
		}

		extOut := m.AsExternalOut()
		if extOut.Body == nil {
			continue
		}

		var ded structures.DedustEventDeposit
		if err := tlb.LoadFromCell(&ded, extOut.Body.BeginParse()); err != nil {
			continue
		}

		logrus.Infof("[DDST] new deposit")
		logrus.Infof("[DDST] from: %s", ded.SenderAddr)
		logrus.Infof("[DDST] amount0: %s, amount1: %s", ded.Amount0, ded.Amount1)
		logrus.Infof("[DDST] reserve0: %s, reserve1: %s", ded.Reserve0, ded.Reserve1)
		logrus.Infof("[DDST] liquidity: %s\n\n", ded.Liquidity)

		deposit := storage.DedustDeposit{
			SenderAddress: ded.SenderAddr.String(),
			Amount0:       decimal.NewFromBigInt(ded.Amount0.Nano(), 0),
			Amount1:       decimal.NewFromBigInt(ded.Amount1.Nano(), 0),
			Reserve0:      decimal.NewFromBigInt(ded.Reserve0.Nano(), 0),
			Reserve1:      decimal.NewFromBigInt(ded.Reserve1.Nano(), 0),
			Liquidity:     decimal.NewFromBigInt(ded.Liquidity.Nano(), 0),
			CreatedAt:     time.Unix(int64(extOut.CreatedAt), 0),
			ProcessedAt:   time.Now(),
		}

		events = append(events, deposit)
	}

	return events, nil
}

func processTxDedustWithdrawal(tx *tlb.Transaction) ([]storage.DedustWithdrawal, error) {
	if tx.IO.Out == nil {
		return nil, nil
	}

	mmOut, err := tx.IO.Out.ToSlice()
	if err != nil {
		return nil, nil
	}

	events := make([]storage.DedustWithdrawal, 0, len(mmOut))

	for _, m := range mmOut {
		if m.MsgType != tlb.MsgTypeExternalOut {
			continue
		}

		extOut := m.AsExternalOut()
		if extOut.Body == nil {
			continue
		}

		var dew structures.DedustEventWithdrawal
		if err := tlb.LoadFromCell(&dew, extOut.Body.BeginParse()); err != nil {
			continue
		}

		logrus.Infof("[DDST] new withdrawal")
		logrus.Infof("[DDST] from: %s", dew.SenderAddr)
		logrus.Infof("[DDST] liquidity: %s", dew.Liquidity)
		logrus.Infof("[DDST] amount0: %s, amount1: %s", dew.Amount0, dew.Amount1)
		logrus.Infof("[DDST] reserve0: %s, reserve1: %s\n\n", dew.Reserve0, dew.Reserve1)

		withdrawal := storage.DedustWithdrawal{
			SenderAddress: dew.SenderAddr.String(),
			Liquidity:     decimal.NewFromBigInt(dew.Liquidity.Nano(), 0),
			Amount0:       decimal.NewFromBigInt(dew.Amount0.Nano(), 0),
			Amount1:       decimal.NewFromBigInt(dew.Amount1.Nano(), 0),
			Reserve0:      decimal.NewFromBigInt(dew.Reserve0.Nano(), 0),
			Reserve1:      decimal.NewFromBigInt(dew.Reserve1.Nano(), 0),
			CreatedAt:     time.Unix(int64(extOut.CreatedAt), 0),
			ProcessedAt:   time.Now(),
		}

		events = append(events, withdrawal)
	}

	return events, nil
}

func (s *Scanner) processTxStonfiSwap(tx *tlb.Transaction) (*storage.StonfiSwap, error) {
	var (
		stonfiPart1 structures.StonfiSwap
		stonfiPart2 structures.StonfiPaymentRequest
	)

	if tx.IO.In.MsgType != tlb.MsgTypeInternal {
		return nil, nil
	}

	// in message
	mIn := tx.IO.In.AsInternal()

	if mIn.Body == nil {
		return nil, nil
	}

	// check struct correctness first
	if err := tlb.LoadFromCell(&stonfiPart1, mIn.Body.BeginParse()); err != nil {
		return nil, nil
	}

	// check that transaction comes from stonfi_router
	if mIn.SrcAddr.String() != address.MustParseAddr("EQB3ncyBUTjZUA5EnFKR5_EnOMI9V1tTEAAPaiU71gc4TiUt").String() {
		logrus.Warningf("[STON.FI] found unknown router address: %s", mIn.SrcAddr.String())
		return nil, nil
	}

	// out messages
	if tx.IO.Out == nil {
		return nil, nil
	}

	mmOut, err := tx.IO.Out.ToSlice()
	if err != nil {
		return nil, nil
	}

	for _, mOut := range mmOut {
		if mOut.MsgType != tlb.MsgTypeInternal {
			continue
		}

		mIn := mOut.AsInternal()
		if mIn.Body == nil {
			continue
		}

		if err := tlb.LoadFromCell(&stonfiPart2, mIn.Body.BeginParse()); err != nil {
			continue
		}

		// break if real payment request found, not the referral one
		if stonfiPart2.OwnerAddress.String() == stonfiPart1.ToAddress.String() {
			break
		}
	}

	var (
		amountIn  string
		tokenIn   string
		amountOut string
		tokenOut  string
	)

	if stonfiPart2.RefData.Amount0Out.String() == "0" {
		amountIn = stonfiPart1.JettonAmount.String()
		tokenIn = stonfiPart2.RefData.Token0.String()
		amountOut = stonfiPart2.RefData.Amount1Out.String()
		tokenOut = stonfiPart2.RefData.Token1.String()

	} else {
		amountIn = stonfiPart1.JettonAmount.String()
		tokenIn = stonfiPart2.RefData.Token1.String()
		amountOut = stonfiPart2.RefData.Amount0Out.String()
		tokenOut = stonfiPart2.RefData.Token0.String()
	}

	logrus.Infof("[STON.FI] new swap found")
	logrus.Infof("[STON.FI] swapper: %s", stonfiPart1.ToAddress)
	logrus.Infof("[STON.FI] amount in: %s %s", amountIn, tokenIn)
	logrus.Infof("[STON.FI] amount out: %s %s\n\n", amountOut, tokenOut)

	ss := &storage.StonfiSwap{
		Address:   stonfiPart1.ToAddress.String(),
		TokenIn:   tokenIn,
		AmountIn:  amountIn,
		TokenOut:  tokenOut,
		AmountOut: amountOut,
	}

	return ss, nil
}
