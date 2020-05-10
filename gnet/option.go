package gnet

type Option func(svr *QNetServer) error

func WithNetMsgCodec(codec INetMsgCodec) Option {
	return func(svr *QNetServer) error {
		svr.netMsgCodec = codec
		svr.react = codec.React
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
		return nil
	}
}
