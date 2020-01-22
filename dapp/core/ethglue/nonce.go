package ethglue

import (
	"context"
	"math/big"
	"sync"
	"sync/atomic"
	"time"

	"github.com/labstack/gommon/log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type NonceManager struct {
	nextNonce       *big.Int
	lastNonceChange time.Time

	ethClient  *ethclient.Client
	ethAccount common.Address

	errStreakCount int64
	mu             sync.Mutex // protects next nonce updates
}

const maxErrStreak = 3
const idleSyncTimeInMinutes = 15

func (n *NonceManager) NextNonce() *big.Int {
	n.mu.Lock()
	defer n.mu.Unlock()

	if time.Now().Sub(n.lastNonceChange).Minutes() >= idleSyncTimeInMinutes {
		// this is the only way to fix the gap (when our nonce is higher than it should be)
		n.syncNonce()
		no := new(big.Int).Set(n.nextNonce)
		n.nextNonce.Add(n.nextNonce, big.NewInt(1))
		return no
	}

	no := new(big.Int).Set(n.nextNonce)
	n.nextNonce.Add(n.nextNonce, big.NewInt(1))
	n.lastNonceChange = time.Now()
	return no
}

func (n *NonceManager) OnAccountChange(addr string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.ethAccount = common.HexToAddress(addr)
	// nonce is per eth account
	n.syncNonce()
}

func (n *NonceManager) OnDial(c *ethclient.Client) {
	n.ethClient = c
}

func (n *NonceManager) OnError(err error) {
	if err == nil {
		atomic.StoreInt64(&n.errStreakCount, 0)
		return
	}
	n.mu.Lock()
	defer n.mu.Unlock()

	if err.Error() == "nonce too low" {
		// either transaction(s) from different machine or
		// somehow transaction not failing instantly but not increasing nonce as well
		// TODO(mmal): investigate theoretical possibility of the second case
		n.nextNonce = n.pendingNonceFromNode()
		log.Errorf("[NonceManager]: synced (forward jump) pendingNonce with network: %v", n.nextNonce.Int64())
		return
	}

	if err.Error() == "gas required exceeds allowance or always failing transaction" {
		// network nonce was for sure not increased - we reuse previous one
		n.nextNonce.Add(n.nextNonce, big.NewInt(-1))
		log.Errorf("[NonceManager]: decreasing nonce due to err: %v", err)
		return
	}

	// generic case unsure what to do
	n.errStreakCount++
	if n.errStreakCount >= maxErrStreak {
		n.errStreakCount = 0
		//TODO(mmal): we should cancel transactions if decreasing nonce (filling the gap) but
		// it is a eth protocol "workaround" as there is no way to officially cancel
		// and costly - requires fake transactions with >10% more gas price
		// looks like when sending new transaction with repeating nonce
		// the transaction with higher price will win (old or new) which means that when the
		// gap closes (our pending nonce reaches network nonce) we will have old stalled
		// transaction executed - possibly out of order and partially
		n.nextNonce = n.pendingNonceFromNode()
		log.Errorf("[NonceManager]: synced pendingNonce with network: %v", n.nextNonce.Int64())
		return
	}
	// maybe just reuse nonce
	n.nextNonce.Add(n.nextNonce, big.NewInt(-1))
	log.Errorf("[NonceManager]: decreasing nonce due to err: %v", err)
}

func (n *NonceManager) pendingNonceFromNode() *big.Int {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Duration(20*time.Second))
	defer cancel()
	//TODO(mmal): unify eth connection management
	no, e := n.ethClient.PendingNonceAt(ctx, n.ethAccount)
	if e != nil {
		log.Errorf("[NonceManager]: pendingNonceAt err %v", e)
	}
	return new(big.Int).SetUint64(no)
}

func (n *NonceManager) syncNonce() {
	n.nextNonce = n.pendingNonceFromNode()
	n.lastNonceChange = time.Now()
}

func (n *NonceManager) DebugPrint() {
	log.Printf("[NonceManager]: nonce stored: %v network: %v",
		n.nextNonce.Int64(), n.pendingNonceFromNode().Int64())
}

func (n *NonceManager) DebugDecrease() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.nextNonce.Add(n.nextNonce, big.NewInt(-1))
}

func (n *NonceManager) DebugForceIdle() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.lastNonceChange = time.Time{}
}

// false doesn't mean we are incorrect!
func (n *NonceManager) DebugNonceEqualsNetwork() bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.nextNonce.Cmp(n.pendingNonceFromNode()) == 0
}
