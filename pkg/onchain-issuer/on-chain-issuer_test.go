package onChainIssuer

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestModuleCall(t *testing.T) {
	r, err := AddClaimHashAndTransitAndWait("https://polygon-mumbai.g.alchemy.com/v2/NylY7Qhw31V8fPfstqzasLEFKcjmDJYs",
		"0xf3bB959314B5D1e4587e1f597ccc289216608ac5",
		"794b190c537189d5b74440122ea1a91546164fc887673f8155665c334d88912d",
		big.NewInt(15), big.NewInt(5))
	require.NoError(t, err)
	fmt.Println(fmt.Sprintf("mined, block number %v, tx %v", r.BlockNumber, r.TxHash))
}

// func TestConnectionAuth(t *testing.T) {
// 	client, _ := ethclient.Dial("https://polygon-mumbai.g.alchemy.com/v2/NylY7Qhw31V8fPfstqzasLEFKcjmDJYs")
// 	// IdentityExample address
// 	address := common.HexToAddress("0xf3bB959314B5D1e4587e1f597ccc289216608ac5")
// 	// create contract connection
// 	instance, _ := NewIdentity(address, client)

// 	owner, _ := instance.Owner(nil)
// 	fmt.Println(fmt.Sprintf("owner %v", owner))

// 	id, _ := instance.GetId(nil)
// 	fmt.Println(fmt.Sprintf("identityId %v", id))

// 	lastClaimsRoot, _ := instance.GetLastClaimsRoot(nil)
// 	fmt.Println(fmt.Sprintf("lastClaimsRoot %v", lastClaimsRoot))

// 	// trx
// 	// private key
// 	privateKey, _ := crypto.HexToECDSA("794b190c537189d5b74440122ea1a91546164fc887673f8155665c334d88912d")
// 	publicKey := privateKey.Public()
// 	publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)

// 	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

// 	nonce, _ := client.PendingNonceAt(context.Background(), fromAddress)
// 	fmt.Println(fmt.Sprintf("nonce %v", nonce))
// 	gasPrice, _ := client.SuggestGasPrice(context.Background())
// 	fmt.Println(fmt.Sprintf("gasPrice %v", gasPrice))

// 	auth, _ := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(80001))
// 	auth.Nonce = big.NewInt(int64(nonce))
// 	auth.Value = big.NewInt(0)      // in wei
// 	auth.GasLimit = uint64(3000000) // in units
// 	auth.GasPrice = gasPrice

// 	tx, err := instance.AddClaimHashAndTransit(auth, big.NewInt(2), big.NewInt(3))
// 	require.NoError(t, err)
// 	fmt.Println(fmt.Sprintf("tx sent: %s", tx.Hash().Hex()))
// 	r, _ := bind.WaitMined(context.Background(), client, tx)
// 	fmt.Println(fmt.Sprintf("mined, block number %v", r.BlockNumber))
// 	fmt.Println(fmt.Sprintf("error %v", r.Status))
// 	for {
// 		header, _ := client.HeaderByNumber(context.Background(), nil)
// 		passedBlocks := big.NewInt(0).Sub(header.Number, r.BlockNumber)
// 		fmt.Println(fmt.Sprintf("passed blocks %v", passedBlocks))
// 		if big.NewInt(10).Cmp(passedBlocks) == -1 {
// 			break
// 		}
// 	}

// }

// func TestConnectionTxEip1559(t *testing.T) {
// 	client, _ := ethclient.Dial("https://polygon-mumbai.g.alchemy.com/v2/NylY7Qhw31V8fPfstqzasLEFKcjmDJYs")
// 	// IdentityExample address
// 	address := common.HexToAddress("0xf3bB959314B5D1e4587e1f597ccc289216608ac5")
// 	// create contract connection
// 	instance, _ := NewIdentity(address, client)

// 	owner, _ := instance.Owner(nil)
// 	fmt.Println(fmt.Sprintf("owner %v", owner))

// 	id, _ := instance.GetId(nil)
// 	fmt.Println(fmt.Sprintf("identityId %v", id))

// 	lastClaimsRoot, _ := instance.GetLastClaimsRoot(nil)
// 	fmt.Println(fmt.Sprintf("lastClaimsRoot %v", lastClaimsRoot))

// 	// trx
// 	// private key
// 	privateKey, _ := crypto.HexToECDSA("794b190c537189d5b74440122ea1a91546164fc887673f8155665c334d88912d")
// 	publicKey := privateKey.Public()
// 	publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)

// 	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

// 	nonce, _ := client.PendingNonceAt(context.Background(), fromAddress)
// 	fmt.Println(fmt.Sprintf("nonce %v", nonce))

// 	ab, _ := IdentityMetaData.GetAbi()
// 	payload, err := ab.Pack("addClaimHashAndTransit", big.NewInt(12), big.NewInt(11))
// 	require.NoError(t, err)

// 	// gasLimit := uint64(9000000) // in units
// 	gasLimit, err := client.EstimateGas(context.Background(), ethereum.CallMsg{
// 		From:  fromAddress, // the sender of the 'transaction'
// 		To:    &address,
// 		Gas:   0,             // wei <-> gas exchange ratio
// 		Value: big.NewInt(0), // amount of wei sent along with the call
// 		Data:  payload,
// 	})
// 	require.NoError(t, err)
// 	fmt.Println(fmt.Sprintf("gasLimit: %v", gasLimit))
// 	value := big.NewInt(0)

// 	cid, _ := client.ChainID(context.Background())

// 	latestBlockHeader, err := client.HeaderByNumber(context.Background(), nil)
// 	require.NoError(t, err)

// 	baseFee := misc.CalcBaseFee(&params.ChainConfig{LondonBlock: big.NewInt(1)}, latestBlockHeader)
// 	fmt.Println(fmt.Sprintf("baseFee calculated: %v", baseFee))
// 	b := math.Round(float64(baseFee.Int64()) * 19.5)
// 	baseFee = big.NewInt(int64(b))
// 	fmt.Println(fmt.Sprintf("baseFee: %v", baseFee))
// 	gasTip, err := client.SuggestGasTipCap(context.Background())
// 	fmt.Println(fmt.Sprintf("gasTip: %v", gasTip))
// 	gasTip = big.NewInt(9200999983)
// 	require.NoError(t, err)

// 	maxGasPricePerFee := big.NewInt(0).Add(baseFee, gasTip)

// 	tx := types.NewTx(&types.DynamicFeeTx{
// 		ChainID:   cid,
// 		Nonce:     nonce,
// 		GasFeeCap: maxGasPricePerFee,
// 		GasTipCap: gasTip,
// 		Gas:       gasLimit,
// 		To:        &address,
// 		Value:     value,
// 		Data:      payload,
// 	})

// 	signedTx, _ := types.SignTx(tx, types.LatestSignerForChainID(cid), privateKey)

// 	client.SendTransaction(context.Background(), signedTx)

// 	fmt.Println(fmt.Sprintf("tx sent: %s", signedTx.Hash().Hex()))

// 	r, _ := bind.WaitMined(context.Background(), client, signedTx)
// 	fmt.Println(fmt.Sprintf("mined, block number %v", r.BlockNumber))

// }
