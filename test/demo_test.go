package test

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
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
	"goeth/build/bindings"
	"log"
	"math/big"
	"testing"
	"time"
)

func TestDemoItShouldAllowSendMoney(t *testing.T) {
	ctx := context.Background()

	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatalf("Failed to generate privateKey: %v", err)
	}

	// Since we are using a simulated backend, we will get the chain ID
	// from the same place that the simulated backend gets it.
	chainID := params.AllDevChainProtocolChanges.ChainID

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	assert.Empty(t, err)

	sim := simulated.NewBackend(map[common.Address]types.Account{
		auth.From: {Balance: big.NewInt(9e18)},
	})

	cl := sim.Client()

	// Deploy
	demoAddr, tx, _, err := bindings.DeployDemo(auth, cl)
	assert.Empty(t, err)
	assert.NotEmpty(t, demoAddr)
	fmt.Printf("Deploy pending: 0x%x\n", tx.Hash())

	userBalanceAfter, err := cl.BalanceAt(context.Background(), auth.From, nil)
	fmt.Printf("user balance before: %s\n", userBalanceAfter)

	sim.Commit()

	// Retrieve the pending nonce
	nonce, err := cl.PendingNonceAt(context.Background(), auth.From)
	assert.Empty(t, err)

	// Send transaction
	sendTx := types.NewTx(&types.DynamicFeeTx{
		ChainID:    chainID,
		Nonce:      nonce,
		GasTipCap:  big.NewInt(5),
		GasFeeCap:  big.NewInt(5),
		Gas:        54321,
		To:         &demoAddr,
		Value:      big.NewInt(200000000000000000),
		Data:       nil,
		AccessList: nil,
		V:          nil,
		R:          nil,
		S:          nil,
	})

	signedTx, err := types.SignTx(sendTx, types.NewLondonSigner(chainID), privateKey)
	assert.Empty(t, err)

	err = cl.SendTransaction(ctx, signedTx)
	assert.Empty(t, err)
	sim.Commit()

	// LOGS

	q := ethereum.FilterQuery{}

	logs, err := cl.FilterLogs(context.Background(), q)
	assert.Empty(t, err)

	fmt.Printf("logs: %+v\n", logs)

	// Check user balance
	userBalanceAfter, err = cl.BalanceAt(context.Background(), auth.From, nil)
	fmt.Printf("user balance after: %s\n", userBalanceAfter)

	demoBalance, err := cl.BalanceAt(context.Background(), demoAddr, nil)
	fmt.Printf("demo balance after: %s\n", demoBalance)

	//fmt.Printf(``, demo.)/
}

func TestSendEtherUsingSimulatedBeTest(t *testing.T) {
	ks := keystore.NewKeyStore("./../build/dev-chain/keystore", keystore.StandardScryptN, keystore.StandardScryptP)
	am := accounts.NewManager(&accounts.Config{InsecureUnlockAllowed: false}, ks)
	wallets := am.Wallets()
	fromAcc := wallets[0].Accounts()[0]
	toAddr := wallets[1].Accounts()[0].Address

	//cl, err := ethclient.Dial("./../build/dev-chain/geth.ipc")
	//assert.Empty(t, err)

	sim := simulated.NewBackend(map[common.Address]types.Account{
		fromAcc.Address: {Balance: big.NewInt(9e18)},
		toAddr:          {},
	})

	cl := sim.Client()

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

	sim.Commit()

	// TODO wailt

	blockNumber, err := cl.BlockNumber(context.Background())
	fmt.Printf("Block number: %d\n", blockNumber)

	//time.Sleep(time.Second)
	blockNumber, err = cl.BlockNumber(context.Background())
	fmt.Printf("Block number: %d\n", blockNumber)

	_ = blockNumber

	senderBalance, err = cl.BalanceAt(context.Background(), fromAcc.Address, nil)
	assert.Empty(t, err)

	receiverBalance, err = cl.BalanceAt(context.Background(), toAddr, nil)
	assert.Empty(t, err)

	fmt.Printf("Sender balance after: %d\n", senderBalance)
	fmt.Printf("Receiver balance after: %d\n", receiverBalance)
}

func TestSendEtherUsingGethDevMode(t *testing.T) {
	ks := keystore.NewKeyStore("./../build/dev-chain/keystore", keystore.StandardScryptN, keystore.StandardScryptP)
	am := accounts.NewManager(&accounts.Config{InsecureUnlockAllowed: false}, ks)
	wallets := am.Wallets()
	fromAcc := wallets[0].Accounts()[0]
	toAddr := wallets[1].Accounts()[0].Address

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
