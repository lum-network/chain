package types

func NewGenesisState(queries []Query) *GenesisState {
	return &GenesisState{Queries: queries}
}

// DefaultGenesis returns the default Capability genesis state.
func DefaultGenesis() *GenesisState {
	var queries []Query
	return NewGenesisState(queries)
}

// Validate performs basic genesis state validation returning an error upon any failure.
func (gs GenesisState) Validate() error {
	// Loop through all queries and validate them
	for _, query := range gs.GetQueries() {
		if err := query.ValidateBasic(); err != nil {
			return err
		}
	}
	return nil
}
