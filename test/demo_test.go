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
	"goeth/build/bindings"
	"log"
	"math/big"
	"testing"
	"time"
)

func TestDemoItShouldAllowSendMoney(t *testing.T) {
	key, err := crypto.GenerateKey()
	if err != nil {
		log.Fatalf("Failed to generate privateKey: %v", err)
	}

	// Since we are using a simulated backend, we will get the chain ID
	// from the same place that the simulated backend gets it.
	chainID := params.AllDevChainProtocolChanges.ChainID

	auth, err := bind.NewKeyedTransactorWithChainID(key, chainID)
	assert.Empty(t, err)

	sim := simulated.NewBackend(map[common.Address]types.Account{
		auth.From: {Balance: big.NewInt(1e18)},
	})

	cl := sim.Client()

	// Deploy
	demoAddr, tx, demo, err := bindings.DeployDemo(auth, cl)
	assert.Empty(t, err)
	assert.NotEmpty(t, demoAddr)
	fmt.Printf("Deploy pending: 0x%x\n", tx.Hash())

	userBalanceAfter, err := cl.BalanceAt(context.Background(), auth.From, nil)
	fmt.Printf("user balance before: %s\n", userBalanceAfter)

	sim.Commit()

	// Retrieve the pending nonce
	auth.Value = new(big.Int).Mul(big.NewInt(12345600), big.NewInt(params.GWei))
	pay, err := demo.Receive(auth)
	assert.Empty(t, err)
	assert.NotEmpty(t, pay)

	sim.Commit()

	// Check user balance
	userBalanceAfter, err = cl.BalanceAt(context.Background(), auth.From, nil)
	fmt.Printf("user balance after: %s\n", userBalanceAfter)

	demoBalance, err := cl.BalanceAt(context.Background(), demoAddr, nil)
	fmt.Printf("demo balance after: %s\n", demoBalance)
}

func TestDemoItShouldAllowSendMoneyOnDevDeployed(t *testing.T) {
	ctx := context.Background()
	_ = ctx

	ks := keystore.NewKeyStore("./../build/dev-chain/keystore", keystore.StandardScryptN, keystore.StandardScryptP)
	am := accounts.NewManager(&accounts.Config{InsecureUnlockAllowed: false}, ks)
	wallets := am.Wallets()
	fromAcc := wallets[0].Accounts()[0]

	// Since we are using a simulated backend, we will get the chain ID
	// from the same place that the simulated backend gets it.

	cl, err := ethclient.Dial("./../build/dev-chain/geth.ipc")
	assert.Empty(t, err)

	chainID, err := cl.ChainID(ctx)
	assert.Empty(t, err)

	auth, err := bind.NewKeyStoreTransactorWithChainID(ks, fromAcc, chainID)
	assert.Empty(t, err)

	err = ks.Unlock(fromAcc, "")
	assert.Empty(t, err)

	userBalanceAfter, err := cl.BalanceAt(context.Background(), auth.From, nil)
	fmt.Printf("user balance bef d: %s\n", userBalanceAfter)

	time.Sleep(2 * time.Second)

	// Try to get
	demoAddr := common.HexToAddress("0x119202429f307724e7c1a36371205c05421f2298")

	demo, err := bindings.NewDemo(demoAddr, cl)
	assert.Empty(t, err)

	auth.Value = new(big.Int).Mul(big.NewInt(12345600), big.NewInt(params.GWei))
	pay, err := demo.Receive(auth)
	assert.Empty(t, err)
	assert.NotEmpty(t, pay)

	userBalanceAfter, err = cl.BalanceAt(context.Background(), auth.From, nil)
	fmt.Printf("user balance bef t: %s\n", userBalanceAfter)

	// Retrieve the pending nonce
	nonce, err := cl.PendingNonceAt(context.Background(), auth.From)
	assert.Empty(t, err)

	// Get suggested gas price
	tipCap, _ := cl.SuggestGasTipCap(context.Background())
	feeCap, _ := cl.SuggestGasPrice(context.Background())

	// Create a new transaction
	sendTx := types.NewTx(
		&types.DynamicFeeTx{
			ChainID:   chainID,
			Nonce:     nonce,
			GasTipCap: tipCap,
			GasFeeCap: feeCap,
			Gas:       uint64(22000),
			To:        &demoAddr,
			Value:     new(big.Int).Mul(big.NewInt(12345600), big.NewInt(params.GWei)),
			Data:      nil,
		})

	signTx, err := ks.SignTxWithPassphrase(fromAcc, "", sendTx, chainID)
	assert.Empty(t, err)

	err = cl.SendTransaction(ctx, signTx)
	assert.Empty(t, err)

	time.Sleep(time.Second * 2)

	// Check user balance
	userBalanceAfter, err = cl.BalanceAt(context.Background(), auth.From, nil)
	fmt.Printf("user balance aft t: %s\n", userBalanceAfter)

	demoBalance, err := cl.BalanceAt(context.Background(), demoAddr, nil)
	fmt.Printf("demo balance aft: %s\n", demoBalance)

	number, err := cl.BlockNumber(ctx)
	fmt.Printf("block number: %d\n", number)

	_ = demo
}

