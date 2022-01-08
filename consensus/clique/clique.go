// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// Package clique implements the proof-of-authority consensus engine.
package clique

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/misc"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/trie"
	lru "github.com/hashicorp/golang-lru"
	"golang.org/x/crypto/sha3"
)

const (
	checkpointInterval = 1024 // Number of blocks after which to save the vote snapshot to the database
	inmemorySnapshots  = 128  // Number of recent vote snapshots to keep in memory
	inmemorySignatures = 4096 // Number of recent block signatures to keep in memory
)

// Clique proof-of-authority protocol constants.
var (
	epochLength = uint64(30000) // Default number of blocks after which to checkpoint and reset the pending votes

	extraVanity = 32                     // Fixed number of extra-data prefix bytes reserved for signer vanity
	extraSeal   = crypto.SignatureLength // Fixed number of extra-data suffix bytes reserved for signer seal

	nonceAuthVote = hexutil.MustDecode("0xffffffffffffffff") // Magic nonce number to vote on adding a new signer
	nonceDropVote = hexutil.MustDecode("0x0000000000000000") // Magic nonce number to vote on removing a signer.

	uncleHash = types.CalcUncleHash(nil) // Always Keccak256(RLP([])) as uncles are meaningless outside of PoW.
)

// Various error messages to mark blocks invalid. These should be private to
// prevent engine specific errors from being referenced in the remainder of the
// codebase, inherently breaking if the engine is swapped out. Please put common
// error types into the consensus package.
var (
	// errUnknownBlock is returned when the list of signers is requested for a block
	// that is not part of the local blockchain.
	errUnknownBlock = errors.New("unknown block")

	// errInvalidVote is returned if a nonce value is something else that the two
	// allowed constants of 0x00..0 or 0xff..f.
	errInvalidVote = errors.New("vote nonce not 0x00..0 or 0xff..f")

	// errMissingVanity is returned if a block's extra-data section is shorter than
	// 32 bytes, which is required to store the signer vanity.
	errMissingVanity = errors.New("extra-data 32 byte vanity prefix missing")

	// errMissingSignature is returned if a block's extra-data section doesn't seem
	// to contain a 65 byte secp256k1 signature.
	errMissingSignature = errors.New("extra-data 65 byte signature suffix missing")

	// errExtraSigners is returned if non-checkpoint block contain signer data in
	// their extra-data fields.
	errExtraSigners = errors.New("non-checkpoint block contains extra signer list")

	// errInvalidMixDigest is returned if a block's mix digest is non-zero.
	errInvalidMixDigest = errors.New("non-zero mix digest")

	// errInvalidUncleHash is returned if a block contains an non-empty uncle list.
	errInvalidUncleHash = errors.New("non empty uncle hash")

	// errInvalidTimestamp is returned if the timestamp of a block is lower than
	// the previous block's timestamp + the minimum block period.
	errInvalidTimestamp = errors.New("invalid timestamp")

	// errInvalidVotingChain is returned if an authorization list is attempted to
	// be modified via out-of-range or non-contiguous headers.
	errInvalidVotingChain = errors.New("invalid voting chain")

	// errUnauthorizedSigner is returned if a header is signed by a non-authorized entity.
	errUnauthorizedSigner = errors.New("unauthorized signer")

	// errRecentlySigned is returned if a header is signed by an authorized entity
	// that already signed a header recently, thus is temporarily not allowed to.
	errRecentlySigned = errors.New("recently signed")

	ErrInvalidBlockWitness   = errors.New("invalid block witness")
	ErrMinerFutureBlock      = errors.New("miner the future block")
	ErrWaitForPrevBlock      = errors.New("wait for last block arrived")
	ErrWaitForRightTime      = errors.New("wait for right time")
	ErrInvalidMinerBlockTime = errors.New("invalid time to miner the block")
)

// SignerFn hashes and signs the data to be signed by a backing account.
type SignerFn func(signer common.Address, mimeType string, message []byte) ([]byte, error)

type MasternodeListFn func(number *big.Int) ([]common.Address, error)
type InvestorFn func(investor common.Address, number *big.Int) (common.Address, error)

// ecrecover extracts the Ethereum account address from a signed header.
func ecrecover(header *types.Header, sigcache *lru.ARCCache) (common.Address, error) {
	// If the signature's already cached, return that
	hash := header.Hash()
	if address, known := sigcache.Get(hash); known {
		return address.(common.Address), nil
	}
	// Retrieve the signature from the header extra-data
	if len(header.Extra) < extraSeal {
		return common.Address{}, errMissingSignature
	}
	signature := header.Extra[len(header.Extra)-extraSeal:]

	// Recover the public key and the Ethereum address
	pubkey, err := crypto.Ecrecover(SealHash(header).Bytes(), signature)
	if err != nil {
		return common.Address{}, err
	}
	var signer common.Address
	copy(signer[:], crypto.Keccak256(pubkey[1:])[12:])
	//sigcache.Add(hash, signer)
	return signer, nil
}

