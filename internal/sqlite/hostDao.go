package sqlite

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"zombiezen.com/go/sqlite"
)

type Host struct {
	Host           string
	CreatedAt      time.Time
	UpdatedAt      *time.Time // may be nil
	LastConnection *time.Time
	Notes          string
	Options        []HostOptions
	Tags           []string
}

func (h *Host) String() string {
	builder := strings.Builder{}
	builder.WriteString("{\n")
	builder.WriteString("Host: ")
	builder.WriteString(h.Host + ",\n")
	builder.WriteString("CreatedAt: ")
	builder.WriteString(strconv.FormatInt(h.CreatedAt.UnixMilli(), 10))
	builder.WriteString(",\n")
	builder.WriteString("UpdatedAt: ")
	if h.UpdatedAt != nil {
		builder.WriteString(strconv.FormatInt(h.UpdatedAt.UnixMilli(), 10))
		builder.WriteString(",\n")
	} else {
		builder.WriteString("<nil>,\n")
	}
	builder.WriteString("LastConnection: ")
	if h.LastConnection != nil {
		builder.WriteString(strconv.FormatInt(h.LastConnection.UnixMilli(), 10))
		builder.WriteString(",\n")
	} else {
		builder.WriteString("<nil>,\n")
	}
	builder.WriteString("Notes: ")
	if h.Notes != "" {
		builder.WriteString(h.Notes)
	} else {
		builder.WriteString("<nil>,\n")
	}
	builder.WriteString(",\n")
	builder.WriteString("Options: [")
	for i, opt := range h.Options {
		if i < len(h.Options)-1 {
			builder.WriteString(",")
		}
		builder.WriteString(opt.String())
	}
	builder.WriteString("],\n")
	return builder.String()
}

type HostOptions struct {
	ID    int64
	Key   string
	Value string
	Host  string
}

func (h *HostOptions) String() string {
	return fmt.Sprintf("{ID: %d, Key: %s, Value: %s}", h.ID, h.Key, h.Value)
}

type HostDao struct {
	conn *Connection
}

const (
	hostInsertString    = `INSERT INTO host (host,created_at,updated_at,last_connection, notes, tags) VALUES (?,?,?,?,?,?)`
	hostOptInsertString = `INSERT INTO host_options (host, key, value) VALUES (?,?,?)`
	hostUpdateString    = `UPDATE host SET (created_at, updated_at, last_connection, notes, tags) VALUES (?,?,?,?,?) WHERE host=?`
	hostOptUpdateString = `INSERT OR IGNORE INTO host_options (host, key, value) VALUES (?, ?, ?);`
	hostUpSert          = `INSERT INTO host (host, created_at, updated_at, last_connection, notes, tags) VALUES (?, ?, ?, ?, ?, ?) 
ON CONFLICT(host) DO UPDATE SET updated_at=MAX(host.updated_at,excluded.updated_at), last_connection=MAX(host.last_connection, excluded.last_connection), notes=excluded.notes,tags=excluded.tags;`
	hostDeleteString = `DELETE FROM host WHERE host=?`
)

func NewHostDao(conn *Connection) *HostDao {
	if conn == nil {
		return nil
	}
	return &HostDao{conn: conn}
}

func ts(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.UnixMilli()
}

