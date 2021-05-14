package gateway

import "github.com/prysmaticlabs/prysm/beacon-chain/rpc/beaconv1"

// Beacon

// beacon/genesis
type GenesisResponseJson struct {
	Data *GenesisResponse_GenesisJson `json:"data"`
}
type GenesisResponse_GenesisJson struct {
	GenesisTime           string `json:"genesis_time" time:"true"`
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

// beacon/blocks/{block_id}/attestations (GET)
type BlockAttestationsResponseJson struct {
	Data []*AttestationJson `json:"data"`
}

// beacon/blocks/{block_id}/attestations (POST)
type SubmitAttestationRequestJson struct {
	Data []*AttestationJson `json:"data"`
}

// beacon/pool/attester_slashings
type AttesterSlashingsPoolResponseJson struct {
	Data []*AttesterSlashingJson `json:"data"`
}

// beacon/pool/proposer_slashings
type ProposerSlashingsPoolResponseJson struct {
	Data []*ProposerSlashingJson `json:"data"`
}

// beacon/pool/voluntary_exits
type VoluntaryExitsPoolResponseJson struct {
	Data []*SignedVoluntaryExitJson `json:"data"`
}

// node/identity
type IdentityResponseJson struct {
	Data *IdentityJson `json:"data"`
}

// node/peers
type PeersResponseJson struct {
	Data []*PeerJson `json:"data"`
}

// node/peers/{peer_id}
type PeerResponseJson struct {
	Data *PeerJson `json:"data"`
}

// node/peer_count
type PeerCountResponseJson struct {
	Data PeerCountResponse_PeerCountJson `json:"data"`
}
type PeerCountResponse_PeerCountJson struct {
	Disconnected  string `json:"disconnected"`
	Connecting    string `json:"connecting"`
	Connected     string `json:"connected"`
	Disconnecting string `json:"disconnecting"`
}

// node/version
type VersionResponseJson struct {
	Data *VersionJson `json:"data"`
}

// debug/beacon/states/{state_id}
type BeaconStateResponseJson struct {
	Data *BeaconStateJson `json:"data"`
}

// debug/beacon/heads
type ForkChoiceHeadsResponseJson struct {
	Data []*ForkChoiceHeadJson `json:"data"`
}

// config/fork_schedule
type ForkScheduleResponseJson struct {
	Data []*ForkJson `json:"data"`
}

// config/deposit_contract
type DepositContractResponseJson struct {
	Data *DepositContractJson `json:"data"`
}

// config/spec
type SpecResponseJson struct {
	Data interface{} `json:"data"`
}

//----------------
// Reusable types.
//----------------

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
	Graffiti          string                     `json:"graffiti" hex:"true"`
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
type IdentityJson struct {
	PeerId             string        `json:"peer_id"`
	Enr                string        `json:"enr"`
	P2PAddresses       []string      `json:"p2p_addresses"`
	DiscoveryAddresses []string      `json:"discovery_addresses"`
	Metadata           *MetadataJson `json:"metadata"`
}
type MetadataJson struct {
	SeqNumber string `json:"seq_number"`
	Attnets   string `json:"attnets" hex:"true"`
}
type PeerJson struct {
	PeerId    string `json:"peer_id"`
	Enr       string `json:"enr"`
	Address   string `json:"address"` // TODO: json tag should be "last_seen_p2p_address"
	State     string `json:"state" enum:"true"`
	Direction string `json:"direction" enum:"true"`
}
type VersionJson struct {
	Version string `json:"version"`
}
type BeaconStateJson struct {
	GenesisTime                 string                    `json:"genesis_time" time:"true"`
	GenesisValidatorsRoot       string                    `json:"genesis_validators_root" hex:"true"`
	Slot                        string                    `json:"slot"`
	Fork                        *ForkJson                 `json:"fork"`
	LatestBlockHeader           *BeaconBlockHeaderJson    `json:"latest_block_header"`
	BlockRoots                  []string                  `json:"block_roots" hex:"true"`
	StateRoots                  []string                  `json:"state_roots" hex:"true"`
	HistoricalRoots             []string                  `json:"historical_roots" hex:"true"`
	Eth1Data                    *Eth1DataJson             `json:"eth1_data"`
	Eth1DataVotes               []*Eth1DataJson           `json:"eth1_data_votes"`
	Eth1DepositIndex            string                    `json:"eth1_deposit_index"`
	Validators                  []*ValidatorJson          `json:"validators"`
	Balances                    []string                  `json:"balances"`
	RandaoMixes                 []string                  `json:"randao_mixes" hex:"true"`
	Slashings                   []string                  `json:"slashings"`
	PreviousEpochAttestations   *[]PendingAttestationJson `json:"previous_epoch_attestations"`
	CurrentEpochAttestations    *[]PendingAttestationJson `json:"current_epoch_attestations"`
	JustificationBits           string                    `json:"justification_bits" hex:"true"`
	PreviousJustifiedCheckpoint *CheckpointJson           `json:"previous_justified_checkpoint"`
	CurrentJustifiedCheckpoint  *CheckpointJson           `json:"current_justified_checkpoint"`
	FinalizedCheckpoint         *CheckpointJson           `json:"finalized_checkpoint"`
}
type ForkJson struct {
	PreviousVersion string `json:"previous_version" hex:"true"`
	CurrentVersion  string `json:"current_version" hex:"true"`
	Epoch           string `json:"epoch"`
}
type ValidatorJson struct {
	PublicKey                  string `json:"public_key" hex:"true"`
	WithdrawalCredentials      string `json:"withdrawal_credentials" hex:"true"`
	EffectiveBalance           string `json:"effective_balance"`
	Slashed                    bool   `json:"slashed"`
	ActivationEligibilityEpoch string `json:"activation_eligibility_epoch"`
	ActivationEpoch            string `json:"activation_epoch"`
	ExitEpoch                  string `json:"exit_epoch"`
	WithdrawableEpoch          string `json:"withdrawable_epoch"`
}
type PendingAttestationJson struct {
	AggregationBits string               `json:"aggregation_bits" hex:"true"`
	Data            *AttestationDataJson `json:"data"`
	InclusionDelay  string               `json:"inclusion_delay"`
	ProposerIndex   string               `json:"proposer_index"`
}
type ForkChoiceHeadJson struct {
	Root string `json:"root" hex:"true"`
	Slot string `json:"slot"`
}
type DepositContractJson struct {
	ChainId string `json:"chain_id"`
	Address string `json:"address"`
}

// ---------------
// Error handling.
// ---------------

type ErrorJson interface {
	StatusCode() int
	SetCode(code int)
	Msg() string
}
type DefaultErrorJson struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}
type SubmitAttestationsErrorJson struct {
	DefaultErrorJson
	Failures []*beaconv1.SingleAttestationVerificationFailure `json:"failures"`
}

func (e *DefaultErrorJson) StatusCode() int {
	return e.Code
}

func (e *DefaultErrorJson) Msg() string {
	return e.Message
}

func (e *DefaultErrorJson) SetCode(code int) {
	e.Code = code
}
