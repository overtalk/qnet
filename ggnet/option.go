package ggnet

import (
	"errors"
	"time"

	"github.com/panjf2000/gnet"

	"github.com/overtalk/qnet"
)

type Option func(svr *QNetServer) error

func WithURL(url string) Option {
	return func(svr *QNetServer) error {
		if svr.ep != nil {
			return errors.New("repetitive server endpoint")
		}

		ep, err := qnet.NewFromString(url)
		if err != nil {
			return err
		}

		return WithEndPoint(ep)(svr)
	}
}

func WithEndPoint(ep *qnet.Endpoint) Option {
	return func(svr *QNetServer) error {
		if svr.ep != nil {
			return errors.New("repetitive server endpoint")
		}

		svr.protocolType = ep.Proto()
		svr.ep = ep
		return nil
	}
}

func WithOnInitComplete(f OnInitCompleteFunc) Option {
	return func(svr *QNetServer) error {
		svr.onInitComplete = f
		return nil
	}
}

func WithOnShutdown(f OnShutdownFunc) Option {
	return func(svr *QNetServer) error {
		svr.onShutdown = f
		return nil
	}
}

func WithOnOpened(f OnOpenedFunc) Option {
	return func(svr *QNetServer) error {
		svr.onOpened = f
		return nil
	}
}

func WithOnClosed(f OnClosedFunc) Option {
	return func(svr *QNetServer) error {
		svr.onClosed = f
		return nil
	}
}

func WithPreWrite(f PreWriteFunc) Option {
	return func(svr *QNetServer) error {
		svr.preWrite = f
		return nil
	}
}

func WithReact(f ReactFunc) Option {
	return func(svr *QNetServer) error {
		svr.react = f
		return nil
	}
}

func WithTick(f TickFunc) Option {
	return func(svr *QNetServer) error {
		svr.tick = f
		svr.options = append(svr.options, gnet.WithTicker(true))
		return nil
	}
}

// ------------------------------------------------------
func WithNetMsgCodec(codec INetMsgCodec) Option {
	return func(svr *QNetServer) error {
		svr.netMsgCodec = codec
		svr.react = codec.React
		return nil
	}
}

func WithCodec(codec gnet.ICodec) Option {
	return func(svr *QNetServer) error {
		svr.options = append(svr.options, gnet.WithCodec(codec))
		return nil
	}
}

func WithLoadBalancing(lb gnet.LoadBalancing) Option {
	return func(svr *QNetServer) error {
		svr.options = append(svr.options, gnet.WithLoadBalancing(lb))
		return nil
	}
}

func WithMulticore(multicore bool) Option {
	return func(svr *QNetServer) error {
		svr.options = append(svr.options, gnet.WithMulticore(multicore))
		return nil
	}
}

func WithNumEventLoop(numEventLoop int) Option {
	return func(svr *QNetServer) error {
		svr.options = append(svr.options, gnet.WithNumEventLoop(numEventLoop))
		return nil
	}
}

func WithTCPKeepAlive(t time.Duration) Option {
	return func(svr *QNetServer) error {
		svr.options = append(svr.options, gnet.WithTCPKeepAlive(t))
		return nil
	}
}

func WithReusePort(reusePort bool) Option {
	return func(svr *QNetServer) error {
		svr.options = append(svr.options, gnet.WithReusePort(reusePort))
		return nil
	}
}

func WithLogger(logger gnet.Logger) Option {
	return func(svr *QNetServer) error {
		svr.options = append(svr.options, gnet.WithLogger(logger))
		return nil
	}
}
