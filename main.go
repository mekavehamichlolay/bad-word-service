package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os/signal"
	"sync"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/mekavehamichlolay/bad-word-service/database"
	"github.com/mekavehamichlolay/bad-word-service/loger"
	"github.com/mekavehamichlolay/bad-word-service/maptree"
	"github.com/mekavehamichlolay/bad-word-service/server"
)

func main() {
	config := server.Configure()
	if config == nil {
		return
	}

	var wg = new(sync.WaitGroup)

	log := loger.NewLoger("bad-word-service.log", wg)
	defer log.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	db, err := database.NewDataBase(ctx, config.DBType, config.DBConnectionString)
	if err != nil {
		log.Err(fmt.Sprintf("Failed to create the database connection: %v", err))
		return
	}
	defer db.Close()

	conn, err := db.GetConn(ctx)
	if err != nil {
		log.Err(fmt.Sprintf("Failed to get connection to database: %v", err))
		return
	}
	defer db.CloseConnection()

	tree := maptree.NewTree()
	if err := tree.Reset(ctx, conn); err != nil {
		log.Err(fmt.Sprintf("Failed to reset the tree: %v", err))
		return
	}
	mainRoute := server.CreateRoute(
		config.SocketPath,
		"main socket for the bad word service",
		func(c net.Conn) {
			defer c.Close()
			if err := c.SetDeadline(time.Now().Add(1 * time.Second)); err != nil {
				log.Err(fmt.Sprintf("Failed to set deadline: %v", err))
				return
			}
			var buffer = make([]byte, 512)
			var text string
			for {
				lengthe, err := c.Read(buffer)
				if err != nil {
					if err == io.EOF {
						text += string(buffer[:lengthe])
						break
					}
					log.Err(fmt.Sprintf("Failed to read from the connection: %v", err))
					return
				}
				if lengthe == 0 {
					break
				}
				text += string(buffer[:lengthe])
			}
			positions := tree.HasWord(text)
			jsoned, err := json.Marshal(positions)
			if err != nil {
				log.Err(fmt.Sprintf("Failed to marshal the positions: %v", err))
				return
			}
			c.Write(jsoned)
		})
	resetRoute := server.CreateRoute(
		config.SocketPath+"reset",
		"reset socket for the bad word service",
		func(c net.Conn) {
			defer c.Close()
			if err := tree.Reset(ctx, conn); err != nil {
				log.Err(fmt.Sprintf("Failed to reset the tree: %v", err))
				return
			}
		})
	killRoute := server.CreateRoute(
		config.SocketPath+"kill",
		"kill socket for the bad word service",
		func(c net.Conn) {
			c.Close()
			cancel()
		})
	allWordsRoute := server.CreateRoute(
		config.SocketPath+"all",
		"all words socket for the bad word service",
		func(c net.Conn) { //TODO: implement this route
			defer c.Close()
		})

	routes := []*server.Route{
		mainRoute, resetRoute, killRoute, allWordsRoute,
	}

	if err := server.StartServer(ctx, wg, routes, log); err != nil {
		log.Err(fmt.Sprintf("Failed to start the server: %v", err))
		cancel()
	}
	wg.Wait()
	log.Info("Server stopped")
	wg.Wait()
}
