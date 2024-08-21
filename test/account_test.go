package test

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/assert"
	"log"
	"math/big"
	"testing"
	"time"
)

func TestSendEtherUsingSimulatedBeTest(t *testing.T) {
	key, err := crypto.GenerateKey()
	if err != nil {
		log.Fatalf("Failed to generate privateKey: %v", err)
	}

	chainID := params.AllDevChainProtocolChanges.ChainID

	auth, err := bind.NewKeyedTransactorWithChainID(key, chainID)

	key2, err := crypto.GenerateKey()
	if err != nil {
		log.Fatalf("Failed to generate key: %v", err)
	}

	to, err := bind.NewKeyedTransactorWithChainID(key2, chainID)

	sim := simulated.NewBackend(map[common.Address]types.Account{
		auth.From: {Balance: big.NewInt(9e18)},
		to.From:   {},
	})

	cl := sim.Client()

	// Before
	senderBalance, err := cl.BalanceAt(context.Background(), auth.From, nil)
	assert.Empty(t, err)

	receiverBalance, err := cl.BalanceAt(context.Background(), to.From, nil)
	assert.Empty(t, err)

	fmt.Printf("Sender balance before: %d\n", senderBalance)
	fmt.Printf("Receiver balance before: %d\n", receiverBalance)

	// Retrieve the pending nonce
	nonce, err := cl.PendingNonceAt(context.Background(), auth.From)
	assert.Empty(t, err)

	// Get suggested gas price
	tipCap, _ := cl.SuggestGasTipCap(context.Background())
	feeCap, _ := cl.SuggestGasPrice(context.Background())

	// Create a new transaction
	tx := types.NewTx(
		&types.DynamicFeeTx{
			ChainID:   chainID,
			Nonce:     nonce,
			GasTipCap: tipCap,
			GasFeeCap: feeCap,
			Gas:       uint64(21000),
			To:        &to.From,
			Value:     new(big.Int).Mul(big.NewInt(1), big.NewInt(params.Ether)),
			Data:      nil,
		})

	signer := types.LatestSignerForChainID(chainID)
	signTx, err := types.SignTx(tx, signer, key)

	err = cl.SendTransaction(context.Background(), signTx)
	assert.Empty(t, err)

	sim.Commit()

	// TODO wailt

	blockNumber, err := cl.BlockNumber(context.Background())
	fmt.Printf("Block number: %d\n", blockNumber)

	//time.Sleep(time.Second)
	blockNumber, err = cl.BlockNumber(context.Background())
	fmt.Printf("Block number: %d\n", blockNumber)

	_ = blockNumber

	senderBalance, err = cl.BalanceAt(context.Background(), auth.From, nil)
	assert.Empty(t, err)

	receiverBalance, err = cl.BalanceAt(context.Background(), to.From, nil)
	assert.Empty(t, err)

	fmt.Printf("Sender balance after: %d\n", senderBalance)
	fmt.Printf("Receiver balance after: %d\n", receiverBalance)
}

func TestSendEtherUsingGethDevMode(t *testing.T) {
	ks := keystore.NewKeyStore("./../build/dev-chain/keystore", keystore.StandardScryptN, keystore.StandardScryptP)
	am := accounts.NewManager(&accounts.Config{InsecureUnlockAllowed: false}, ks)
	wallets := am.Wallets()
	fromAcc := wallets[0].Accounts()[0]

	toacc, err := ks.NewAccount(`asddsa789753awewqeq@ew`)
	assert.Empty(t, err)

	toAddr := toacc.Address

	cl, err := ethclient.Dial("./../build/dev-chain/geth.ipc")
	assert.Empty(t, err)

	// Before
	senderBalance, err := cl.BalanceAt(context.Background(), fromAcc.Address, nil)
	assert.Empty(t, err)

	receiverBalance, err := cl.BalanceAt(context.Background(), toAddr, nil)
	assert.Empty(t, err)

	fmt.Printf("Sender balance before: %d\n", senderBalance)
	fmt.Printf("Receiver balance before: %d\n", receiverBalance)

	// Retrieve the chainid (needed for signer)
	chainid, err := cl.ChainID(context.Background())
	assert.Empty(t, err)

	// Retrieve the pending nonce
	nonce, err := cl.PendingNonceAt(context.Background(), fromAcc.Address)
	assert.Empty(t, err)

	// Get suggested gas price
	tipCap, _ := cl.SuggestGasTipCap(context.Background())
	feeCap, _ := cl.SuggestGasPrice(context.Background())

	// Create a new transaction
	tx := types.NewTx(
		&types.DynamicFeeTx{
			ChainID:   chainid,
			Nonce:     nonce,
			GasTipCap: tipCap,
			GasFeeCap: feeCap,
			Gas:       uint64(21000),
			To:        &toAddr,
			Value:     new(big.Int).Mul(big.NewInt(1), big.NewInt(params.Ether)),
			Data:      nil,
		})

	signTx, err := ks.SignTxWithPassphrase(fromAcc, "", tx, chainid)
	assert.Empty(t, err)

	err = cl.SendTransaction(context.Background(), signTx)
	assert.Empty(t, err)

	blockNumber, err := cl.BlockNumber(context.Background())
	fmt.Printf("Block number: %d\n", blockNumber)

	time.Sleep(time.Second)

	blockNumber, err = cl.BlockNumber(context.Background())
	fmt.Printf("Block number: %d\n", blockNumber)

	senderBalance, err = cl.BalanceAt(context.Background(), fromAcc.Address, nil)
	assert.Empty(t, err)

	receiverBalance, err = cl.BalanceAt(context.Background(), toAddr, nil)
	assert.Empty(t, err)

	fmt.Printf("Sender balance after: %d\n", senderBalance)
	fmt.Printf("Receiver balance after: %d\n", receiverBalance)
}
