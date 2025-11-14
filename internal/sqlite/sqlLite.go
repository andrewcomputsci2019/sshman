package sqlite

import (
	"fmt"
	"log/slog"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

type Connection struct {
	conn *sqlite.Conn
}

type executeObject struct {
	executeStr string
	args       []any
}

type executeObjectName struct {
	executeStr string
	args       []any
	names      map[string]any
}

func CreateAndLoadDB(path string) (*Connection, error) {
	conn, err := sqlite.OpenConn(path, sqlite.OpenReadWrite, sqlite.OpenCreate)
	if err != nil {
		return nil, err
	}
	sqlCon := &Connection{
		conn: conn,
	}
	err = sqlCon.createTable()
	if err != nil {
		return nil, err
	}
	return sqlCon, nil
}

func (conn *Connection) Close() {
	slog.Debug("Closing sqlite connection", "function", "Connection.Close")
	err := conn.conn.Close()
	if err == nil {
		return
	}
	slog.Error("Error closing sqlite connection", "function", "Connection.Close", "Error", err.Error())
}

func (conn *Connection) createTable() error {
	if conn == nil {
		return fmt.Errorf("sqlite connection is nil")
	}
	sqlCon := conn.conn
	createTableString := `
	CREATE TABLE IF NOT EXISTS hosts (
		host TEXT NOT NULL PRIMARY KEY,
		created_at INTEGER NOT NULL,
		updated_at INTEGER
	);
	
	CREATE TABLE IF NOT EXISTS host_config_options (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		host TEXT NOT NULL REFERENCES hosts(host) ON DELETE CASCADE,
		key TEXT NOT NULL,
		value TEXT NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_host_options_host
	ON host_config_options(host);
	`
	err := sqlitex.ExecScript(sqlCon, createTableString)
	if err != nil {
		return err
	}
	return nil
}

func (conn *Connection) query(query string, res func(stmt *sqlite.Stmt) error, args ...any) error {

	if conn == nil {
		return fmt.Errorf("sqlite connection is nil")
	}
	err := sqlitex.Execute(conn.conn, query, &sqlitex.ExecOptions{
		Args:       args,
		ResultFunc: res,
	})
	if err != nil {
		return err
	}
	return nil
}

func (conn *Connection) queryNamed(query string, res func(stmt *sqlite.Stmt) error, names map[string]any, args ...any) error {
	if conn == nil {
		return fmt.Errorf("sqlite connection is nil")
	}
	err := sqlitex.Execute(conn.conn, query, &sqlitex.ExecOptions{
		Args:       args,
		ResultFunc: res,
		Named:      names,
	})
	if err != nil {
		return err
	}
	return nil
}

func (conn *Connection) execute(insert string, args ...any) error {
	if conn == nil {
		return fmt.Errorf("sqlite connection is nil")
	}
	err := sqlitex.Execute(conn.conn, insert, &sqlitex.ExecOptions{
		Args: args,
	})
	if err != nil {
		return err
	}
	return nil
}

func (conn *Connection) executeNamed(insert string, names map[string]any, args ...any) error {
	if conn == nil {
		return fmt.Errorf("sqlite connection is nil")
	}
	err := sqlitex.Execute(conn.conn, insert, &sqlitex.ExecOptions{
		Args:  args,
		Named: names,
	})
	if err != nil {
		return err
	}
	return nil
}

func (conn *Connection) batchExecute(statements ...executeObject) error {
	dbCon := conn.conn
	endFn, err := sqlitex.ExclusiveTransaction(dbCon)
	if err != nil {
		return err
	}
	defer endFn(&err)

	for _, statement := range statements {
		err = conn.execute(statement.executeStr, statement.args...)
		if err != nil {
			return err
		}
	}
	return nil
}

func (conn *Connection) executeNamedBatch(statements ...executeObjectName) error {
	dbCon := conn.conn
	endFn, err := sqlitex.ExclusiveTransaction(dbCon)
	if err != nil {
		return err
	}
	defer endFn(&err)
	for _, statement := range statements {
		err = conn.executeNamed(statement.executeStr, statement.names, statement.args...)
		if err != nil {
			return err
		}
	}
	return nil
}