func TestDemoItShouldAllowSendMoneyOnDev(t *testing.T) {
	ctx := context.Background()
	_ = ctx

	ks := keystore.NewKeyStore("./../build/dev-chain/keystore", keystore.StandardScryptN, keystore.StandardScryptP)
	am := accounts.NewManager(&accounts.Config{InsecureUnlockAllowed: false}, ks)
	wallets := am.Wallets()
	fromAcc := wallets[0].Accounts()[0]

	cl, err := ethclient.Dial("./../build/dev-chain/geth.ipc")
	assert.Empty(t, err)

	chainID, err := cl.ChainID(ctx)
	assert.Empty(t, err)

	auth, err := bind.NewKeyStoreTransactorWithChainID(ks, fromAcc, chainID)
	assert.Empty(t, err)

	err = ks.Unlock(fromAcc, "")
	assert.Empty(t, err)

	userBalanceAfter, err := cl.BalanceAt(context.Background(), auth.From, nil)
	fmt.Printf("user balance bef d: %s\n", userBalanceAfter)

	// Deploy
	demoAddr, tx, demo, err := bindings.DeployDemo(auth, cl)
	assert.Empty(t, err)
	assert.NotEmpty(t, demoAddr)
	fmt.Printf("Deploy pending: 0x%x\n", tx.Hash())
	fmt.Printf("Deploy addr: 0x%x\n", demoAddr)

	time.Sleep(2 * time.Second)

	hash, pending, err := cl.TransactionByHash(ctx, tx.Hash())
	assert.Empty(t, err)

	_, _ = hash, pending

	// Try to get
	demo, err = bindings.NewDemo(demoAddr, cl)
	assert.Empty(t, err)

	owner, err := demo.Owner(nil)
	assert.Empty(t, err)

	_ = owner

	userBalanceAfter, err = cl.BalanceAt(context.Background(), auth.From, nil)
	fmt.Printf("user balance bef t: %s\n", userBalanceAfter)

	// Retrieve the pending nonce
	auth.Value = new(big.Int).Mul(big.NewInt(12345600), big.NewInt(params.GWei))
	pay, err := demo.Receive(auth)
	assert.Empty(t, err)
	assert.NotEmpty(t, pay)

	time.Sleep(time.Second * 2)

	// Check user balance
	userBalanceAfter, err = cl.BalanceAt(context.Background(), auth.From, nil)
	fmt.Printf("user balance aft t: %s\n", userBalanceAfter)

	demoBalance, err := cl.BalanceAt(context.Background(), demoAddr, nil)
	fmt.Printf("demo balance aft: %s\n", demoBalance)

	number, err := cl.BlockNumber(ctx)
	fmt.Printf("block number: %d\n", number)

	_ = demo
}

// 115792089237316195423570985008687907853269984665640564039457584007913129639927
// 115792089237316195423570985008687907853269984665640564039457583908483379526293

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