// Clique is the proof-of-authority consensus engine proposed to support the
// Ethereum testnet following the Ropsten attacks.
type Clique struct {
	config *params.CliqueConfig // Consensus engine configuration parameters
	db     ethdb.Database       // Database to store and retrieve snapshot checkpoints

	recents    *lru.ARCCache // Snapshots for recent block to speed up reorgs
	signatures *lru.ARCCache // Signatures of recent blocks to speed up mining

	proposals map[common.Address]bool // Current list of proposals we are pushing

	// signer common.Address // Ethereum address of the signing key
	signer common.Address
	signFn SignerFn     // Signer function to authorize hashes with
	lock   sync.RWMutex // Protects the signer fields

	// The fields below are for testing only
	fakeDiff bool // Skip difficulty verifications

	masternodeListFn MasternodeListFn //get current all masternodes
	investorFn       InvestorFn       //get current all masternodes

	cacheNumber uint64
	cacheNodes  []common.Address
	cacheHash   common.Hash
	witnesses   []common.Address
}

// New creates a Clique proof-of-authority consensus engine with the initial
// signers set to the ones provided by the user.
func New(config *params.CliqueConfig, db ethdb.Database) *Clique {
	// Set any missing consensus parameters to their defaults
	conf := *config
	if conf.Epoch == 0 {
		conf.Epoch = epochLength
	}
	// Allocate the snapshot caches and create the engine
	recents, _ := lru.NewARC(inmemorySnapshots)
	signatures, _ := lru.NewARC(inmemorySignatures)

	return &Clique{
		config:      &conf,
		db:          db,
		recents:     recents,
		signatures:  signatures,
		proposals:   make(map[common.Address]bool),
		cacheNumber: 0,
	}
}

// Author implements consensus.Engine, returning the Ethereum address recovered
// from the signature in the header's extra-data section.
func (c *Clique) Author(header *types.Header) (common.Address, error) {
	return ecrecover(header, c.signatures)
}

// VerifyHeader checks whether a header conforms to the consensus rules.
func (c *Clique) VerifyHeader(chain consensus.ChainHeaderReader, header *types.Header, seal bool) error {
	return c.verifyHeader(chain, header, nil)
}

// VerifyHeaders is similar to VerifyHeader, but verifies a batch of headers. The
// method returns a quit channel to abort the operations and a results channel to
// retrieve the async verifications (the order is that of the input slice).
func (c *Clique) VerifyHeaders(chain consensus.ChainHeaderReader, headers []*types.Header, seals []bool) (chan<- struct{}, <-chan error) {
	abort := make(chan struct{})
	results := make(chan error, len(headers))

	go func() {
		for i, header := range headers {
			err := c.verifyHeader(chain, header, headers[:i])

			select {
			case <-abort:
				return
			case results <- err:
			}
		}
	}()
	return abort, results
}

// verifyHeader checks whether a header conforms to the consensus rules.The
// caller may optionally pass in a batch of parents (ascending order) to avoid
// looking those up from the database. This is useful for concurrently verifying
// a batch of new headers.
func (c *Clique) verifyHeader(chain consensus.ChainHeaderReader, header *types.Header, parents []*types.Header) error {
	if header.Number == nil {
		return errUnknownBlock
	}
	// number := header.Number.Uint64()

	// Don't waste time checking blocks from the future
	if header.Time > uint64(time.Now().Unix()) {
		return consensus.ErrFutureBlock
	}
	// Checkpoint blocks need to enforce zero beneficiary
	//checkpoint := (number % c.config.Epoch) == 0
	//if checkpoint && header.Coinbase != (common.Address{}) {
	//	return errInvalidCheckpointBeneficiary
	//}
	// Nonces must be 0x00..0 or 0xff..f, zeroes enforced on checkpoints
	//if !bytes.Equal(header.Nonce[:], nonceAuthVote) && !bytes.Equal(header.Nonce[:], nonceDropVote) {
	//	return errInvalidVote
	//}
	//if checkpoint && !bytes.Equal(header.Nonce[:], nonceDropVote) {
	//	return errInvalidCheckpointVote
	//}
	// Check that the extra-data contains both the vanity and signature
	if len(header.Extra) < extraVanity {
		return errMissingVanity
	}
	if len(header.Extra) < extraVanity+extraSeal {
		return errMissingSignature
	}
	// Ensure that the extra-data contains a signer list on checkpoint, but none otherwise
	signersBytes := len(header.Extra) - extraVanity - extraSeal
	// if !checkpoint && signersBytes != 0 {
	if signersBytes != 0 {
		return errExtraSigners
	}
	//if checkpoint && signersBytes%common.AddressLength != 0 {
	//	return errInvalidCheckpointSigners
	//}
	// Ensure that the mix digest is zero as we don't have fork protection currently
	if header.MixDigest != (common.Hash{}) {
		return errInvalidMixDigest
	}
	// Ensure that the block doesn't contain any uncles which are meaningless in PoA
	if header.UncleHash != uncleHash {
		return errInvalidUncleHash
	}
	// Ensure that the block's difficulty is meaningful (may not be correct at this point)
	//if number > 0 {
	//	if header.Difficulty == nil || (header.Difficulty.Cmp(diffInTurn) != 0 && header.Difficulty.Cmp(diffNoTurn) != 0) {
	//		return errInvalidDifficulty
	//	}
	//}
	// Verify that the gas limit is <= 2^63-1
	cap := uint64(0x7fffffffffffffff)
	if header.GasLimit > cap {
		return fmt.Errorf("invalid gasLimit: have %v, max %v", header.GasLimit, cap)
	}
	// If all checks passed, validate any special fields for hard forks
	//if err := misc.VerifyForkHashes(chain.Config(), header, false); err != nil {
	//	return err
	//}
	// All basic checks passed, verify cascading fields
	return c.verifyCascadingFields(chain, header, parents)
}

