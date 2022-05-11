package domain

import "context"

type AssetAllocationStrategy struct {
	//strategy string
}

type AssetAllocationStorage interface {
}

var _ Strategy = (*AssetAllocationStrategy)(nil)

func (s AssetAllocationStrategy) Name() string {
	return "Asset Allocation"
}

func (s AssetAllocationStrategy) Parameters() []StrategyParameter {
	return []StrategyParameter{}
}

func (s AssetAllocationStrategy) ValidateParameter(param StrategyParameter) bool {
	return true
}

func (s *AssetAllocationStrategy) Run(ctx context.Context, storage interface{}, exchanges []Exchange, logger Logger) error {
	// check balance
	balance, err := exchanges[0].Balances(nil)
	if err != nil {
		return err
	}

	// calc change allocation
	balance[0] = balance[1]
	// check can we buy some asset from index
	// buy asset
	return nil
}
