package storage

type StonfiSwap struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement:true;"`
	Address   string
	TokenIn   string
	AmountIn  string
	TokenOut  string
	AmountOut string
}
