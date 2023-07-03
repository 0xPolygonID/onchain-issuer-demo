package blockchain

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/iden3/go-merkletree-sql/v2"
	"github.com/iden3/go-schema-processor/verifiable"
)

const blockConfirmations = 2

// ProcessOnChainClaim send transaction to blockchain
// and wait for the transaction to be mined and confirmed (2 blocks)
func ProcessOnChainClaim(
	rpcUrl string,
	contractAddress string,
	pk string,
	hashIndex *big.Int,
	hashValue *big.Int,
) (verifiable.Iden3SparseMerkleTreeProof, error) {
	client, err := ethclient.Dial(rpcUrl)
	if err != nil {
		return verifiable.Iden3SparseMerkleTreeProof{}, err
	}
	defer client.Close()

	chid, err := client.ChainID(context.Background())
	if err != nil {
		return verifiable.Iden3SparseMerkleTreeProof{}, err
	}

	onChainIssuer, err := NewIdentity(
		common.HexToAddress(contractAddress),
		client,
	)
	if err != nil {
		return verifiable.Iden3SparseMerkleTreeProof{}, err
	}

	tx, err := sendTx(
		onChainIssuer,
		pk,
		chid,
		func(opts *bind.TransactOpts) (*types.Transaction, error) {
			return onChainIssuer.AddClaimHashAndTransit(
				opts,
				hashIndex,
				hashValue,
			)
		},
	)
	if err != nil {
		return verifiable.Iden3SparseMerkleTreeProof{}, err
	}

	if err = waitConfirmation(
		client,
		tx,
		blockConfirmations,
	); err != nil {
		return verifiable.Iden3SparseMerkleTreeProof{}, err
	}

	mtpProof, err := buildMTPProof(onChainIssuer, hashIndex)
	if err != nil {
		return verifiable.Iden3SparseMerkleTreeProof{}, err
	}

	return mtpProof, nil
}

func RevokeOnChainClaim(
	rpcUrl string,
	contractAddress string,
	pk string,
	nonce uint64,
) error {
	client, err := ethclient.Dial(rpcUrl)
	if err != nil {
		return err
	}
	defer client.Close()

	chid, err := client.ChainID(context.Background())
	if err != nil {
		return err
	}

	onChainIssuer, err := NewIdentity(
		common.HexToAddress(contractAddress),
		client,
	)
	if err != nil {
		return err
	}

	tx, err := sendTx(
		onChainIssuer,
		pk,
		chid,
		func(opts *bind.TransactOpts) (*types.Transaction, error) {
			return onChainIssuer.RevokeClaimAndTransit(
				opts,
				nonce,
			)
		},
	)
	if err != nil {
		return err
	}

	return waitConfirmation(
		client,
		tx,
		blockConfirmations,
	)
}

func IsRevockedClaim(
	rpcUrl string,
	contractAddress string,
	nonce uint64,
) (bool, error) {
	client, err := ethclient.Dial(rpcUrl)
	if err != nil {
		return false, err
	}
	defer client.Close()

	onChainIssuer, err := NewIdentity(
		common.HexToAddress(contractAddress),
		client,
	)
	if err != nil {
		return false, err
	}

	revocationProof, err := onChainIssuer.GetRevocationProof(&bind.CallOpts{}, nonce)
	if err != nil {
		return false, err
	}

	return revocationProof.Existence, nil
}

func sendTx(
	onChainIssuer *Identity,
	pk string,
	chid *big.Int,
	call func(*bind.TransactOpts) (*types.Transaction, error),
) (*types.Transaction, error) {
	privateKey, err := crypto.HexToECDSA(pk)
	if err != nil {
		return nil, err
	}

	opts, err := bind.NewKeyedTransactorWithChainID(
		privateKey,
		chid,
	)
	if err != nil {
		return nil, err
	}
	tx, err := call(opts)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func waitConfirmation(
	client *ethclient.Client,
	tx *types.Transaction,
	blockConfirmations int64,
) error {
	r, err := bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		return err
	}

	tick := time.NewTicker(5 * time.Second)
	defer tick.Stop()

	for {
		<-tick.C
		header, err := client.HeaderByNumber(context.Background(), nil)
		if err != nil {
			return err
		}
		passedBlocks := big.NewInt(0).Sub(header.Number, r.BlockNumber)
		if big.NewInt(blockConfirmations).Cmp(passedBlocks) == -1 {
			return nil
		}
	}
}

func buildMTPProof(
	onChainIssuer *Identity,
	claimIndexHash *big.Int,
) (verifiable.Iden3SparseMerkleTreeProof, error) {
	proof, err := onChainIssuer.GetClaimProof(&bind.CallOpts{}, claimIndexHash)
	if err != nil {
		return verifiable.Iden3SparseMerkleTreeProof{}, err
	}
	bigState, err := onChainIssuer.GetIdentityLatestState(&bind.CallOpts{})
	if err != nil {
		return verifiable.Iden3SparseMerkleTreeProof{}, err
	}
	roots, err := onChainIssuer.GetRootsByState(&bind.CallOpts{}, bigState)
	if err != nil {
		return verifiable.Iden3SparseMerkleTreeProof{}, err
	}

	state, err := merkletree.NewHashFromBigInt(bigState)
	if err != nil {
		return verifiable.Iden3SparseMerkleTreeProof{}, err
	}
	rootOfRoots, err := merkletree.NewHashFromBigInt(roots.RootsRoot)
	if err != nil {
		return verifiable.Iden3SparseMerkleTreeProof{}, err
	}
	claimTreeRoot, err := merkletree.NewHashFromBigInt(roots.ClaimsRoot)
	if err != nil {
		return verifiable.Iden3SparseMerkleTreeProof{}, err
	}
	revocationTreeRoot, err := merkletree.NewHashFromBigInt(roots.RevocationsRoot)
	if err != nil {
		return verifiable.Iden3SparseMerkleTreeProof{}, err
	}

	mtp, err := convertChainProofToMerkleProof(&proof)
	if err != nil {
		return verifiable.Iden3SparseMerkleTreeProof{}, err
	}

	return verifiable.Iden3SparseMerkleTreeProof{
		Type: verifiable.Iden3SparseMerkleTreeProofType,
		IssuerData: verifiable.IssuerData{
			// TODO(illia-korotia): empty roots should be nil?
			State: verifiable.State{
				RootOfRoots:        strpoint(rootOfRoots.Hex()),
				ClaimsTreeRoot:     strpoint(claimTreeRoot.Hex()),
				RevocationTreeRoot: strpoint(revocationTreeRoot.Hex()),
				Value:              strpoint(state.Hex()),
			},
		},
		MTP: mtp,
	}, nil
}

func convertChainProofToMerkleProof(proof *SmtLibProof) (*merkletree.Proof, error) {
	nodeAuxIndex, err := merkletree.NewHashFromBigInt(
		proof.AuxIndex,
	)
	if err != nil {
		return nil, err
	}
	nodeAuxValue, err := merkletree.NewHashFromBigInt(
		proof.AuxValue,
	)
	if err != nil {
		return nil, err
	}

	siblings := make([]*merkletree.Hash, 0, len(proof.Siblings))
	for _, s := range proof.Siblings {
		h, err := merkletree.NewHashFromBigInt(s)
		if err != nil {
			return nil, err
		}
		siblings = append(siblings, h)
	}

	return merkletree.NewProofFromData(
		proof.Existence,
		siblings,
		&merkletree.NodeAux{nodeAuxIndex, nodeAuxValue},
	)
}

func strpoint(s string) *string {
	return &s
}
