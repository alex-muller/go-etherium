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

	// Watch event
	sink := make(chan *bindings.DemoPaid)
	paid, err := demo.WatchPaid(&bind.WatchOpts{}, sink, nil)

	_ = paid

	go func() {
		for demoPaid := range sink {
			// Check that paid
			fmt.Printf(`paid: %v`, demoPaid)
		}
	}()

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

	// Watch event
	sink := make(chan *bindings.DemoPaid)
	paid, err := demo.WatchPaid(&bind.WatchOpts{}, sink, nil)

	_ = paid

	go func() {
		for demoPaid := range sink {
			// Check that paid
			fmt.Printf(`paid: %v`, demoPaid)
		}
	}()

	time.Sleep(1 * time.Second)

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
