package qrouter

import (
	"errors"

	"github.com/pg-sharding/spqr/pkg/config"
	"github.com/pg-sharding/spqr/pkg/models/kr"
	"github.com/pg-sharding/spqr/qdb/qdb"
	spqrparser "github.com/pg-sharding/spqr/yacc/console"
)

type LocalQrouter struct {
	shid string
}

func (l *LocalQrouter) AddWorldShard(name string, cfg *config.ShardCfg) error {

	panic("implement me")
}

func (l *LocalQrouter) WorldShardsRoutes() []ShardRoute {

	panic("implement me")
}

func (l *LocalQrouter) WorldShards() []string {
	panic("implement me")
}

var _ Qrouter = &LocalQrouter{}

func NewLocalQrouter(shid string) (*LocalQrouter, error) {
	return &LocalQrouter{
		shid,
	}, nil
}

func (l *LocalQrouter) Subscribe(krid string, krst *qdb.KeyRangeStatus, noitfyio chan<- interface{}) error {
	panic("implement me")
}

func (l *LocalQrouter) Unite(req *spqrparser.UniteKeyRange) error {
	panic("implement me")
}

func (l *LocalQrouter) AddLocalTable(tname string) error {
	return errors.New("local qrouter does not support sharding")
}

func (l *LocalQrouter) AddKeyRange(kr kr.KeyRange) error {
	return errors.New("local qrouter does not support sharding")
}

func (l *LocalQrouter) Shards() []string {
	return []string{l.shid}
}

func (l *LocalQrouter) KeyRanges() []kr.KeyRange {
	return nil
}

func (l *LocalQrouter) AddDataShard(name string, cfg *config.ShardCfg) error {
	return errors.New("local qrouter does not support sharding")
}

func (l *LocalQrouter) Lock(krid string) error {
	return errors.New("local qrouter does not support sharding")
}

func (l *LocalQrouter) UnLock(krid string) error {
	return errors.New("local qrouter does not support sharding")
}

func (l *LocalQrouter) Split(req *spqrparser.SplitKeyRange) error {
	return errors.New("local qrouter does not support sharding")
}

func (l *LocalQrouter) AddShardingColumn(col string) error {
	return errors.New("local qoruter does not supprort sharding")
}

func (l *LocalQrouter) Route(q string) []ShardRoute {
	return []ShardRoute{
		{
			Shkey: kr.ShardKey{
				Name: l.shid,
			},
		},
	}
}
