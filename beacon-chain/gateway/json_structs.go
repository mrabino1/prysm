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
	StateRoot string `json:"state_root" hex:"true"` // TODO: json tag should be "root"
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

// beacon/headers/{block_id}
type BlockHeaderResponseJson struct {
	Data *BlockHeaderContainerJson `json:"data"`
}

// beacon/blocks/{block_id}
type BlockResponseJson struct {
	Data *BeaconBlockContainerJson `json:"data"`
}

// beacon/blocks/{block_id}/root
type BlockRootResponseJson struct {
	Data *BlockRootContainerJson `json:"data"`
}

// beacon/blocks/{block_id}/attestations
type BlockAttestationsResponseJson struct {
	Data []*AttestationJson `json:"data"`
}

// beacon/pool/attester_slashings
type AttesterSlashingsPoolResponseJson struct {
	Data []*AttesterSlashingJson `json:"data"`
}

// Reusable types.
type CheckpointJson struct {
	Epoch string `json:"epoch"`
	Root  string `json:"root" hex:"true"`
}
type BlockRootContainerJson struct {
	Root string `json:"root" hex:"true"`
}
type BeaconBlockContainerJson struct {
	Message   *BeaconBlockJson `json:"message"`
	Signature string           `json:"signature" hex:"true"`
}
type BeaconBlockJson struct {
	Slot          string               `json:"slot"`
	ProposerIndex string               `json:"proposer_index"`
	ParentRoot    string               `json:"parent_root" hex:"true"`
	StateRoot     string               `json:"state_root" hex:"true"`
	Body          *BeaconBlockBodyJson `json:"body"`
}
type BeaconBlockBodyJson struct {
	RandaoReveal      string                     `json:"randao_reveal" hex:"true"`
	Eth1Data          *Eth1DataJson              `json:"eth1_data"`
	Graffiti          string                     `json:"graffiti"`
	ProposerSlashings []*ProposerSlashingJson    `json:"proposer_slashings"`
	AttesterSlashings []*AttesterSlashingJson    `json:"attester_slashings"`
	Attestations      []*AttestationJson         `json:"attestations"`
	Deposits          []*DepositJson             `json:"deposits"`
	VoluntaryExits    []*SignedVoluntaryExitJson `json:"voluntary_exits"`
}
type BlockHeaderContainerJson struct {
	Root      string                          `json:"root" hex:"true"`
	Canonical bool                            `json:"canonical"`
	Header    *BeaconBlockHeaderContainerJson `json:"header"`
}
type BeaconBlockHeaderContainerJson struct {
	Message   *BeaconBlockHeaderJson `json:"message"`
	Signature string                 `json:"signature" hex:"true"`
}
type SignedBeaconBlockHeaderJson struct {
	Header    *BeaconBlockHeaderJson `json:"header"` // TODO: json tag should be "message"
	Signature string                 `json:"signature" hex:"true"`
}
type BeaconBlockHeaderJson struct {
	Slot          string `json:"slot"`
	ProposerIndex string `json:"proposer_index"`
	ParentRoot    string `json:"parent_root" hex:"true"`
	StateRoot     string `json:"state_root" hex:"true"`
	BodyRoot      string `json:"body_root" hex:"true"`
}
type Eth1DataJson struct {
	DepositRoot  string `json:"deposit_root" hex:"true"`
	DepositCount string `json:"deposit_count"`
	BlockHash    string `json:"block_hash" hex:"true"`
}
type ProposerSlashingJson struct {
	Header_1 *SignedBeaconBlockHeaderJson `json:"header_1"` // TODO: json tag should be "signed_header_1"
	Header_2 *SignedBeaconBlockHeaderJson `json:"header_2"` // TODO: json tag should be "signed_header_2"
}
type AttesterSlashingJson struct {
	Attestation_1 *IndexedAttestationJson `json:"attestation_1"`
	Attestation_2 *IndexedAttestationJson `json:"attestation_2"`
}
type IndexedAttestationJson struct {
	AttestingIndices []string             `json:"attesting_indices"`
	Data             *AttestationDataJson `json:"data"`
	Signature        string               `json:"signature" hex:"true"`
}
type AttestationJson struct {
	AggregationBits string               `json:"aggregation_bits" hex:"true"`
	Data            *AttestationDataJson `json:"data"`
	Signature       string               `json:"signature" hex:"true"`
}
type AttestationDataJson struct {
	Slot            string          `json:"slot"`
	CommitteeIndex  string          `json:"committee_index"` // TODO: json tag should be "index"
	BeaconBlockRoot string          `json:"beacon_block_root" hex:"true"`
	Source          *CheckpointJson `json:"source"`
	Target          *CheckpointJson `json:"target"`
}
type DepositJson struct {
	// TODO: How to transcode?
	Proof []string          `json:"proof" hex:"true"`
	Data  *Deposit_DataJson `json:"data"`
}
type Deposit_DataJson struct {
	PublicKey             string `json:"public_key" hex:"true"` // TODO: json tag should be "pubkey"
	WithdrawalCredentials string `json:"withdrawal_credentials" hex:"true"`
	Amount                string `json:"amount"`
	Signature             string `json:"signature" hex:"true"`
}
type SignedVoluntaryExitJson struct {
	Exit      *VoluntaryExitJson `json:"exit"` // TODO: json tag should be "message"
	Signature string             `json:"signature" hex:"true"`
}
type VoluntaryExitJson struct {
	Epoch          string `json:"epoch"`
	ValidatorIndex string `json:"validator_index"`
}

// Error handling.
type ErrorJson struct {
	Message string `json:"message"`
}