func (dao *HostDao) Insert(host Host) error {
	insertHostString := hostInsertString
	insertOptString := hostOptInsertString
	err := dao.conn.transaction(func() error {
		tagsJoined := strings.Join(host.Tags, ",")
		err := dao.conn.execute(insertHostString, host.Host, ts(&host.CreatedAt), ts(host.UpdatedAt), ts(host.LastConnection), host.Notes, tagsJoined)
		if err != nil {
			return err
		}
		for _, opt := range host.Options {
			err = dao.conn.execute(insertOptString, host.Host, opt.Key, opt.Value)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (dao *HostDao) Update(host Host) error {
	updateString := hostUpdateString
	updateOptString := hostOptUpdateString
	deleteOptString, args := generateDeleteStringOpts(&host)
	err := dao.conn.transaction(func() error {
		tagsJoined := strings.Join(host.Tags, ",")
		err := dao.conn.execute(updateString, ts(&host.CreatedAt), ts(host.UpdatedAt), ts(host.LastConnection), host.Notes, tagsJoined, host.Host)
		if err != nil {
			return err
		}
		for _, opt := range host.Options {
			err = dao.conn.execute(updateOptString, host.Host, opt.Key, opt.Value)
			if err != nil {
				return err
			}
		}
		err = dao.conn.execute(deleteOptString, args...)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func generateDeleteStringOpts(host *Host) (string, []any) {
	// some fields like identity_file, localForward, SendEnv, CertificateFile can multiple valid defs,
	// so we shouldn't remove those if they exist in the dao, this method basically removes any opt entry not in the dao from the database
	if len(host.Options) == 0 { // if host options are empty just delete all opts from table
		return `DELETE FROM host_options WHERE host = ?`, []any{host.Host}
	}
	deleteBuilder := strings.Builder{}
	deleteBuilder.WriteString(`WITH new_values(host, key, value) AS (VALUES `)
	args := make([]any, 0, len(host.Options)*3+1)

	for i, opt := range host.Options {
		if i < len(host.Options)-1 {
			deleteBuilder.WriteString(", ")
		}
		deleteBuilder.WriteString("(?, ?, ?)")
		args = append(args, host.Host, opt.Key, opt.Value)
	}
	deleteBuilder.WriteString(")\n")
	deleteBuilder.WriteString(`
DELETE FROM host_options ho
WHERE ho.host = ?
  AND NOT EXISTS (
      SELECT 1 FROM new_values nv
      WHERE nv.host  = ho.host
        AND nv.key   = ho.key
        AND nv.value = ho.value
  );
`)
	args = append(args, host.Host)
	return deleteBuilder.String(), args
}

func (dao *HostDao) InsertMany(hosts ...Host) error {
	if hosts == nil || len(hosts) <= 0 {
		return fmt.Errorf("hosts is empty")
	}
	hostInsertString := hostInsertString
	hostOptInsertString := hostOptInsertString
	err := dao.conn.transaction(func() error {
		for _, host := range hosts {
			tagsJoined := strings.Join(host.Tags, ",")
			err := dao.conn.execute(hostInsertString, host.Host, ts(&host.CreatedAt), ts(host.UpdatedAt), ts(host.LastConnection), host.Notes, tagsJoined)
			if err != nil {
				return err
			}
			for _, opt := range host.Options {
				err = dao.conn.execute(hostOptInsertString, host.Host, opt.Key, opt.Value)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (dao *HostDao) UpdateMany(hosts ...Host) error {
	if hosts == nil || len(hosts) <= 0 {
		return fmt.Errorf("hosts is empty")
	}
	err := dao.conn.transaction(func() error {
		for _, host := range hosts {
			hostUpdateString := hostUpdateString
			hostOptUpdateString := hostOptUpdateString
			deleteOptString, args := generateDeleteStringOpts(&host)
			tagsJoined := strings.Join(host.Tags, ",")
			err := dao.conn.execute(hostUpdateString, ts(&host.CreatedAt), ts(host.UpdatedAt), ts(host.LastConnection), host.Notes, tagsJoined, host.Host)
			if err != nil {
				return err
			}
			for _, opt := range host.Options {
				err = dao.conn.execute(hostOptUpdateString, host.Host, opt.Key, opt.Value)
				if err != nil {
					return err
				}
			}
			err = dao.conn.execute(deleteOptString, args...)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (dao *HostDao) InsertOrUpdate(host Host) error {
	err := dao.conn.transaction(func() error {
		hostUpSertString := hostUpSert
		hostOptUpdate := hostOptUpdateString
		deleteOptString, args := generateDeleteStringOpts(&host)
		tagsJoined := strings.Join(host.Tags, ",")
		err := dao.conn.execute(hostUpSertString, host.Host, ts(&host.CreatedAt), ts(host.UpdatedAt), ts(host.LastConnection), host.Notes, tagsJoined)
		if err != nil {
			return err
		}
		for _, opt := range host.Options {
			err = dao.conn.execute(hostOptUpdate, host.Host, opt.Key, opt.Value)
			if err != nil {
				return err
			}
		}
		err = dao.conn.execute(deleteOptString, args...)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (dao *HostDao) InsertOrUpdateMany(hosts ...Host) error {
	err := dao.conn.transaction(func() error {
		hostUpSertString := hostUpSert
		hostOptUpdate := hostOptUpdateString
		for _, host := range hosts {
			deleteOptString, args := generateDeleteStringOpts(&host)
			tagsJoined := strings.Join(host.Tags, ",")
			err := dao.conn.execute(hostUpSertString, host.Host, ts(&host.CreatedAt), ts(host.UpdatedAt), ts(host.LastConnection), host.Notes, tagsJoined)
			if err != nil {
				return err
			}
			for _, opt := range host.Options {
				err = dao.conn.execute(hostOptUpdate, host.Host, opt.Key, opt.Value)
				if err != nil {
					return err
				}
			}
			err = dao.conn.execute(deleteOptString, args...)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (dao *HostDao) Delete(host Host) error {
	deleteString := hostDeleteString
	err := dao.conn.execute(deleteString, host.Host)
	if err != nil {
		return err
	}
	return nil
}

func (dao *HostDao) serializeHostFromStatement(stmt *sqlite.Stmt, host *Host) error {
	host.Host = stmt.GetText("host")
	host.CreatedAt = time.UnixMilli(stmt.GetInt64("created_at"))
	updateAtIdx := stmt.ColumnIndex("updated_at")
	lastConnectionIdx := stmt.ColumnIndex("last_connection")
	if updateAtIdx < 0 || lastConnectionIdx < 0 {
		return fmt.Errorf("update at or last connection index out of range")
	}
	if stmt.ColumnType(updateAtIdx) != sqlite.TypeNull {
		host.UpdatedAt = new(time.Time)
		*host.UpdatedAt = time.UnixMilli(stmt.ColumnInt64(updateAtIdx))
	}
	if stmt.ColumnType(lastConnectionIdx) != sqlite.TypeNull {
		host.LastConnection = new(time.Time)
		*host.LastConnection = time.UnixMilli(stmt.ColumnInt64(lastConnectionIdx))
	}
	host.Notes = stmt.GetText("notes")
	host.Tags = strings.Split(stmt.GetText("tags"), ",")
	return nil
}

func (dao *HostDao) serializeHostOptionFromStatement(stmt *sqlite.Stmt, host *Host) error {
	opt := HostOptions{}
	opt.ID = stmt.ColumnInt64(0)
	opt.Key = stmt.GetText("key")
	opt.Value = stmt.GetText("value")
	opt.Host = host.Host
	host.Options = append(host.Options, opt)
	return nil
}

func (dao *HostDao) Get(host string) (Host, error) {
	hostItem := Host{}
	onResHostQuery := func(stmt *sqlite.Stmt) error {
		//hostItem.Host = stmt.GetText("host")
		//hostItem.CreatedAt = time.UnixMilli(stmt.GetInt64("created_at"))
		//updateAtIdx := stmt.ColumnIndex("updated_at")
		//lastConnectionIdx := stmt.ColumnIndex("last_connection")
		//if updateAtIdx < 0 || lastConnectionIdx < 0 {
		//	return fmt.Errorf("update at or last connection index out of range")
		//}
		//if stmt.ColumnType(updateAtIdx) != sqlite.TypeNull {
		//	hostItem.UpdatedAt = new(time.Time)
		//	*hostItem.UpdatedAt = time.UnixMilli(stmt.ColumnInt64(updateAtIdx))
		//}
		//if stmt.ColumnType(lastConnectionIdx) != sqlite.TypeNull {
		//	hostItem.LastConnection = new(time.Time)
		//	*hostItem.LastConnection = time.UnixMilli(stmt.ColumnInt64(lastConnectionIdx))
		//}
		err := dao.serializeHostFromStatement(stmt, &hostItem)
		if err != nil {
			return err
		}
		return nil
	}
	onResOptQuery := func(stmt *sqlite.Stmt) error {
		//opt := HostOptions{}
		//opt.ID = stmt.ColumnInt64(0)
		//opt.Key = stmt.GetText("key")
		//opt.Value = stmt.GetText("value")
		//opt.Host = hostItem.Host
		//hostItem.Options = append(hostItem.Options, opt)
		err := dao.serializeHostOptionFromStatement(stmt, &hostItem)
		if err != nil {
			return err
		}
		return nil
	}
	err := dao.conn.query(`SELECT * FROM host where host = ?`, onResHostQuery, host)
	if err != nil {
		return Host{}, err
	}
	err = dao.conn.query(`SELECT * FROM host_options where host = ?`, onResOptQuery, host)
	return hostItem, nil
}

func (dao *HostDao) GetN(n uint, offset uint) ([]Host, error) {
	queryString := `SELECT * FROM host LIMIT ? OFFSET ?`
	queryOptString := `SELECT * FROM host_options where host = ?`
	hosts := make([]*Host, 0)
	err := dao.conn.query(queryString, func(stmt *sqlite.Stmt) error {
		hostItem := &Host{}
		err := dao.serializeHostFromStatement(stmt, hostItem)
		if err != nil {
			return err
		}
		hosts = append(hosts, hostItem)
		return nil
	}, n, offset)
	if err != nil {
		return nil, err
	}
	for _, host := range hosts {
		err = dao.conn.query(queryOptString, func(stmt *sqlite.Stmt) error {
			err := dao.serializeHostOptionFromStatement(stmt, host)
			if err != nil {
				return err
			}
			return nil
		}, host.Host)
		if err != nil {
			return nil, err
		}
	}
	res := make([]Host, 0)
	for _, host := range hosts {
		res = append(res, *host)
	}
	return res, nil
}

func (dao *HostDao) GetAll() ([]Host, error) {
	queryString := `SELECT * FROM host`
	queryOptString := `SELECT * FROM host_options`
	hosts := map[string]*Host{}
	err := dao.conn.query(queryString, func(stmt *sqlite.Stmt) error {
		host := &Host{}
		err := dao.serializeHostFromStatement(stmt, host)
		if err != nil {
			return err
		}
		hosts[host.Host] = host
		return nil
	})
	if err != nil {
		return nil, err
	}
	err = dao.conn.query(queryOptString, func(stmt *sqlite.Stmt) error {
		host := stmt.GetText("host")
		hostObj, ok := hosts[host]
		if !ok {
			return fmt.Errorf("host not found in previous query %s", host)
		}
		err := dao.serializeHostOptionFromStatement(stmt, hostObj)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	hostSlice := make([]Host, 0, len(hosts))
	// convert map into slice
	for _, host := range hosts {
		hostSlice = append(hostSlice, *host)
	}
	return hostSlice, nil
}

func (dao *HostDao) Count() (uint, error) {
	queryString := `SELECT COUNT(*) FROM host`
	var count uint
	err := dao.conn.query(queryString, func(stmt *sqlite.Stmt) error {
		count = uint(stmt.ColumnInt64(0))
		return nil
	})
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (dao *HostDao) CountOpts(host string) (uint, error) {
	queryString := `SELECT COUNT(*) FROM host_options WHERE host = ?`
	var count uint
	err := dao.conn.query(queryString, func(stmt *sqlite.Stmt) error {
		count = uint(stmt.ColumnInt64(0))
		return nil
	}, host)
	if err != nil {
		return 0, err
	}
	return count, nil
}
