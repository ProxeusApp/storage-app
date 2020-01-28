package ethereum

import (
	"bytes"
	"math/big"
	"strings"

	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/ProxeusApp/storage-app/spp/eth"
)

type (
	xesClient struct {
		baseClient *baseClient
		xesAddress common.Address
		xesABI     abi.ABI

		xesTokenContractCaller *eth.XESTokenContractCaller
	}
)

const (
	EventXesReceive = "eventXesReceive"
	EventXesSend    = "eventXesSend"
)

func NewXesClient(baseClient *baseClient, xesAddress common.Address, xesABI abi.ABI) *xesClient {
	xesClient := new(xesClient)
	xesClient.baseClient = baseClient
	xesClient.xesAddress = xesAddress
	xesClient.xesABI = xesABI
	return xesClient
}

func (me *xesClient) balanceXESof(ethAddress string) (*big.Int, error) {
	from := common.HexToAddress(ethAddress)
	auth := &bind.TransactOpts{
		From: from,
	}
	ctx, cancel := me.baseClient.ctxWithTimeout()
	defer cancel()
	ops := &bind.CallOpts{Pending: false, From: auth.From, Context: ctx}
	return me.xesTokenContractCaller.BalanceOf(ops, from)
}

func (me *xesClient) proxeusFSAllowance(ethAddr string, pfsAddress common.Address) (*big.Int, error) {
	ctx, cancel := me.baseClient.ctxWithTimeout()
	opts := &bind.CallOpts{Pending: false, From: common.HexToAddress(ethAddr), Context: ctx}
	b, err := me.xesTokenContractCaller.Allowance(opts, opts.From, pfsAddress)
	cancel()
	return b, err
}

func (me *xesClient) logAsXesTransfer(lg *types.Log, recent bool) (*eth.XESTokenContractTransfer, error) {
	const transfer = "Transfer"
	var eventKey string
	var ok bool
	if eventKey, ok = me.isNotXesEventOrAlreadyExecuted(transfer, lg); ok {
		return nil, nil
	}

	event := new(eth.XESTokenContractTransfer)
	if err := me.eventFromLog(event, lg, transfer); err != nil {
		return nil, err
	}

	me.baseClient.alreadyExecutedSuccessfully(eventKey)
	if !me.xesEventInterestingForMe(event) {
		return nil, nil
	}
	event.Raw = *lg
	return event, nil
}

func (me *xesClient) eventFromLog(out interface{}, lg *types.Log, eventType string) error {
	pfsLogUnpacker := bind.NewBoundContract(me.xesAddress, me.xesABI, me.baseClient.ethwsconn, me.baseClient.ethwsconn,
		me.baseClient.ethwsconn)
	err := pfsLogUnpacker.UnpackLog(out, eventType, *lg)
	if err != nil {
		return err // not our event type
	}
	return nil
}

func (me *xesClient) xesEventInterestingForMe(event *eth.XESTokenContractTransfer) bool {
	current := me.baseClient.currentEthAddress()
	return current == event.FromAddress || current == event.ToAddress
}

func (me *xesClient) isNotXesEventOrAlreadyExecuted(eventName string, lg *types.Log) (eventKey string, doNotProceed bool) {
	if !me.isXesEvent(eventName, lg) {
		return eventKey, true
	}
	eventKey, ok := me.baseClient.alreadyExecutedRecently(lg)
	if ok {
		return eventKey, true
	}
	return eventKey, false
}

func (me *xesClient) isXesEvent(eventName string, lg *types.Log) bool {
	if len(lg.Topics) < 0 || !bytes.Equal(me.xesABI.Events[eventName].Id().Bytes(), lg.Topics[0].Bytes()) {
		return false
	}
	return true
}

var ErrEventXesTransferNotInteresting = errors.New("Either from or to address must be current address.")

func (me *xesClient) handleXesTransferEvent(from, to common.Address, txHash common.Hash, xesAmount *big.Int) error {

	evd := &PendingTx{}
	if from == me.baseClient.currentEthAddress() {
		evd.Type = EventXesSend
		evd.Who = []string{to.String()}
	} else if to == me.baseClient.currentEthAddress() {
		evd.Type = EventXesReceive
		evd.Who = []string{from.String()}
	} else {
		return ErrEventXesTransferNotInteresting
	}

	evd.CurrentAddress = me.baseClient.currentAddress
	evd.TxHash = strings.ToLower(txHash.Hex())
	evd.XesAmount = xesAmount

	return me.baseClient.notify(evd, evd.TxHash, StatusSuccess)
}
