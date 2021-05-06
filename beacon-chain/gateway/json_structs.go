package gateway

// Beacon

// beacon/genesis
type GenesisResponseJson struct {
	Data *GenesisResponse_GenesisJson `json:"data"`
}
type GenesisResponse_GenesisJson struct {
	GenesisTime           string `json:"genesis_time"`
	GenesisValidatorsRoot string `json:"genesis_validators_root" hex:"true"`
	GenesisForkVersion    string `json:"genesis_fork_version" hex:"true"`
}

// beacon/states/{state_id}/root
type StateRootResponseJson struct {
	Data *StateRootResponse_StateRootJson `json:"data"`
}
type StateRootResponse_StateRootJson struct {
	// TODO: json tag should be "root" - we need to change the name in ethereumapis
	StateRoot string `json:"state_root" hex:"true"`
}

// beacon/states/{state_id}/fork
type StateForkResponseJson struct {
	Data *ForkJson `json:"data"`
}
type ForkJson struct {
	PreviousVersion string `json:"previous_version" hex:"true"`
	CurrentVersion  string `json:"current_version" hex:"true"`
	Epoch           string `json:"epoch"`
}

// beacon/states/{state_id}/finality_checkpoints
type StateFinalityCheckpointResponseJson struct {
	Data *StateFinalityCheckpointResponse_StateFinalityCheckpointJson `json:"data"`
}
type StateFinalityCheckpointResponse_StateFinalityCheckpointJson struct {
	PreviousJustified *CheckpointJson `json:"previous_justified"`
	CurrentJustified  *CheckpointJson `json:"current_justified"`
	Finalized         *CheckpointJson `json:"finalized"`
}

// Reusable types.
type CheckpointJson struct {
	Epoch string `json:"epoch"`
	Root  string `json:"root" hex:"true"`
}

type ErrorJson struct {
	Message string `json:"message"`
}
