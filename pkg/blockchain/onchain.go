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
		hashIndex,
		hashValue,
		chid,
	)
	if err != nil {
		return verifiable.Iden3SparseMerkleTreeProof{}, err
	}

	r, err := bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		return verifiable.Iden3SparseMerkleTreeProof{}, err
	}

	tick := time.NewTicker(5 * time.Second)
	defer tick.Stop()

	for {
		<-tick.C
		header, err := client.HeaderByNumber(context.Background(), nil)
		if err != nil {
			return verifiable.Iden3SparseMerkleTreeProof{}, err
		}
		passedBlocks := big.NewInt(0).Sub(header.Number, r.BlockNumber)
		if big.NewInt(blockConfirmations).Cmp(passedBlocks) == -1 {
			break
		}
	}

	mtpProof, err := buildMTPProof(onChainIssuer, hashIndex)
	if err != nil {
		return verifiable.Iden3SparseMerkleTreeProof{}, err
	}

	return mtpProof, nil
}

func sendTx(
	onChainIssuer *Identity,
	pk string,
	hashIndex *big.Int,
	hashValue *big.Int,
	chid *big.Int,
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
	tx, err := onChainIssuer.AddClaimHashAndTransit(
		opts, hashIndex, hashValue,
	)
	if err != nil {
		return nil, err
	}

	return tx, nil
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

	rootOfRoots := roots.RevocationsRoot.Text(16)
	claimTreeRoot := roots.ClaimsRoot.Text(16)
	revocationTreeRoot := roots.RevocationsRoot.Text(16)
	state := bigState.Text(16)

	mtp, err := convertChainProofToMerkleProof(&proof)
	if err != nil {
		return verifiable.Iden3SparseMerkleTreeProof{}, err
	}

	return verifiable.Iden3SparseMerkleTreeProof{
		Type: verifiable.Iden3SparseMerkleTreeProofType,
		IssuerData: verifiable.IssuerData{
			State: verifiable.State{
				RootOfRoots:        &rootOfRoots,
				ClaimsTreeRoot:     &claimTreeRoot,
				RevocationTreeRoot: &revocationTreeRoot,
				Value:              &state,
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