// verifyCascadingFields verifies all the header fields that are not standalone,
// rather depend on a batch of previous headers. The caller may optionally pass
// in a batch of parents (ascending order) to avoid looking those up from the
// database. This is useful for concurrently verifying a batch of new headers.
func (c *Clique) verifyCascadingFields(chain consensus.ChainHeaderReader, header *types.Header, parents []*types.Header) error {
	// The genesis block is the always valid dead-end
	number := header.Number.Uint64()
	if number == 0 {
		return nil
	}
	// Ensure that the block's timestamp isn't too close to its parent
	var parent *types.Header
	if len(parents) > 0 {
		parent = parents[len(parents)-1]
	} else {
		parent = chain.GetHeader(header.ParentHash, number-1)
	}
	if parent == nil || parent.Number.Uint64() != number-1 || parent.Hash() != header.ParentHash {
		return consensus.ErrUnknownAncestor
	}
	if parent.Time+c.config.Period > header.Time {
		return errInvalidTimestamp
	}
	// Verify that the gasUsed is <= gasLimit
	if header.GasUsed > header.GasLimit {
		return fmt.Errorf("invalid gasUsed: have %d, gasLimit %d", header.GasUsed, header.GasLimit)
	}
	if !chain.Config().IsLondon(header.Number) {
		// Verify BaseFee not present before EIP-1559 fork.
		if header.BaseFee != nil {
			return fmt.Errorf("invalid baseFee before fork: have %d, want <nil>", header.BaseFee)
		}
		if err := misc.VerifyGaslimit(parent.GasLimit, header.GasLimit); err != nil {
			return err
		}
	} else if err := misc.VerifyEip1559Header(chain.Config(), parent, header); err != nil {
		// Verify the header's EIP-1559 attributes.
		return err
	}
	// Retrieve the snapshot needed to verify this header and cache it
	//snap, err := c.snapshot(chain, number-1, header.ParentHash, parents)
	//if err != nil {
	//	return err
	//}
	// If the block is a checkpoint block, verify the signer list
	//if number%c.config.Epoch == 0 {
	//	signers := make([]byte, len(snap.Signers)*common.AddressLength)
	//	for i, signer := range snap.signers() {
	//		copy(signers[i*common.AddressLength:], signer[:])
	//	}
	//	extraSuffix := len(header.Extra) - extraSeal
	//	if !bytes.Equal(header.Extra[extraVanity:extraSuffix], signers) {
	//		return errMismatchingCheckpointSigners
	//	}
	//}
	// All basic checks passed, verify the seal and return
	return c.verifySeal(chain, header, parents)
}

// snapshot retrieves the authorization snapshot at a given point in time.
func (c *Clique) snapshot(chain consensus.ChainHeaderReader, number uint64, hash common.Hash, parents []*types.Header) (*Snapshot, error) {
	// Search for a snapshot in memory or on disk for checkpoints
	var (
		headers []*types.Header
		snap    *Snapshot
	)
	for snap == nil {
		// If an in-memory snapshot was found, use that
		if s, ok := c.recents.Get(hash); ok {
			snap = s.(*Snapshot)
			break
		}
		// If an on-disk checkpoint snapshot can be found, use that
		if number%checkpointInterval == 0 {
			if s, err := loadSnapshot(c.config, c.signatures, c.db, hash); err == nil {
				log.Trace("Loaded voting snapshot from disk", "number", number, "hash", hash)
				snap = s
				break
			}
		}
		// If we're at the genesis, snapshot the initial state. Alternatively if we're
		// at a checkpoint block without a parent (light client CHT), or we have piled
		// up more headers than allowed to be reorged (chain reinit from a freezer),
		// consider the checkpoint trusted and snapshot it.
		if number == 0 || (number%c.config.Epoch == 0 && (len(headers) > params.FullImmutabilityThreshold || chain.GetHeaderByNumber(number-1) == nil)) {
			checkpoint := chain.GetHeaderByNumber(number)
			if checkpoint != nil {
				hash := checkpoint.Hash()

				signers := make([]common.Address, (len(checkpoint.Extra)-extraVanity-extraSeal)/common.AddressLength)
				for i := 0; i < len(signers); i++ {
					copy(signers[i][:], checkpoint.Extra[extraVanity+i*common.AddressLength:])
				}
				snap = newSnapshot(c.config, c.signatures, number, hash, signers)
				if err := snap.store(c.db); err != nil {
					return nil, err
				}
				log.Info("Stored checkpoint snapshot to disk", "number", number, "hash", hash)
				break
			}
		}
		// No snapshot for this header, gather the header and move backward
		var header *types.Header
		if len(parents) > 0 {
			// If we have explicit parents, pick from there (enforced)
			header = parents[len(parents)-1]
			if header.Hash() != hash || header.Number.Uint64() != number {
				return nil, consensus.ErrUnknownAncestor
			}
			parents = parents[:len(parents)-1]
		} else {
			// No explicit parents (or no more left), reach out to the database
			header = chain.GetHeader(hash, number)
			if header == nil {
				return nil, consensus.ErrUnknownAncestor
			}
		}
		headers = append(headers, header)
		number, hash = number-1, header.ParentHash
	}
	// Previous snapshot found, apply any pending headers on top of it
	for i := 0; i < len(headers)/2; i++ {
		headers[i], headers[len(headers)-1-i] = headers[len(headers)-1-i], headers[i]
	}
	snap, err := snap.apply(headers)
	if err != nil {
		return nil, err
	}
	c.recents.Add(snap.Hash, snap)

	// If we've generated a new checkpoint snapshot, save to disk
	if snap.Number%checkpointInterval == 0 && len(headers) > 0 {
		if err = snap.store(c.db); err != nil {
			return nil, err
		}
		log.Trace("Stored voting snapshot to disk", "number", snap.Number, "hash", snap.Hash)
	}
	return snap, err
}

