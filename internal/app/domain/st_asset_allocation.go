package domain

type AssetAllocationStrategy struct {
	//strategy string
}

type AssetAllocationStorage interface {
}

func (s *AssetAllocationStrategy) Run(storage AssetAllocationStorage, exchange Exchange, logger Logger) error {
	// check balance
	balance, err := exchange.Balances(nil)
	if err != nil {
		return err
	}

	// calc change allocation
	balance[0] = balance[1]
	// check can we buy some asset from index
	// buy asset
	return nil
}
