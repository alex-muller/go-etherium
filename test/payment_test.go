package test

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/assert"
	"goeth/build/bindings"
	"log"
	"math/big"
	"testing"
	"time"
)

func TestDeploy(t *testing.T) {
	ctx := context.Background()
	
	key, err := crypto.GenerateKey()
	if err != nil {
		log.Fatalf("Failed to generate key: %v", err)
	}
	
	// Since we are using a simulated backend, we will get the chain ID
	// from the same place that the simulated backend gets it.
	chainID := params.AllDevChainProtocolChanges.ChainID
	
	auth, err := bind.NewKeyedTransactorWithChainID(key, chainID)
	assert.Empty(t, err)
	
	sim := simulated.NewBackend(map[common.Address]types.Account{
		auth.From: {Balance: big.NewInt(9e18)},
	})
	
	cl := sim.Client()
	
	userBalanceBefore, err := sim.Client().BalanceAt(context.Background(), auth.From, nil)
	fmt.Printf(`user balance before: %s`, userBalanceBefore)
	fmt.Println()
	
	// Deploy
	addr, tx, payments, err := bindings.DeployPayments(auth, cl)
	assert.Empty(t, err)
	assert.NotEmpty(t, addr)
	fmt.Printf("Deploy pending: 0x%x\n", tx.Hash())
	
	sim.Commit()
	
	// Check balance
	balance, err := payments.CurrentBalance(nil)
	assert.Empty(t, err)
	
	assert.Equal(t, `0`, balance.String())
	
	balanceAt, err := sim.Client().BalanceAt(ctx, addr, nil)
	assert.Empty(t, err)
	assert.Equal(t, `0`, balanceAt.String())
	
	// Send
	var message = `hello!`
	
	pay := *auth
	pay.Value = big.NewInt(900000000000000000)
	
	tr, err := payments.Pay(&pay, message)
	assert.Empty(t, err)
	assert.NotEmpty(t, tr)
	
	sim.Commit()
	
	payments, err = bindings.NewPayments(addr, cl)
	assert.Empty(t, err)
	
	// Check balance
	balance, err = payments.CurrentBalance(nil)
	assert.Empty(t, err)
	
	assert.Equal(t, `900000000000000000`, balance.String())
	
	balanceAt, err = sim.Client().BalanceAt(context.Background(), addr, nil)
	assert.Empty(t, err)
	assert.Equal(t, `900000000000000000`, balanceAt.String())
	
	// Check user balance
	userBalanceAfter, err := sim.Client().BalanceAt(context.Background(), auth.From, nil)
	fmt.Printf(`user balance after: %s`, userBalanceAfter)
	
	// Check payment
	payment, err := payments.GetPayment(nil, auth.From, big.NewInt(0))
	assert.Empty(t, err)
	
	assert.Equal(t, message, payment.Message)
	assert.Equal(t, auth.From, payment.From)
	assert.Equal(t, `900000000000000000`, payment.Amount.String())
	assert.WithinDuration(t, time.Now(), time.Unix(payment.Timestamp.Int64(), 0), time.Second)
}