// VerifyUncles implements consensus.Engine, always returning an error for any
// uncles as this consensus mechanism doesn't permit uncles.
func (c *Clique) VerifyUncles(chain consensus.ChainReader, block *types.Block) error {
	if len(block.Uncles()) > 0 {
		return errors.New("uncles not allowed")
	}
	return nil
}

// verifySeal checks whether the signature contained in the header satisfies the
// consensus protocol requirements. The method accepts an optional list of parent
// headers that aren't yet part of the local blockchain to generate the snapshots
// from.
func (c *Clique) verifySeal(chain consensus.ChainHeaderReader, header *types.Header, parents []*types.Header) error {
	// Verifying the genesis block is not supported
	number := header.Number.Uint64()
	if number == 0 {
		return errUnknownBlock
	}
	// Resolve the authorization key and check against signers
	signer, err := ecrecover(header, c.signatures)
	if err != nil {
		return err
	}
	if signer != header.Coinbase {
		return errors.New("verifySeal ERROR: c.signer != header.Coinbase")
	}

	var parent *types.Header
	if len(parents) > 0 {
		parent = parents[len(parents)-1]
	} else {
		parent = chain.GetHeader(header.ParentHash, number-1)
	}
	witness, witnessNext, witnessesHash, misc, err := c.lookup(header.Time, parent)
	if err != nil {
		// log.Warn("Verify Seal", "info", err)
	} else {
		if witness != signer {
			return fmt.Errorf("Invalid block witness signer: %s, witness: %s\n", signer.String(), witness.String())
		}
		var nonce types.BlockNonce
		copy(nonce[0:4], witnessNext.Bytes()[0:4])
		copy(nonce[4:8], witnessesHash.Bytes()[0:4])
		if !bytes.Equal(nonce[:], header.Nonce[:]) {
			return fmt.Errorf("Invalid block nonce, expect: %x, current: %x\n", nonce[:], header.Nonce[:])
		}
		if header.Difficulty.Cmp(misc) != 0 {
			return fmt.Errorf("Invalid block difficulty, expect: %s, current: %s\n", misc.String(), header.Difficulty.String())
		}
	}

	//if _, ok := snap.Signers[signer]; !ok {
	//	return errUnauthorizedSigner
	//}
	//for seen, recent := range snap.Recents {
	//	if recent == signer {
	//		// Signer is among recents, only fail if the current block doesn't shift it out
	//		if limit := uint64(len(snap.Signers)/2 + 1); seen > number-limit {
	//			return errRecentlySigned
	//		}
	//	}
	//}
	//// Ensure that the difficulty corresponds to the turn-ness of the signer
	//if !c.fakeDiff {
	//	inturn := snap.inturn(header.Number.Uint64(), signer)
	//	if inturn && header.Difficulty.Cmp(diffInTurn) != 0 {
	//		return errWrongDifficulty
	//	}
	//	if !inturn && header.Difficulty.Cmp(diffNoTurn) != 0 {
	//		return errWrongDifficulty
	//	}
	//}
	return nil
}

