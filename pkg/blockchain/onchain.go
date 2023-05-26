package onChainIssuer

import (
	"context"
	"crypto/ecdsa"
	"math"
	"math/big"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/misc"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
)

func AddClaimHashAndTransitAndWait(rpcUrl string, contractAddress string, pk string, hashIndex *big.Int, hashValue *big.Int) (*types.Receipt, error) {
	client, err := ethclient.Dial(rpcUrl)
	if err != nil {
		return nil, err
	}

	tx, err := SendSignedTxAddClaimHashAndTransit(client, pk, contractAddress, hashIndex, hashValue)

	if err != nil {
		return nil, err
	}
	r, err := bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		return nil, err
	}
	for {
		header, _ := client.HeaderByNumber(context.Background(), nil)
		passedBlocks := big.NewInt(0).Sub(header.Number, r.BlockNumber)
		if big.NewInt(2).Cmp(passedBlocks) == -1 {
			break
		}
	}
	return r, nil
}

func SendSignedTxAddClaimHashAndTransit(client *ethclient.Client, pk string, contractAddress string,
	hashIndex *big.Int, hashValue *big.Int) (*types.Transaction, error) {
	privateKey, _ := crypto.HexToECDSA(pk)
	publicKey := privateKey.Public()
	publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	toAddress := common.HexToAddress(contractAddress)

	nonce, _ := client.PendingNonceAt(context.Background(), fromAddress)

	ab, _ := IdentityMetaData.GetAbi()
	payload, err := ab.Pack("addClaimHashAndTransit", hashIndex, hashValue)
	if err != nil {
		return nil, err
	}

	// gasLimit := uint64(9000000) // in units
	gasLimit, err := client.EstimateGas(context.Background(), ethereum.CallMsg{
		From:  fromAddress, // the sender of the 'transaction'
		To:    &toAddress,
		Gas:   0,             // wei <-> gas exchange ratio
		Value: big.NewInt(0), // amount of wei sent along with the call
		Data:  payload,
	})
	if err != nil {
		return nil, err
	}
	value := big.NewInt(0)

	cid, _ := client.ChainID(context.Background())

	latestBlockHeader, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	baseFee := misc.CalcBaseFee(&params.ChainConfig{LondonBlock: big.NewInt(1)}, latestBlockHeader)
	b := math.Round(float64(baseFee.Int64()) * 1.25)
	baseFee = big.NewInt(int64(b))
	gasTip, err := client.SuggestGasTipCap(context.Background())
	if err != nil {
		return nil, err
	}

	maxGasPricePerFee := big.NewInt(0).Add(baseFee, gasTip)

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   cid,
		Nonce:     nonce,
		GasFeeCap: maxGasPricePerFee,
		GasTipCap: gasTip,
		Gas:       gasLimit,
		To:        &toAddress,
		Value:     value,
		Data:      payload,
	})

	signedTx, _ := types.SignTx(tx, types.LatestSignerForChainID(cid), privateKey)

	client.SendTransaction(context.Background(), signedTx)

	return signedTx, nil
}
