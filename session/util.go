package session

import (
	"net"
	"time"
)

// WaitAction wait some action done or several seconds
func WaitAction(act func(), timeout time.Duration) (success bool) {
	waitChan := make(chan struct{})
	go func() {
		act()
		close(waitChan)
	}()
	select {
	case <-waitChan:
		success = true
	case <-time.After(timeout):
		success = false
	}
	return
}

// IsNetTimeout check whether it is an timeout error
func IsNetTimeout(err error) bool {
	neterr, ok := err.(net.Error)
	return ok && neterr.Timeout()
}