// Prepare implements consensus.Engine, preparing all the consensus fields of the
// header for running the transactions on top.
func (c *Clique) Prepare(chain consensus.ChainHeaderReader, header *types.Header) error {
	// header.Nonce = types.BlockNonce{}
	number := header.Number.Uint64()
	if len(header.Extra) < extraVanity {
		header.Extra = append(header.Extra, bytes.Repeat([]byte{0x00}, extraVanity-len(header.Extra))...)
	}
	header.Extra = header.Extra[:extraVanity]
	header.Extra = append(header.Extra, make([]byte, extraSeal)...)
	parent := chain.GetHeader(header.ParentHash, number-1)
	if parent == nil {
		return consensus.ErrUnknownAncestor
	}
	if header.Difficulty == nil {
		header.Difficulty = common.Big0
	}
	//header.Difficulty = c.CalcDifficulty(chain, header.Time, parent)
	//header.Witness = c.signer
	return nil

	//// If the block isn't a checkpoint, cast a random vote (good enough for now)
	//header.Coinbase = common.Address{}
	//header.Nonce = types.BlockNonce{}
	//
	//number := header.Number.Uint64()
	//// Assemble the voting snapshot to check which votes make sense
	//snap, err := c.snapshot(chain, number-1, header.ParentHash, nil)
	//if err != nil {
	//	return err
	//}
	//if number%c.config.Epoch != 0 {
	//	c.lock.RLock()
	//
	//	// Gather all the proposals that make sense voting on
	//	addresses := make([]common.Address, 0, len(c.proposals))
	//	for address, authorize := range c.proposals {
	//		if snap.validVote(address, authorize) {
	//			addresses = append(addresses, address)
	//		}
	//	}
	//	// If there's pending proposals, cast a vote on them
	//	if len(addresses) > 0 {
	//		header.Coinbase = addresses[rand.Intn(len(addresses))]
	//		if c.proposals[header.Coinbase] {
	//			copy(header.Nonce[:], nonceAuthVote)
	//		} else {
	//			copy(header.Nonce[:], nonceDropVote)
	//		}
	//	}
	//	c.lock.RUnlock()
	//}
	//// Set the correct difficulty
	////header.Difficulty = calcDifficulty(snap, c.signer)
	//header.Difficulty = common.Big1
	//// Ensure the extra data has all its components
	//if len(header.Extra) < extraVanity {
	//	header.Extra = append(header.Extra, bytes.Repeat([]byte{0x00}, extraVanity-len(header.Extra))...)
	//}
	//header.Extra = header.Extra[:extraVanity]
	//
	//if number%c.config.Epoch == 0 {
	//	for _, signer := range snap.signers() {
	//		header.Extra = append(header.Extra, signer[:]...)
	//	}
	//}
	//header.Extra = append(header.Extra, make([]byte, extraSeal)...)
	//
	//// Mix digest is reserved for now, set to empty
	//header.MixDigest = common.Hash{}
	//
	//// Ensure the timestamp has the correct delay
	//parent := chain.GetHeader(header.ParentHash, number-1)
	//if parent == nil {
	//	return consensus.ErrUnknownAncestor
	//}
	//header.Time = parent.Time + c.config.Period
	//if header.Time < uint64(time.Now().Unix()) {
	//	header.Time = uint64(time.Now().Unix())
	//}
	//return nil
}

