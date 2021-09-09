package console

import "C"
import (
	"crypto/tls"
	"fmt"
	"github.com/pg-sharding/spqr/internal/qdb"

	"github.com/jackc/pgproto3"
	"github.com/pg-sharding/spqr/internal/config"
	"github.com/pg-sharding/spqr/internal/qlog"
	"github.com/pg-sharding/spqr/internal/qrouter"
	"github.com/pg-sharding/spqr/internal/rrouter"
	"github.com/pg-sharding/spqr/yacc/spqrparser"
	"github.com/pkg/errors"
	"github.com/wal-g/tracelog"
)

type Console interface {
	Serve(cl rrouter.Client) error
	ProcessQuery(q string, cl rrouter.Client) error
	Shutdown() error
}

type ConsoleDB struct {
	cfg     *tls.Config
	Qrouter qrouter.Qrouter
	Qlog    qlog.Qlog

	stchan chan struct{}
}

var _ Console = &ConsoleDB{}

func (c *ConsoleDB) Shutdown() error {
	return nil
}

func NewConsole(cfg *tls.Config, Qrouter qrouter.Qrouter, wal qlog.Qlog, stchan chan struct{}) *ConsoleDB {
	return &ConsoleDB{
		Qrouter: Qrouter,
		Qlog:    wal,
		cfg:     cfg,
		stchan:  stchan,
	}
}

func (c *ConsoleDB) Databases(cl rrouter.Client) error {
	for _, msg := range []pgproto3.BackendMessage{
		&pgproto3.RowDescription{Fields: []pgproto3.FieldDescription{
			{
				Name:                 "show dbs",
				TableOID:             0,
				TableAttributeNumber: 0,
				DataTypeOID:          25,
				DataTypeSize:         -1,
				TypeModifier:         -1,
				Format:               0,
			},
		},
		},
		&pgproto3.DataRow{Values: [][]byte{[]byte("show dbs")}},
		&pgproto3.CommandComplete{},
		&pgproto3.ReadyForQuery{},
	} {
		if err := cl.Send(msg); err != nil {
			tracelog.InfoLogger.Print(err)
		}
	}

	return nil
}

func (c *ConsoleDB) Pools(cl rrouter.Client) error {
	for _, msg := range []pgproto3.BackendMessage{
		&pgproto3.RowDescription{Fields: []pgproto3.FieldDescription{
			{
				Name:                 "fortune",
				TableOID:             0,
				TableAttributeNumber: 0,
				DataTypeOID:          25,
				DataTypeSize:         -1,
				TypeModifier:         -1,
				Format:               0,
			},
		},
		},
		&pgproto3.DataRow{Values: [][]byte{[]byte("show pools")}},
		&pgproto3.CommandComplete{},
		&pgproto3.ReadyForQuery{},
	} {
		if err := cl.Send(msg); err != nil {
			tracelog.InfoLogger.Print(err)
		}
	}

	return nil
}

func (c *ConsoleDB) AddShardingColumn(cl rrouter.Client, stmt *spqrparser.ShardingColumn) error {

	tracelog.InfoLogger.Printf("received create column request %s", stmt.ColName)

	err := c.Qrouter.AddShardingColumn(stmt.ColName)

	for _, msg := range []pgproto3.BackendMessage{
		&pgproto3.RowDescription{Fields: []pgproto3.FieldDescription{
			{
				Name:                 "fortune",
				TableOID:             0,
				TableAttributeNumber: 0,
				DataTypeOID:          25,
				DataTypeSize:         -1,
				TypeModifier:         -1,
				Format:               0,
			},
		},
		},
		&pgproto3.DataRow{Values: [][]byte{[]byte(fmt.Sprintf("created sharding column %s, err %w", stmt.ColName, err))}},
		&pgproto3.CommandComplete{},
		&pgproto3.ReadyForQuery{},
	} {
		if err := cl.Send(msg); err != nil {
			tracelog.InfoLogger.Print(err)
		}
	}

	return nil
}

func (c *ConsoleDB) SplitKeyRange(cl rrouter.Client, splitReq *spqrparser.SplitKeyRange) error {
	if err := c.Qrouter.Split(splitReq); err != nil {
		return err
	}

	tracelog.InfoLogger.Printf("splitted key range %v by %v", splitReq.KeyRangeFromID, splitReq.Border)

	for _, msg := range []pgproto3.BackendMessage{
		&pgproto3.RowDescription{Fields: []pgproto3.FieldDescription{
			{
				Name:                 "spqr",
				TableOID:             0,
				TableAttributeNumber: 0,
				DataTypeOID:          25,
				DataTypeSize:         -1,
				TypeModifier:         -1,
				Format:               0,
			},
		},
		},
		&pgproto3.DataRow{Values: [][]byte{[]byte(fmt.Sprintf("split key range %v by %v", splitReq.KeyRangeFromID, splitReq.Border))}},
		&pgproto3.CommandComplete{},
		&pgproto3.ReadyForQuery{},
	} {
		if err := cl.Send(msg); err != nil {
			tracelog.InfoLogger.Print(err)
		}
	}

	return nil
}

