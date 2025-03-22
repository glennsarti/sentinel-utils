package internal

// import (
// 	"github.com/creachadair/jrpc2"
// 	"github.com/creachadair/jrpc2/channel"
// 	"github.com/creachadair/jrpc2/server"
// )

// // singleServer is a wrapper around jrpc2.NewServer providing support
// // for server.Service (Assigner/Finish interface)
// type ServerWrapper struct {
// 	srv        *jrpc2.Server
// 	finishFunc func(jrpc2.ServerStatus)
// }

// func NewServerWrapper(svc server.Service, opts *jrpc2.ServerOptions) (*ServerWrapper, error) {
// 	assigner, err := svc.Assigner()
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &ServerWrapper{
// 		srv: jrpc2.NewServer(assigner, opts),
// 		finishFunc: func(status jrpc2.ServerStatus) {
// 			svc.Finish(assigner, status)
// 		},
// 	}, nil
// }

// func (ss *ServerWrapper) Start(ch channel.Channel) {
// 	ss.srv = ss.srv.Start(ch)
// }

// func (ss *ServerWrapper) StartAndWait(ch channel.Channel) {
// 	ss.Start(ch)
// 	ss.Wait()
// }

// func (ss *ServerWrapper) Wait() {
// 	status := ss.srv.WaitStatus()
// 	ss.finishFunc(status)
// }

// func (ss *ServerWrapper) Stop() {
// 	ss.srv.Stop()
// }
