package database

import (
	"context"
	"database/sql"
)

type DataBase struct {
	db   *sql.DB
	conn *sql.Conn
}

func NewDataBase(ctx context.Context, dbType, dbConnectionString string) (*DataBase, error) {
	db, err := sql.Open(dbType, dbConnectionString)
	if err != nil {
		return nil, err
	}
	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	return &DataBase{db: db, conn: conn}, nil
}
func (db *DataBase) GetConn(ctx context.Context) (*sql.Conn, error) {
	conn, err := db.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	db.conn = conn

	return db.conn, nil
}
func (db *DataBase) CloseConnection() {
	db.conn.Close()
}
func (db *DataBase) Close() {
	db.db.Close()
}