func (c *ConsoleDB) LockKeyRange(cl rrouter.Client, krid string) error {
	tracelog.InfoLogger.Printf("received lock key range req for id %v", krid)
	if err := c.Qrouter.Lock(krid); err != nil {
		return err
	}

	for _, msg := range []pgproto3.BackendMessage{
		&pgproto3.RowDescription{Fields: []pgproto3.FieldDescription{
			{
				Name:                 "spqr",
				TableOID:             0,
				TableAttributeNumber: 0,
				DataTypeOID:          25,
				DataTypeSize:         -1,
				TypeModifier:         -1,
				Format:               0,
			},
		},
		},
		&pgproto3.DataRow{Values: [][]byte{[]byte(fmt.Sprintf("lock key range with id %v", krid))}},
		&pgproto3.CommandComplete{},
		&pgproto3.ReadyForQuery{},
	} {
		if err := cl.Send(msg); err != nil {
			tracelog.InfoLogger.Print(err)
		}
	}

	return nil
}

func (c *ConsoleDB) AddKeyRange(cl rrouter.Client, kr *spqrparser.KeyRange) error {

	tracelog.InfoLogger.Printf("received create key range request %s for shard", kr.ShardID)

	err := c.Qrouter.AddKeyRange(qdb.KeyRange{
		KeyRangeID: kr.KeyRangeID,
		ShardID: kr.KeyRangeID,
		To: kr.To,
		From: kr.From,
	})

	for _, msg := range []pgproto3.BackendMessage{
		&pgproto3.RowDescription{Fields: []pgproto3.FieldDescription{
			{
				Name:                 "fortune",
				TableOID:             0,
				TableAttributeNumber: 0,
				DataTypeOID:          25,
				DataTypeSize:         -1,
				TypeModifier:         -1,
				Format:               0,
			},
		},
		},
		&pgproto3.DataRow{Values: [][]byte{[]byte(fmt.Sprintf("created key range from %d to %d, err %v", kr.From, kr.To, err))}},
		&pgproto3.CommandComplete{},
		&pgproto3.ReadyForQuery{},
	} {
		if err := cl.Send(msg); err != nil {
			tracelog.InfoLogger.Print(err)
		}
	}

	return nil
}

func (c *ConsoleDB) AddShard(cl rrouter.Client, shard *spqrparser.Shard, cfg *config.ShardCfg) error {

	err := c.Qrouter.AddShard(shard.Name, cfg)

	for _, msg := range []pgproto3.BackendMessage{
		&pgproto3.RowDescription{Fields: []pgproto3.FieldDescription{
			{
				Name:                 "fortune",
				TableOID:             0,
				TableAttributeNumber: 0,
				DataTypeOID:          25,
				DataTypeSize:         -1,
				TypeModifier:         -1,
				Format:               0,
			},
		},
		},
		&pgproto3.DataRow{Values: [][]byte{[]byte(fmt.Sprintf("created shard with name %s, %w", shard.Name, err))}},
		&pgproto3.CommandComplete{},
		&pgproto3.ReadyForQuery{},
	} {
		if err := cl.Send(msg); err != nil {
			tracelog.InfoLogger.Print(err)
		}
	}

	return nil
}

func (c *ConsoleDB) KeyRanges(cl rrouter.Client) error {

	tracelog.InfoLogger.Printf("listing key ranges")

	for _, msg := range []pgproto3.BackendMessage{
		&pgproto3.RowDescription{Fields: []pgproto3.FieldDescription{
			{
				Name:                 "spqr key ranges",
				TableOID:             0,
				TableAttributeNumber: 0,
				DataTypeOID:          25,
				DataTypeSize:         -1,
				TypeModifier:         -1,
				Format:               0,
			},
		},
		},
	} {
		if err := cl.Send(msg); err != nil {
			tracelog.InfoLogger.Print(err)
			return err
		}
	}

	for _, kr := range c.Qrouter.KeyRanges() {
		if err := cl.Send(&pgproto3.DataRow{
			Values: [][]byte{[]byte(fmt.Sprintf("key range %v for kr with %s", kr.KeyRangeID, kr.ShardID))},
		}); err != nil {
			tracelog.InfoLogger.Print(err)
		}
	}

	if err := cl.Send(&pgproto3.DataRow{
		Values: [][]byte{[]byte(fmt.Sprintf("local node"))},
	}); err != nil {
		tracelog.InfoLogger.Print(err)
	}

	for _, msg := range []pgproto3.BackendMessage{
		&pgproto3.CommandComplete{CommandTag: "SELECT 1"},
		&pgproto3.ReadyForQuery{},
	} {
		if err := cl.Send(msg); err != nil {
			tracelog.InfoLogger.Print(err)
		}
	}

	return nil
}

