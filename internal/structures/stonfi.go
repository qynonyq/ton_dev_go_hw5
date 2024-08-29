package structures

import (
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
)

type (
	// input message
	StonfiSwap struct {
		_             tlb.Magic        `tlb:"#25938561"`
		QueryID       uint64           `tlb:"## 64"`
		ToAddress     *address.Address `tlb:"addr"`
		SenderAddress *address.Address `tlb:"addr"`
		JettonAmount  tlb.Coins        `tlb:"."`
		MinOut        tlb.Coins        `tlb:"."`
		HasRefAddress bool             `tlb:"bool"`
		RefAddress    *address.Address `tlb:"?HasRefAddress addr"`
	}

	// output message
	StonfiPaymentRequest struct {
		_            tlb.Magic        `tlb:"#f93bb43f"`
		QueryID      uint64           `tlb:"## 64"`
		OwnerAddress *address.Address `tlb:"addr"`
		ExitCode     uint32           `tlb:"## 32"`
		RefData      struct {
			Amount0Out tlb.Coins        `tlb:"."`
			Token0     *address.Address `tlb:"addr"`
			Amount1Out tlb.Coins        `tlb:"."`
			Token1     *address.Address `tlb:"addr"`
		} `tlb:"^"`
	}
)
