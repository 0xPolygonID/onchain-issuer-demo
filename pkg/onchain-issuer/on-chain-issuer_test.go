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
		"PK",
		big.NewInt(15), big.NewInt(5))
	require.NoError(t, err)
	fmt.Println(fmt.Sprintf("mined, block number %v, tx %v", r.BlockNumber, r.TxHash))
}