func (c *ConsoleDB) Shards(cl rrouter.Client) error {

	tracelog.InfoLogger.Printf("listing shards")

	for _, msg := range []pgproto3.BackendMessage{
		&pgproto3.RowDescription{Fields: []pgproto3.FieldDescription{
			{
				Name:                 "spqr shards",
				TableOID:             0,
				TableAttributeNumber: 0,
				DataTypeOID:          25,
				DataTypeSize:         -1,
				TypeModifier:         -1,
				Format:               0,
			},
		},
		},
	} {
		if err := cl.Send(msg); err != nil {
			tracelog.InfoLogger.Print(err)
			return err
		}
	}

	for _, shard := range c.Qrouter.Shards() {
		if err := cl.Send(&pgproto3.DataRow{
			Values: [][]byte{[]byte(fmt.Sprintf("shard with ID %s", shard))},
		}); err != nil {
			tracelog.InfoLogger.Print(err)
		}
	}

	if err := cl.Send(&pgproto3.DataRow{
		Values: [][]byte{[]byte(fmt.Sprintf("local node"))},
	}); err != nil {
		tracelog.InfoLogger.Print(err)
	}

	for _, msg := range []pgproto3.BackendMessage{
		&pgproto3.CommandComplete{CommandTag: "SELECT 1"},
		&pgproto3.ReadyForQuery{},
	} {
		if err := cl.Send(msg); err != nil {
			tracelog.InfoLogger.Print(err)
		}
	}

	return nil
}

func (c *ConsoleDB) ProcessQuery(q string, cl rrouter.Client) error {
	tstmt, err := spqrparser.Parse(q)
	if err != nil {
		return err
	}

	tracelog.InfoLogger.Printf("Get '%s', parsed %T", q, tstmt)

	switch stmt := tstmt.(type) {
	case *spqrparser.Show:

		tracelog.InfoLogger.Printf("parsed %s", stmt.Cmd)

		switch stmt.Cmd {

		case spqrparser.ShowPoolsStr:
			return c.Pools(cl)
		case spqrparser.ShowDatabasesStr:
			return c.Databases(cl)
		case spqrparser.ShowShardsStr:
			return c.Shards(cl)
		case spqrparser.ShowKeyRangesStr:
			return c.KeyRanges(cl)
		default:
			tracelog.InfoLogger.Printf("Unknown default %s", stmt.Cmd)

			return errors.New("Unknown default cmd: " + stmt.Cmd)
		}
	case *spqrparser.SplitKeyRange:
		err := c.SplitKeyRange(cl, stmt)
		if err != nil {
			c.Qlog.DumpQuery(q)
		}
		return err
	case *spqrparser.Lock:
		err := c.LockKeyRange(cl, stmt.KeyRangeID)
		if err != nil {
			c.Qlog.DumpQuery(q)
		}
		return err
	case *spqrparser.ShardingColumn:
		err := c.AddShardingColumn(cl, stmt)
		if err != nil {
			c.Qlog.DumpQuery(q)
		}
		return err
	case *spqrparser.KeyRange:
		err := c.AddKeyRange(cl, stmt)
		if err != nil {
			c.Qlog.DumpQuery(q)
		}
		return err
	case *spqrparser.Shard:
		err := c.AddShard(cl, stmt, &config.ShardCfg{})
		if err != nil {
			c.Qlog.DumpQuery(q)
		}
		return err
	case *spqrparser.Shutdown:
		c.stchan <- struct{}{}
		return nil
	default:
		tracelog.InfoLogger.Printf("got unexcepted console request %v %T", tstmt, tstmt)
		if err := cl.DefaultReply(); err != nil {
			tracelog.ErrorLogger.Fatal(err)
		}
	}

	return nil
}

func (c *ConsoleDB) Serve(cl rrouter.Client) error {
	tracelog.InfoLogger.Print("console.Serve start")
	queries, err := c.Qlog.Recover(config.Get().DataFolder)
	if err != nil {
		return errors.Wrap(err, "Serve can't start")
	}
	for _, query := range queries {
		if err := c.ProcessQuery(query, cl); err != nil {
			continue // TODO fix 'syntax error'
			// return errors.Wrap(err, "Serve init fail")
		}
	}

	tracelog.InfoLogger.Printf("Succesfully init %d queries", len(queries))

	for {
		msg, err := cl.Receive()

		if err != nil {
			return err
		}

		switch v := msg.(type) {
		case *pgproto3.Query:
			if err := c.ProcessQuery(v.String, cl); err != nil {
				_ = cl.ReplyErr(err.Error())
				return err
			}
		default:
			tracelog.InfoLogger.Printf("got unexpected postgresql proto message with type %T", v)
		}
	}
}