// Finalize implements consensus.Engine, ensuring no uncles are set.
func (c *Clique) Finalize(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs []*types.Transaction, uncles []*types.Header) {
	// Block reward
	reward, _ := new(big.Int).SetString("8000000000000000000", 10)
	state.AddBalance(params.MasternodeContractAddress, reward)
	balanceMintKey := getNodeAttrKey(header.Coinbase.Bytes()[:], 10)
	balancePledgeKey := getNodeAttrKey(header.Coinbase.Bytes()[:], 8)
	balancePledgeVal := state.GetState(params.MasternodeContractAddress, balancePledgeKey)
	balancePledge := new(big.Int).SetBytes(balancePledgeVal.Bytes())
	blockRegisterKey := getNodeAttrKey(header.Coinbase.Bytes()[:], 5)
	blockRegisterVal := state.GetState(params.MasternodeContractAddress, blockRegisterKey)
	if blockRegisterVal == (common.Hash{}) && balancePledge.Cmp(params.MasternodeCost) < 0 {
		// Genesis node
		balancePledge1 := new(big.Int).Add(balancePledge, reward)
		if balancePledge1.Cmp(params.MasternodeCost) < 0 {
			state.SetState(params.MasternodeContractAddress, balancePledgeKey, common.BytesToHash(balancePledge1.Bytes()))
		} else {
			state.SetState(params.MasternodeContractAddress, balancePledgeKey, common.BytesToHash(params.MasternodeCost.Bytes()))
			balanceMint1 := new(big.Int).Sub(balancePledge1, params.MasternodeCost)
			state.SetState(params.MasternodeContractAddress, balanceMintKey, common.BytesToHash(balanceMint1.Bytes()))
		}
	} else {
		balanceMintVal := state.GetState(params.MasternodeContractAddress, balanceMintKey)
		balanceMint := new(big.Int).SetBytes(balanceMintVal.Bytes())
		balanceMint.Add(balanceMint, reward)
		state.SetState(params.MasternodeContractAddress, balanceMintKey, common.BytesToHash(balanceMint.Bytes()))
	}
	// Online check
	parent := chain.GetHeader(header.ParentHash, header.Number.Uint64()-1)
	if (header.Time-parent.Time) > 3 && header.Number.Uint64() > 1 && header.Coinbase != (common.Address{}) {
		log.Warn("More than 3 seconds", "number", header.Number.String())
	}
	if header.Number.Uint64() > 1 &&
		header.Coinbase != (common.Address{}) &&
		bytes.Equal(header.Nonce[4:8], parent.Nonce[4:8]) &&
		!bytes.Equal(header.Coinbase.Bytes()[0:4], parent.Nonce[0:4]) {
		preNid, err := c.getPreNode(parent.Number.Uint64(), header.Coinbase)
		if err != nil {
			log.Error("Failed to get pre node", "number", header.Number.Uint64(), "error", err.Error())
			return
		}
		log.Warn("Offline detected", "expect", common.Bytes2Hex(parent.Nonce[0:4]), "current", common.Bytes2Hex(header.Coinbase[0:4]), "number", header.Number.Uint64())
		blockOnlineKey := getNodeAttrKey(preNid.Bytes()[:], 6)
		blockOnlineVal := state.GetState(params.MasternodeContractAddress, blockOnlineKey)
		if blockOnlineVal != (common.Hash{}) {
			// countOnlineNode -= 1
			countOnlineNodeKey := common.HexToHash("03")
			countOnlineNodeVal := state.GetState(params.MasternodeContractAddress, countOnlineNodeKey)
			countOnlineNode := new(big.Int).SetBytes(countOnlineNodeVal.Bytes())
			countOnlineNode = new(big.Int).Sub(countOnlineNode, big.NewInt(1))
			countOnlineNodeVal = common.BytesToHash(countOnlineNode.Bytes())
			state.SetState(params.MasternodeContractAddress, countOnlineNodeKey, countOnlineNodeVal)
			// nodes[nid].blockOnline = 0
			state.SetState(params.MasternodeContractAddress, blockOnlineKey, common.Hash{})
			log.Warn("Offline setting", "nid", preNid, "online", countOnlineNode.String(), "number", header.Number.Uint64())
			// reset preOnlineNode
			preOnlineNodeKey := getNodeAttrKey(preNid.Bytes()[:], 2)
			preOnlineNodeVal := state.GetState(params.MasternodeContractAddress, preOnlineNodeKey)
			nextOnlineNodeKey := getNodeAttrKey(preNid.Bytes()[:], 3)
			nextOnlineNodeVal := state.GetState(params.MasternodeContractAddress, nextOnlineNodeKey)
			if preOnlineNodeVal != (common.Hash{}) {
				nextOnlineNodeKey1 := getNodeAttrKey(preOnlineNodeVal.Bytes()[12:32], 3)
				// nodes[preOnlineNode].nextOnlineNode = nextOnlineNode
				state.SetState(params.MasternodeContractAddress, nextOnlineNodeKey1, nextOnlineNodeVal)
				state.SetState(params.MasternodeContractAddress, preOnlineNodeKey, common.Hash{})
			}
			// reset nextOnlineNode
			if nextOnlineNodeVal != (common.Hash{}) {
				preOnlineNodeKey1 := getNodeAttrKey(nextOnlineNodeVal.Bytes()[12:32], 2)
				// nodes[nextOnlineNode].preOnlineNode = preOnlineNode
				state.SetState(params.MasternodeContractAddress, preOnlineNodeKey1, preOnlineNodeVal)
				state.SetState(params.MasternodeContractAddress, nextOnlineNodeKey, common.Hash{})
			} else {
				lastOnlineNodeKey := common.HexToHash("01")
				state.SetState(params.MasternodeContractAddress, lastOnlineNodeKey, preOnlineNodeVal)
			}
		}
	}
	header.Root = state.IntermediateRoot(chain.Config().IsEIP158(header.Number))
	header.UncleHash = types.CalcUncleHash(nil)
}

func getNodeAttrKey(nid []byte, index int64) common.Hash {
	var nodeKeyRaw [64]byte
	nodeKeyRaw[63] = 7
	copy(nodeKeyRaw[12:32], nid[:])
	nodeKey := new(big.Int).SetBytes(crypto.Keccak256(nodeKeyRaw[:]))
	return common.BytesToHash(new(big.Int).Add(nodeKey, big.NewInt(index)).Bytes())
}

// FinalizeAndAssemble implements consensus.Engine, ensuring no uncles are set,
// nor block rewards given, and returns the final block.
func (c *Clique) FinalizeAndAssemble(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs []*types.Transaction, uncles []*types.Header, receipts []*types.Receipt) (*types.Block, error) {
	// Finalize block
	c.Finalize(chain, header, state, txs, uncles)

	// Assemble and return the final block for sealing
	return types.NewBlock(header, txs, nil, receipts, trie.NewStackTrie(nil)), nil
}

// Authorize injects a private key into the consensus engine to mint new blocks
// with.
func (c *Clique) Authorize(witnesses []common.Address, signFn SignerFn) {
	c.lock.Lock()
	defer c.lock.Unlock()

	log.Info("Clique Authorize ", "witnesses", len(c.witnesses))

	c.witnesses = witnesses
	c.signFn = signFn
}

