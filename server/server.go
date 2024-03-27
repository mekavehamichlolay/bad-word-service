package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/mekavehamichlolay/bad-word-service/loger"
)

type route struct {
	path        string
	description string
	handler     func(net.Conn)
}
type Route route

func (r *Route) Start(ctx context.Context, wg *sync.WaitGroup, loger loger.Loger) error {
	ctx, cancel := context.WithCancel(ctx)
	socket, err := net.Listen("unix", r.path)
	if err != nil {
		cancel()
		return err
	}
	defer socket.Close()
	loger.Info(fmt.Sprintf("listening on %s...", r.path))
	wg.Add(1)
	go func() {
		<-ctx.Done()
		if err := socket.Close(); err != nil {
			loger.Err(fmt.Sprintf("error closing the socket: %s", err.Error()))
		}
		if err := os.Remove(socket.Addr().String()); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				loger.Err(fmt.Sprintf("error removing the socket file: %s", err.Error()))
			}
		}
		wg.Done()
	}()
	errorsN := 0
	for {
		conn, err := socket.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				break
			}
			loger.Err(fmt.Sprintf("error accepting a connection: %s", err.Error()))
			errorsN++
			if errorsN >= 5 {
				break
			}
			continue
		}
		go r.handler(conn)
	}
	cancel()
	return nil
}

func CreateRoute(path, description string, handler func(net.Conn)) *Route {
	return &Route{
		path:        path,
		description: description,
		handler:     handler,
	}
}

func StartServer(ctx context.Context, wg *sync.WaitGroup, routes []*Route, loger loger.Loger) error {
	for _, r := range routes {
		wg.Add(1)
		go func(r *Route) error {
			defer wg.Done()
			if err := r.Start(ctx, wg, loger); err != nil {
				loger.Err(fmt.Sprintf("error starting the server: %s", err.Error()))
				return err
			}
			return nil
		}(r)
	}
	return nil
}
