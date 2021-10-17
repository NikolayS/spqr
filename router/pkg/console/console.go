package console

import "C"
import (
	"crypto/tls"
	"fmt"

	"github.com/jackc/pgproto3"
	"github.com/pg-sharding/spqr/pkg/client"
	"github.com/pg-sharding/spqr/pkg/config"
	"github.com/pg-sharding/spqr/pkg/models/kr"
	"github.com/pg-sharding/spqr/router/pkg/qlog"
	qlogprovider "github.com/pg-sharding/spqr/router/pkg/qlog/provider"
	"github.com/pg-sharding/spqr/router/pkg/qrouter"
	spqrparser "github.com/pg-sharding/spqr/yacc/console"
	"github.com/pkg/errors"
	"github.com/wal-g/tracelog"
)

type Console interface {
	Serve(cl client.Client) error
	ProcessQuery(q string, cl client.Client) error
	Shutdown() error
}

type Local struct {
	cfg     *tls.Config
	Qrouter qrouter.Qrouter
	Qlog    qlog.Qlog

	stchan chan struct{}
}

var _ Console = &Local{}

func (c *Local) Shutdown() error {
	return nil
}

func NewConsole(cfg *tls.Config, Qrouter qrouter.Qrouter, stchan chan struct{}) (*Local, error) {
	localQlog, err := qlogprovider.NewLocalQlog(config.Get().DataFolder)
	if err != nil {
		return nil, err
	}
	return &Local{
		Qrouter: Qrouter,
		Qlog:    localQlog,
		cfg:     cfg,
		stchan:  stchan,
	}, nil
}

func (c *Local) Databases(cl client.Client) error {
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

func (c *Local) Pools(cl client.Client) error {
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

func (c *Local) AddShardingColumn(cl client.Client, stmt *spqrparser.ShardingColumn) error {

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

func (c *Local) SplitKeyRange(cl client.Client, splitReq *spqrparser.SplitKeyRange) error {
	if err := c.Qrouter.Split(splitReq); err != nil {
		return err
	}

	tracelog.InfoLogger.Printf("splitted key range %v by %v", splitReq.KeyRangeFromID, splitReq.Border)

	for _, msg := range []pgproto3.BackendMessage{
		&pgproto3.RowDescription{Fields: []pgproto3.FieldDescription{
			{
				Name:                 "worldmock",
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

func (c *Local) LockKeyRange(cl client.Client, krid string) error {
	tracelog.InfoLogger.Printf("received lock key range req for id %v", krid)
	if err := c.Qrouter.Lock(krid); err != nil {
		return err
	}

	for _, msg := range []pgproto3.BackendMessage{
		&pgproto3.RowDescription{Fields: []pgproto3.FieldDescription{
			{
				Name:                 "worldmock",
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

func (c *Local) AddKeyRange(cl client.Client, keyRange *spqrparser.KeyRange) error {

	tracelog.InfoLogger.Printf("received create key range request %s for shard", keyRange.ShardID)

	err := c.Qrouter.AddKeyRange(kr.KeyRange{
		ID:         keyRange.KeyRangeID,
		Shid:       keyRange.ShardID,
		UpperBound: keyRange.To,
		LowerBound: keyRange.From,
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
		&pgproto3.DataRow{Values: [][]byte{[]byte(fmt.Sprintf("created key range from %d to %d, err %v", keyRange.From, keyRange.To, err))}},
		&pgproto3.CommandComplete{},
		&pgproto3.ReadyForQuery{},
	} {
		if err := cl.Send(msg); err != nil {
			tracelog.InfoLogger.Print(err)
		}
	}

	return nil
}

func (c *Local) AddShard(cl client.Client, shard *spqrparser.Shard, cfg *config.ShardCfg) error {

	err := c.Qrouter.AddDataShard(shard.Name, cfg)

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

func (c *Local) KeyRanges(cl client.Client) error {

	tracelog.InfoLogger.Printf("listing key ranges")

	for _, msg := range []pgproto3.BackendMessage{
		&pgproto3.RowDescription{Fields: []pgproto3.FieldDescription{
			{
				Name:                 "worldmock key ranges",
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
			Values: [][]byte{[]byte(fmt.Sprintf("key range %v for kr with %s", kr.ID, kr.Shid))},
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

func (c *Local) Shards(cl client.Client) error {

	tracelog.InfoLogger.Printf("listing shards")

	for _, msg := range []pgproto3.BackendMessage{
		&pgproto3.RowDescription{Fields: []pgproto3.FieldDescription{
			{
				Name:                 "worldmock shards",
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

func (c *Local) ProcessQuery(q string, cl client.Client) error {
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
			_ = c.Qlog.DumpQuery(q)
		}
		return err
	case *spqrparser.Lock:
		err := c.LockKeyRange(cl, stmt.KeyRangeID)
		if err != nil {
			_ = c.Qlog.DumpQuery(q)
		}
		return err
	case *spqrparser.ShardingColumn:
		err := c.AddShardingColumn(cl, stmt)
		if err != nil {
			_ = c.Qlog.DumpQuery(q)
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
			_ = c.Qlog.DumpQuery(q)
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

const greeting = `

		SQPR router admin console

	Here you can configure your routing rules
------------------------------------------------

	You can find documentation here 
https://github.com/pg-sharding/spqr/tree/master/doc/router

`

func (c *Local) Serve(cl client.Client) error {

	for _, msg := range []pgproto3.BackendMessage{
		&pgproto3.Authentication{Type: pgproto3.AuthTypeOk},
		&pgproto3.ParameterStatus{Name: "integer_datetimes", Value: "on"},
		&pgproto3.ParameterStatus{Name: "server_version", Value: "console"},
		&pgproto3.NoticeResponse{
			Message: greeting,
		},
		&pgproto3.ReadyForQuery{},
	} {
		if err := cl.Send(msg); err != nil {
			tracelog.ErrorLogger.Fatal(err)
		}
	}

	tracelog.InfoLogger.Print("console.Serve start")

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
