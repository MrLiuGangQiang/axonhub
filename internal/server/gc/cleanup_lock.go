package gc

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"

	"github.com/looplj/axonhub/internal/log"
)

// The value is arbitrary but must remain stable so all AxonHub instances contend
// for the same PostgreSQL advisory lock.
const postgresCleanupLockKey int64 = 7496163520481175303

const mysqlCleanupLockKey = "axonhub:gc:cleanup"

// acquireCleanupLock acquires a best-effort cross-instance cleanup lock. SQLite
// deployments are commonly single-process and return a no-op lock. Returning
// acquired=false lets a concurrent cleanup run finish without duplicating its
// scans and writes.
func (w *Worker) acquireCleanupLock(ctx context.Context) (func(context.Context), bool, error) {
	driver, ok := w.Ent.Driver().(*entsql.Driver)
	if !ok {
		return nil, false, fmt.Errorf("failed to get underlying SQL driver")
	}

	switch driver.Dialect() {
	case dialect.Postgres, dialect.MySQL:
	default:
		return func(context.Context) {}, true, nil
	}

	conn, err := driver.DB().Conn(ctx)
	if err != nil {
		return nil, false, fmt.Errorf("failed to acquire cleanup lock connection: %w", err)
	}

	release := func(releaseCtx context.Context) {
		defer func() { _ = conn.Close() }()

		var query string
		var arg any
		if driver.Dialect() == dialect.Postgres {
			query = "SELECT pg_advisory_unlock($1)"
			arg = postgresCleanupLockKey
		} else {
			query = "SELECT RELEASE_LOCK(?)"
			arg = mysqlCleanupLockKey
		}

		if _, err := conn.ExecContext(releaseCtx, query, arg); err != nil {
			log.Warn(releaseCtx, "Failed to release cleanup lock", log.Cause(err))
		}
	}

	if driver.Dialect() == dialect.Postgres {
		var acquired bool
		if err := conn.QueryRowContext(ctx, "SELECT pg_try_advisory_lock($1)", postgresCleanupLockKey).Scan(&acquired); err != nil {
			_ = conn.Close()
			return nil, false, fmt.Errorf("failed to acquire PostgreSQL cleanup lock: %w", err)
		}
		if !acquired {
			_ = conn.Close()
			return nil, false, nil
		}

		return release, true, nil
	}

	var acquired int
	if err := conn.QueryRowContext(ctx, "SELECT GET_LOCK(?, 0)", mysqlCleanupLockKey).Scan(&acquired); err != nil {
		_ = conn.Close()
		return nil, false, fmt.Errorf("failed to acquire MySQL cleanup lock: %w", err)
	}
	if acquired != 1 {
		_ = conn.Close()
		return nil, false, nil
	}

	return release, true, nil
}
