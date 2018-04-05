package spvwallet

import (
	"errors"
	"time"

	. "github.com/elastos/Elastos.ELA.SPV/common"
	"github.com/elastos/Elastos.ELA.SPV/p2p"
)

const (
	RequestTimeout = 15
	MaxRetryTimes  = 3
)

type RequestHandler interface {
	OnSendRequest(peer *p2p.Peer, reqType uint8, hash Uint256)
	OnRequestTimeout(Uint256)
}

type Request struct {
	peer       *p2p.Peer
	hash       Uint256
	reqType    uint8
	retryTimes int
	doneChan   chan byte
	handler    RequestHandler
}

func (r *Request) Start() error {
	if r.handler == nil {
		return errors.New("RequestHandler not set")
	}
	r.doneChan = make(chan byte)
	go r.sendRequest()
	return nil
}

func (r *Request) sendRequest() {
	r.handler.OnSendRequest(r.peer, r.reqType, r.hash)
	timer := time.NewTimer(time.Second * RequestTimeout)
	select {
	case <-timer.C:
		if r.retryTimes >= MaxRetryTimes {
			r.Finish()
			r.handler.OnRequestTimeout(r.hash)
			break
		}
		r.retryTimes++
		r.sendRequest()
	case <-r.doneChan:
		timer.Stop()
	}
}

func (r *Request) Finish() {
	if r.doneChan != nil {
		r.doneChan <- 1
		r.doneChan = nil
	}
}