// Seal implements consensus.Engine, attempting to create a sealed block using
// the local signing credentials.
func (c *Clique) Seal(chain consensus.ChainHeaderReader, block *types.Block, results chan<- *types.Block, stop <-chan struct{}) error {
	header := block.Header()

	// Sealing the genesis block is not supported
	number := header.Number.Uint64()
	if number == 0 {
		return errUnknownBlock
	}
	// For 0-period chains, refuse to seal empty blocks (no reward but would spin sealing)
	//if c.config.Period == 0 && len(block.Transactions()) == 0 {
	//	return errors.New("sealing paused while waiting for transactions")
	//}
	// Don't hold the signer fields for the entire sealing procedure
	//c.lock.RLock()
	//signer, signFn := c.signer, c.signFn
	//c.lock.RUnlock()

	//fmt.Println("signer: ", c.signer)

	// Bail out if we're unauthorized to sign a block
	//snap, err := c.snapshot(chain, number-1, header.ParentHash, nil)
	//if err != nil {
	//	return err
	//}
	//if _, authorized := snap.Signers[signer]; !authorized {
	//	return errUnauthorizedSigner
	//}
	//// If we're amongst the recent signers, wait for the next block
	//for seen, recent := range snap.Recents {
	//	if recent == signer {
	//		// Signer is among recents, only wait if the current block doesn't shift it out
	//		if limit := uint64(len(snap.Signers)/2 + 1); number < limit || seen > number-limit {
	//			return errors.New("signed recently, must wait for others")
	//		}
	//	}
	//}
	// Sweet, the protocol permits us to sign the block, wait for our time
	//delay := time.Unix(int64(header.Time), 0).Sub(time.Now()) // nolint: gosimple
	//if header.Difficulty.Cmp(diffNoTurn) == 0 {
	//	// It's not our turn explicitly to sign, delay it a bit
	//	wiggle := time.Duration(len(snap.Signers)/2+1) * wiggleTime
	//	delay += time.Duration(rand.Int63n(int64(wiggle)))
	//
	//	log.Trace("Out-of-turn signing requested", "wiggle", common.PrettyDuration(wiggle))
	//}
	// Sign all the things!
	// sighash, err := signFn(accounts.Account{Address: signer}, accounts.MimetypeClique, CliqueRLP(header))
	sighash, err := c.signFn(c.signer, accounts.MimetypeClique, SealHash(header).Bytes())
	if err != nil {
		return err
	}
	copy(header.Extra[len(header.Extra)-extraSeal:], sighash)
	if c.signer != header.Coinbase {
		return errors.New("Seal ERROR: c.signer != header.Coinbase")
	}
	// Wait until sealing is terminated or delay timeout.
	//log.Trace("Waiting for slot to sign and propagate", "delay", common.PrettyDuration(delay))
	//go func() {
	//	select {
	//	case <-stop:
	//		return
	//	case <-time.After(delay):
	//	}
	//
	//	select {
	//	case results <- block.WithSeal(header):
	//	default:
	//		log.Warn("Sealing result is not read by miner", "sealhash", SealHash(header))
	//	}
	//}()
	results <- block.WithSeal(header)
	return nil
}

// CalcDifficulty is the difficulty adjustment algorithm. It returns the difficulty
// that a new block should have:
// * DIFF_NOTURN(2) if BLOCK_NUMBER % SIGNER_COUNT != SIGNER_INDEX
// * DIFF_INTURN(1) if BLOCK_NUMBER % SIGNER_COUNT == SIGNER_INDEX
func (c *Clique) CalcDifficulty(chain consensus.ChainHeaderReader, time uint64, parent *types.Header) *big.Int {
	return big.NewInt(1)

	//snap, err := c.snapshot(chain, parent.Number.Uint64(), parent.Hash(), nil)
	//if err != nil {
	//	return nil
	//}
	//return calcDifficulty(snap, c.signer)
}

func calcDifficulty(snap *Snapshot, signer string) *big.Int {
	return big.NewInt(1)

	//if snap.inturn(snap.Number+1, signer) {
	//	return new(big.Int).Set(diffInTurn)
	//}
	//return new(big.Int).Set(diffNoTurn)
}

// SealHash returns the hash of a block prior to it being sealed.
func (c *Clique) SealHash(header *types.Header) common.Hash {
	return SealHash(header)
}

// Close implements consensus.Engine. It's a noop for clique as there are no background threads.
func (c *Clique) Close() error {
	return nil
}

// APIs implements consensus.Engine, returning the user facing RPC API to allow
// controlling the signer voting.
func (c *Clique) APIs(chain consensus.ChainHeaderReader) []rpc.API {
	return []rpc.API{{
		Namespace: "clique",
		Version:   "1.0",
		Service:   &API{chain: chain, clique: c},
		Public:    false,
	}}
}

func (c *Clique) checkTime(lastBlock *types.Block, now uint64) error {
	quotientsLast := lastBlock.Time() / c.config.Period
	quotients := now / c.config.Period
	remainder := now % c.config.Period
	if lastBlock.Time() >= (quotients*c.config.Period + 2) {
		return ErrMinerFutureBlock
	}
	if (quotients > quotientsLast) && (remainder == 0) {
		return nil
	}
	return ErrWaitForPrevBlock
}

