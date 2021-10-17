package pkg

import (
	"fmt"
	"golang.org/x/xerrors"

	"github.com/jackc/pgproto3"
	"github.com/pg-sharding/spqr/pkg/config"
	"github.com/pg-sharding/spqr/qdb/qdb"
	"github.com/pg-sharding/spqr/router/pkg/qrouter"
	"github.com/pg-sharding/spqr/router/pkg/rrouter"
	"github.com/wal-g/tracelog"
)

type Qinteractor interface {
}

type QinteractorImpl struct {
}

func reroute(rst *rrouter.RelayStateImpl, v *pgproto3.Query) error {
	tracelog.InfoLogger.Printf("rerouting")
	_ = rst.Cl.ReplyNotice(fmt.Sprintf("rerouting your connection"))

	shrdRoutes, err := rst.Reroute(v)

	if err == qrouter.ShardMatchError {
		// do not reset connection
		return err
	}

	if err != nil {
		tracelog.InfoLogger.Printf("encounter %w", err)
		_ = rst.UnRouteWithError( nil, err)
		return err
	}

	_ = rst.Cl.ReplyNotice(fmt.Sprintf("matched shard routes %v", shrdRoutes))

	if err := rst.Connect(shrdRoutes); err != nil {
		tracelog.InfoLogger.Printf("encounter %w while initialing server connection", err)
		_ = rst.Reset()
		_ = rst.Cl.ReplyErr(err.Error())
		return err
	}

	return nil
}

func Frontend(qr qrouter.Qrouter, cl rrouter.RouterClient, cmngr rrouter.ConnManager) error {

	tracelog.InfoLogger.Printf("process Frontend for user %s %s", cl.Usr(), cl.DB())

	_ = cl.ReplyNotice(fmt.Sprintf("process Frontend for user %s %s", cl.Usr(), cl.DB()))

	rst := rrouter.NewRelayState(qr, cl, cmngr)

	for {
		msg, err := cl.Receive()
		if err != nil {
			tracelog.ErrorLogger.Printf("failed to receive msg %w", err)
			return err
		}

		tracelog.InfoLogger.Printf("received msg %v", msg)

		switch q := msg.(type) {
		case *pgproto3.Query:
			// txactive == 0 || activeSh == nil
			if cmngr.ValidateReRoute(rst) {
				if err := reroute(rst, q); err == qrouter.ShardMatchError {

					if !config.Get().RouterConfig.WorldShardFallback {
						return err
					}
					// fallback to execute query on wolrd shard (s)

					//

					_, _ = rst.RerouteWorld()
					if err := rst.ConnectWold(); err != nil {
						_ = rst.UnRouteWithError(nil, xerrors.Errorf("failed to fallback on world shard: %w", err))
						continue
					}

				} else if err != nil {
					continue
				}
			}

			var txst byte
			var err error
			if txst, err = rst.RelayStep(q); err != nil {
				if rst.ShouldRetry(err) {
					ch := make(chan interface{})

					status := qdb.KRUnLocked
					_ = rst.Qr.Subscribe(rst.TargetKeyRange.ID, &status, ch)
					<-ch
					// retry on master

					shrds, err := rst.Reroute(q)

					if err != nil {
						return err
					}

					if err := rst.Connect(shrds); err != nil {
						return err
					}

					rst.ReplayBuff()
				}
				return err
			}

			if err := rst.CompleteRelay(txst); err != nil {
				return err
			}

			tracelog.InfoLogger.Printf("active shards are %v", rst.ActiveShards)

		default:
		}
	}
}
