package listener

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/ezhdanovskiy/postgres-go-experiments/internal/config"
)

type Listener struct {
	log *zap.SugaredLogger

	db *sqlx.DB

	channelName string
	listener    *pq.Listener
	failed      chan error

	ctx    context.Context
	cancel context.CancelFunc
}

func NewListener(logger *zap.SugaredLogger, cfg *config.Config) (*Listener, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("migrate NewWithDatabaseInstance: %w", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return nil, fmt.Errorf("migrate Up: %w", err)
	}

	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNoChange {
		return nil, fmt.Errorf("migrate version: %w", err)
	}
	logger.With("version", version, "dirty", dirty).Info("Migrations applied")

	listener := &Listener{
		log:         logger,
		db:          db,
		failed:      make(chan error, 2),
		channelName: cfg.DBChannelName,
	}

	pgListener := pq.NewListener(dsn, 10*time.Second, time.Minute, listener.logListener)

	if err := pgListener.Listen(cfg.DBChannelName); err != nil {
		_ = pgListener.Close()
		return nil, fmt.Errorf("pg listen: %w", err)
	}
	listener.listener = pgListener

	listener.ctx, listener.cancel = context.WithCancel(context.Background())

	return listener, nil
}

func (l *Listener) Run() error {
	l.log.Info("Run Listener")
	defer l.log.Info("Listener stopped")

	for {
		select {
		case e := <-l.listener.Notify:
			l.log.Infof("event: %+v", e)
			if e == nil {
				continue
			}
			go l.handleAlerts()
		case err := <-l.failed:
			return err
		case <-l.ctx.Done():
			return nil
		}
	}
}

func (l *Listener) logListener(event pq.ListenerEventType, err error) {
	l.log.Debugf("logListener: %v", event)
	if err != nil {
		l.log.Errorf("listener error: %s", err)
	}
	if event == pq.ListenerEventConnectionAttemptFailed {
		l.failed <- err
	}
}

func (l *Listener) Stop() {
	if l.cancel != nil {
		l.cancel()
	}
}

func (l *Listener) handleAlerts() {
	log := l.log.With("method", "handleAlerts")
	log.Info("handleAlerts start")

	const query = `
SELECT * 
FROM alerts 
WHERE marked_to_send_at IS NOT NULL 
  AND sent_at IS NULL 
LIMIT 1
FOR UPDATE SKIP LOCKED
`

	for {
		tx, err := l.db.Beginx()
		if err != nil {
			log.Errorf("start transaction: %s", err)
			return
		}

		var alert Alert
		err = tx.QueryRowx(query).StructScan(&alert)
		if err != nil {
			if err == sql.ErrNoRows {
				break
			}
			log.Errorf("query: %s", err)
			return
		}
		go l.handleOneAlert(tx, alert)

		select {
		case <-l.ctx.Done():
			log.Errorf("ctx.Done: %s", err)
			return
		default:
		}
	}
}

func (l *Listener) handleOneAlert(tx *sqlx.Tx, alert Alert) {
	log := l.log.With("method", "handleOneAlert", "alert_id", alert.ID)
	log.Info("handleOneAlert start")
	err := l.sendAlert(alert)
	if err != nil {
		log.Errorf("send alert: %s", err)
		_ = tx.Rollback()
		return
	}

	_, err = tx.Exec("UPDATE alerts SET sent_at = now() WHERE id = $1", alert.ID)
	if err != nil {
		log.Errorf("update: %s", err)
		_ = tx.Rollback()
		return
	}

	if err := tx.Commit(); err != nil {
		log.Errorf("commit: %s", err)
	}
}

func (l *Listener) sendAlert(alert Alert) error {
	log := l.log.With("alert_id", alert.ID)
	log.Infof("sendAlert start")
	time.Sleep(time.Second * 3)
	log.Infof("sendAlert stop")
	return nil
}