func (c *Clique) lookup(now uint64, lastBlock *types.Header) (common.Address, common.Address, common.Hash, *big.Int, error) {
	quotientsLast := lastBlock.Time / c.config.Period
	quotients := now / c.config.Period
	if quotientsLast >= quotients {
		return common.Address{}, common.Address{}, common.Hash{}, common.Big0, fmt.Errorf("[LOOKUP] Invalid Period")
	}
	if lastBlock.Time > now {
		return common.Address{}, common.Address{}, common.Hash{}, common.Big0, fmt.Errorf("[LOOKUP] Invalid lastBlock.Time")
	}
	err := c.freshCacheNodes(lastBlock.Number.Uint64())
	if err != nil {
		return common.Address{}, common.Address{}, common.Hash{}, common.Big0, err
	}
	length := uint64(len(c.cacheNodes))
	nextNth := quotients % length
	nextNth2 := nextNth + 1
	if nextNth2 == length {
		nextNth2 = 0
	}
	index := big.NewInt(int64(length) * 1000000)
	index.Add(index, big.NewInt(int64(nextNth)))
	return c.cacheNodes[nextNth], c.cacheNodes[nextNth2], c.cacheHash, index, nil
}

func (c *Clique) CheckWitness(lastBlock *types.Block, now int64) (common.Address, common.Address, common.Hash, *big.Int, error) {
	if err := c.checkTime(lastBlock, uint64(now)); err != nil {
		return common.Address{}, common.Address{}, common.Hash{}, big.NewInt(0), err
	}

	witness, witnessNext, witnessesHash, misc, err := c.lookup(uint64(now), lastBlock.Header())
	if err != nil {
		return common.Address{}, common.Address{}, common.Hash{}, big.NewInt(0), err
	}

	for _, signer := range c.witnesses {
		if witness == signer {
			c.setSigner(signer)
			log.Info("🐸 Found my witness", "witness", witness.String())
			return signer, witnessNext, witnessesHash, misc, nil
		}
	}
	return common.Address{}, common.Address{}, common.Hash{}, big.NewInt(0), ErrInvalidBlockWitness
}

func (c *Clique) getPreNode(fixedNumber uint64, nid common.Address) (common.Address, error) {
	err := c.freshCacheNodes(fixedNumber)
	if err != nil {
		return common.Address{}, err
	}
	nodesLen := int64(len(c.cacheNodes))
	for i := int64(0); i < nodesLen; i++ {
		if c.cacheNodes[i] == nid {
			if i == 0 {
				return c.cacheNodes[nodesLen-1], nil
			} else {
				return c.cacheNodes[i-1], nil
			}
		}
	}
	return common.Address{}, fmt.Errorf("Not found pre node of %s at number %d!", nid.String(), fixedNumber)
}

func (c *Clique) freshCacheNodes(number uint64) error {
	if number != c.cacheNumber || number == 0 {
		nodes, err := c.masternodeListFn(big.NewInt(int64(number)))
		if err != nil {
			return fmt.Errorf("Failed to get nodes at number %d!", number)
		}
		c.cacheNumber = number
		c.cacheNodes = nodes
		c.cacheHash = witnessesHash(nodes)
	}
	return nil
}

func (c *Clique) setSigner(signer common.Address) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.signer = signer
}

func (c *Clique) SetMasternodeListFn(masternodeListFn MasternodeListFn) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.masternodeListFn = masternodeListFn
}

func (c *Clique) SetInvestorFn(investorFn InvestorFn) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.investorFn = investorFn
}

// SealHash returns the hash of a block prior to it being sealed.
func SealHash(header *types.Header) (hash common.Hash) {
	hasher := sha3.NewLegacyKeccak256()
	encodeSigHeader(hasher, header)
	hasher.(crypto.KeccakState).Read(hash[:])
	return hash
}

// CliqueRLP returns the rlp bytes which needs to be signed for the proof-of-authority
// sealing. The RLP to sign consists of the entire header apart from the 65 byte signature
// contained at the end of the extra data.
//
// Note, the method requires the extra data to be at least 65 bytes, otherwise it
// panics. This is done to avoid accidentally using both forms (signature present
// or not), which could be abused to produce different hashes for the same header.
func CliqueRLP(header *types.Header) []byte {
	b := new(bytes.Buffer)
	encodeSigHeader(b, header)
	return b.Bytes()
}

func encodeSigHeader(w io.Writer, header *types.Header) {
	enc := []interface{}{
		header.ParentHash,
		header.UncleHash,
		header.Coinbase,
		header.Root,
		header.TxHash,
		header.ReceiptHash,
		header.Bloom,
		header.Difficulty,
		header.Number,
		header.GasLimit,
		header.GasUsed,
		header.Time,
		header.Extra[:len(header.Extra)-crypto.SignatureLength], // Yes, this will panic if extra is too short
		header.MixDigest,
		header.Nonce,
	}
	if header.BaseFee != nil {
		enc = append(enc, header.BaseFee)
	}
	if err := rlp.Encode(w, enc); err != nil {
		panic("can't encode: " + err.Error())
	}
}

func witnessesHash(witnesses []common.Address) (hash common.Hash) {
	hasher := sha3.NewLegacyKeccak256()
	if err := rlp.Encode(hasher, witnesses); err != nil {
		panic("can't encode witnesses: " + err.Error())
	}
	hasher.(crypto.KeccakState).Read(hash[:])
	return hash
}
