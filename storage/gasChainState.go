package storage

type Gas uint64

const (
	STOREBYTES Gas = 20
	READFLAT   Gas = 20
	READBYTES  Gas = 2
	WRITEFLAT  Gas = 200
	WRITEBYTES Gas = 20
	VERIFYSIG  Gas = 50
	HASHBYTES  Gas = 5
	CHECKEXIST Gas = 20
	DELETE     Gas = 50
)

// Calculate the gas used for each action, will be embedded with GasChainState.
type GasCalculator interface {
	// Consume amount of Gas for the Category
	Consume(amount, category Gas, allowOverflow bool) bool

	// Get the max amount of Gas the GasCalculator accept
	GetLimit() Gas

	// Get the current consumed Gas
	GetConsumed() Gas

	// Check if the block has fullfill the Gas Limit
	IsEnough() bool
}

var _ GasCalculator = gasCalculator{}

type gasCalculator struct {
	limit    Gas
	consumed Gas
}

func (g gasCalculator) IsEnough() bool {
	if g.consumed >= g.limit {
		return true
	}
	return false
}

func (g gasCalculator) Consume(amount, category Gas, allowOverflow bool) bool {
	currentGasCost := amount * category
	if allowOverflow {
		g.consumed += currentGasCost
		return true
	} else {
		if g.consumed >= g.limit {
			return false
		} else {
			g.consumed += currentGasCost
			return true
		}
	}
}

func (g gasCalculator) GetLimit() Gas {
	return g.limit
}

func (g gasCalculator) GetConsumed() Gas {
	return g.consumed
}

func NewGasCalculator(limit Gas) GasCalculator {
	return &gasCalculator{
		limit:    limit,
		consumed: 0,
	}
}

type GasChainState struct {
	*ChainState
	GasCalculator
}

func (g *GasChainState) Set(key, value []byte) error {
	ok := g.GasCalculator.Consume(Gas(1), WRITEFLAT, false)
	if !ok {
		return ErrExceedGasLimit
	}
	err := g.ChainState.Set(key, value)
	if err != nil {
		return err
	}
	g.GasCalculator.Consume(Gas(len(value)), WRITEBYTES, true)
	return nil
}

func (g *GasChainState) Get(key StoreKey, lastCommit bool) []byte {
	ok := g.GasCalculator.Consume(Gas(1), READFLAT, false)
	if !ok {
		log.Error(ErrExceedGasLimit.Error())
		return nil
	}
	value := g.ChainState.Get(key, lastCommit)
	ok = g.GasCalculator.Consume(Gas(len(value)), READBYTES, true)
	return value
}

func (g *GasChainState) Exists(key StoreKey) bool {
	ok := g.GasCalculator.Consume(Gas(1), CHECKEXIST, false)
	if !ok {
		log.Error(ErrExceedGasLimit.Error())
		return false
	}
	exist := g.ChainState.Exists(key)
	return exist
}

func (g *GasChainState) Remove(key []byte) ([]byte, bool) {
	ok := g.GasCalculator.Consume(Gas(1), DELETE, false)
	if !ok {
		log.Error(ErrExceedGasLimit.Error())
		return nil, false
	}
	value, ok := g.ChainState.Remove(key)
	return value, ok
}
