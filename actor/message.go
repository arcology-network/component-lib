package actor

import (
	"encoding/binary"
	"encoding/gob"

	ethCommon "github.com/HPISTechnologies/3rd-party/eth/common"
	"github.com/HPISTechnologies/common-lib/common"
)

const (
	MsgHeight                   = "height"
	MsgInclusive                = "inclusive"
	MsgConflictInclusive        = "conflictInclusive"
	MsgBlockCompleted           = "blockCompleted"
	MsgClearReceipts            = "clearReceipts"
	MsgInclusiveReceipts        = "inclusiveReceipts"
	MsgClearEuresults           = "clearEuresults"
	MsgInclusiveEuresults       = "inclusiveEuresults"
	MsgCoinbase                 = "coinbase"
	MsgTxHash                   = "txhash"
	MsgRcptHash                 = "rcpthash"
	MsgAcctHash                 = "accthash"
	MsgGasUsed                  = "gasused"
	MsgParentInfo               = "parentinfo"
	MsgInitParentInfo           = "initParentinfo"
	MsgNewParentInfo            = "newparentinfo"
	MsgTxLocal                  = "txLocal"
	MsgCheckedTxs               = "checkedTxs"
	MsgGroupedTxs               = "groupedTxs"
	MsgTxBlocks                 = "txBlocks"
	MsgTxLocals                 = "txLocals"
	MsgTxRemotes                = "txRemotes"
	MsgStartSub                 = "startSub"
	MsgStarting                 = "starting"
	MsgMetaBlock                = "metablock"
	MsgSelectedTx               = "selectedtx"
	MsgPendingBlock             = "pendingblock"
	MsgExecTime                 = "execTime"
	MsgReceiptHashList          = "receiptHashList"
	MsgEuResults                = "euResults"
	MsgTxAccessRecords          = "txAccessRecords"
	MsgExecutingLogs            = "executingLogs"
	MsgPreProcessedEuResults    = "preProcessedEuResults"
	MsgApcHandle                = "apchandle"
	MsgReapinglist              = "reapinglist"
	MsgReapingCompleted         = "reapingCompleted"
	MsgExecuted                 = "executed"
	MsgInitExecuted             = "initExecuted"
	MsgStatesUpdates            = "statesUpdates"
	MsgExecutorParameter        = "executorParameter"
	MsgApcBlock                 = "apcBlock"
	MsgSelectCompletedReceipts  = "selectCompletedReceipts"
	MsgClearCompletedReceipts   = "clearCompletedReceipts"
	MsgSelectCompletedEuresults = "selectCompletedEuresults"
	MsgClearCompletedEuresults  = "clearCompletedEuresults"
	MsgSelectedReceipts         = "selectedReceipts"
	MsgReceipts                 = "receipts"
	MsgCheckingTxs              = "checkingtxs"
	MsgMessager                 = "messager"
	MsgBlockStamp               = "blockStamp"
	//MsgInitStatesUpdates        = "initStatesUpdates"
	MsgLatestHeight           = "latestHeight"
	MsgInitApcHandle          = "initapchandle"
	MsgBlockCompleted_Success = "success"
	MsgBlockCompleted_Failed  = "failed"
	MsgMessagersReaped        = "messagersReaped"
	MsgArbitrateList          = "arbitrateList"
	MsgTxsToExecute           = "txsToExecute"
	MsgTxsExecuteResults      = "txsExecuteResults"
	MsgEuResultSelected       = "euResultSelected"
	MsgTxs                    = "txs"
	MsgPrecedingList          = "precedingList"
	MsgPrecedingsEuresult     = "precedingsEuresult"
	MsgQuery                  = "query"
	MsgReapCommand            = "reapCommand"
	MsgInitReapCommand        = "initReapCommand"
	MsgStartReapCommand       = "startReapCommand"
	MsgAppHash                = "appHash"
	MsgSpawnedRelations       = "spawnedRelations"
	MsgSavedRoothash          = "savedRoothash"
	MsgSavedStateUpdates      = "savedStateUpdates"
	MsgClearCompleted         = "clearCompleted"
	MsgClearCommand           = "clearCommand"

	MsgVertifyHeader      = "blockHeader"
	MsgNodeRole           = "nodeRole"
	MsgBlockRole_Propose  = "propose"
	MsgBlockRole_Validate = "validate"
	MsgProposeBlock       = "proposeblock"
	MsgValidateBlock      = "validateblock"
	MsgBlockVertified     = "blockvertified"
)

type Comparable interface {
	Equals(rhs Comparable) bool
}

const (
	MessageFrom_Local  = byte(0)
	MessageFrom_Remote = byte(1)
)

func init() {
	gob.Register(&Message{})
}

type Message struct {
	From        byte
	Msgid       uint64
	Name        string
	Height      uint64
	Round       uint64
	Data        interface{}
	encodedSize uint32
}

func NewMessage() *Message {
	return &Message{
		Msgid:  0,
		Name:   "",
		Height: 0,
		Round:  0,
		Data:   nil,
	}
}

func (m *Message) Equals(rhs Comparable) bool {
	other := rhs.(*Message)
	return m.Name == other.Name && m.Height == other.Height && m.Round == other.Round && m.Data == other.Data
}

func (m *Message) CopyHeader() *Message {
	return &Message{
		From:        m.From,
		Msgid:       m.Msgid,
		Name:        m.Name,
		Height:      m.Height,
		Round:       m.Round,
		encodedSize: m.encodedSize,
	}
}

func (msg *Message) Encode() ([]byte, error) {
	data, err := common.GobEncode(&msg)
	msg.encodedSize = uint32(len(data))
	return data, err
}

func (msg *Message) Decode(data []byte) error {
	msg.encodedSize = uint32(len(data))
	return common.GobDecode(data, &msg)
}

func (msg *Message) Size() uint32 {
	return msg.encodedSize
}

func (msg *Message) GetHeader() (uint64, uint64, uint64) {
	return msg.Height, msg.Round, msg.Msgid
}

func (msg *Message) Hash() ethCommon.Hash {
	hash := ethCommon.Hash{}
	binary.LittleEndian.PutUint64(hash[:], uint64(msg.Msgid))
	return hash
}
