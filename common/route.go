package common

import "time"

// IOutProtocol protocol message
type IOutProtocol interface {
	Marshal() ([]byte, error)
}

// BytesOutProtocol a raw bytes message
type BytesOutProtocol []byte

// Marshal get the encoded bytes
func (m BytesOutProtocol) Marshal() ([]byte, error) {
	return []byte(m), nil
}

// String implement the Stringer interface
func (m BytesOutProtocol) String() string {
	return string(m)
}

// IRequest client request
type IRequest interface {
	GetMID() uint8
	GetAID() uint8
	GetProtoVer() uint8
	GetData() []byte
	GetSign() []byte
}

// IAction action handler
type IAction interface {
	GetAID() uint8
	Handle(IRequest) IOutProtocol
}

type noneAction struct{}

func (*noneAction) GetAID() uint8 { return 0 }
func (*noneAction) Handle(_ IRequest) IOutProtocol {
	return BytesOutProtocol(nil)
}

// NoneAction an action doing nothing
var NoneAction IAction = &noneAction{}

// IModule module handler
type IModule interface {
	GetMID() uint8
	Handle(IRequest) IOutProtocol
}

type noneModule struct{}

func (*noneModule) GetMID() uint8 { return 0 }
func (*noneModule) Handle(_ IRequest) IOutProtocol {
	return BytesOutProtocol(nil)
}

// NoneModule a module doing nothing
var NoneModule IModule = &noneModule{}

type baseModule struct {
	mid     uint8
	actions map[uint8]IAction
}

// NewModule create a IModule instance
func NewModule(mid uint8, acts ...IAction) IModule {
	modActions := make(map[uint8]IAction, len(acts))
	for _, v := range acts {
		modActions[v.GetAID()] = v
	}
	return &baseModule{mid: mid, actions: modActions}
}

func (m *baseModule) GetMID() uint8 {
	return m.mid
}

func (m *baseModule) Handle(r IRequest) IOutProtocol {
	actionID := r.GetAID()
	act, ok := m.actions[actionID]
	if !ok {
		act = NoneAction
		// TODO: log
		//zaplog.S.Errorf("module %d: action(%d) not found", m.mid, actionID)
	}
	return act.Handle(r)
}

var _ IModule = (*baseModule)(nil)

// IRouteEnabler enable or disable some routes
type IRouteEnabler interface {
	Enabled(uint8, uint8) bool
}

type fullRouteEnabler struct{}

func (*fullRouteEnabler) Enabled(_ uint8, _ uint8) bool { return true }

// FullRouteEnabler a fullRouteEnabler struct
var FullRouteEnabler IRouteEnabler = &fullRouteEnabler{}

// ITimeouter wait a while and return a timeout proto.Message
type ITimeouter interface {
	Timeout() time.Duration
	Result() IOutProtocol
}

// Router a module router
type Router struct {
	modules  map[uint8]IModule
	enabler  IRouteEnabler
	timeout  ITimeouter
	noneResp IOutProtocol
}

// RouterOptionFunc set the Router's option
type RouterOptionFunc func(*Router)

// OptionRouteEnabler set Router's enabler
func OptionRouteEnabler(enabler IRouteEnabler) RouterOptionFunc {
	return func(r *Router) {
		r.enabler = enabler
	}
}

// OptionTimeoutResponse set Router's timeout
func OptionTimeoutResponse(timeout ITimeouter) RouterOptionFunc {
	return func(r *Router) {
		r.timeout = timeout
	}
}

// OptionNoneResponse set Router's noneResp
func OptionNoneResponse(resp IOutProtocol) RouterOptionFunc {
	return func(r *Router) {
		r.noneResp = resp
	}
}

// NewRouter create a Router struct
func NewRouter(opts ...RouterOptionFunc) *Router {
	router := &Router{map[uint8]IModule{}, FullRouteEnabler, nil, nil}
	for _, opt := range opts {
		opt(router)
	}
	return router
}

// Register register several modules
func (router *Router) Register(modules ...IModule) {
	for _, m := range modules {
		router.modules[m.GetMID()] = m
	}
}

// Dispatch dispath each client's request
func (router *Router) Dispatch(r IRequest) (IOutProtocol, bool) {
	moduleID := r.GetMID()
	actionID := r.GetAID()

	var module IModule
	if router.enabler.Enabled(moduleID, actionID) {
		var ok bool
		module, ok = router.modules[moduleID]
		if !ok {
			//TODO: log
			//zaplog.S.Errorf("router: module(%d) not found", moduleID)
			return router.noneResp, false
		}
	} else {
		//TODO: log
		//zaplog.S.Errorf("router: module(%d) action(%d) disabled", moduleID, actionID)
		return router.noneResp, false
	}
	if router.timeout == nil {
		return module.Handle(r), false
	}
	// timeout to handle a request
	result := make(chan IOutProtocol, 1)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				//TODO : log
			}
		}()
		result <- module.Handle(r)
	}()
	select {
	case pb := <-result:
		return pb, false
	case <-time.After(router.timeout.Timeout()):
		return router.timeout.Result(), true
	}
}
