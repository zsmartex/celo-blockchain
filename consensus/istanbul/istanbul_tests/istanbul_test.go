package istanbul_tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/celo-org/celo-blockchain/consensus/istanbul"
	"github.com/celo-org/celo-blockchain/consensus/istanbul/core"
	"github.com/celo-org/celo-blockchain/test"
	"github.com/stretchr/testify/require"
)

// CapturingMessageSender is a sender implementation that cpatures messages
// sent on a channel.
type CapturingMessageSender struct {
	Msgs chan MsgAndDest
}

// MsgAndDest simply wraps messages and the target destinations so that they
// can be sent as one down a channel.
type MsgAndDest struct {
	Msg  *istanbul.Message
	Dest []common.Address
}

func NewCapturingMessageSender() *CapturingMessageSender {
	return &CapturingMessageSender{
		Msgs: make(chan MsgAndDest),
	}
}

func (s *CapturingMessageSender) Send(payload []byte, addresses []common.Address) error {
	msg := new(istanbul.Message)
	err := msg.FromPayload(payload, nil)
	if err != nil {
		panic(fmt.Sprintf("Failed to decode message after sending: %v", err))
	}
	mnd := MsgAndDest{
		Msg:  msg,
		Dest: addresses,
	}
	s.Msgs <- mnd
	return nil
}

type Timer interface {
	AfterFunc(time.Duration, func())
	Stop()
}

func NewTestTimer() *TestTimer {
	return &TestTimer{}
}

type TestTimer struct {
	F func()
}

func (t *TestTimer) AfterFunc(_ time.Duration, f func()) {
	t.F = f
}

func (t *TestTimer) Stop() {
	t.F = nil
}

func NewTestTimers() *core.Timers {
	return &core.Timers{
		RoundChange:       NewTestTimer(),
		ResendRoundChange: NewTestTimer(),
		FuturePreprepare:  NewTestTimer(),
	}
}

func TestCommit(t *testing.T) {
	accounts := test.Accounts(3)
	gc, ec, err := test.BuildConfig(accounts)
	require.NoError(t, err)
	genesis, err := test.ConfigureGenesis(accounts, gc, ec)
	require.NoError(t, err)

	sender := NewCapturingMessageSender()
	n, err := test.NewNode(&accounts.ValidatorAccounts()[0], &accounts.DeveloperAccounts()[0], test.BaseNodeConfig, ec, genesis, sender, core.NewDefaultTimers())
	require.NoError(t, err)
	defer n.Close()

	m := <-sender.Msgs
	println(m.Msg.Code)
}