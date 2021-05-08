package notifier

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/ezhdanovskiy/postgres-go-experiments/internal/config"
)

type Notifier struct {
	log         *zap.SugaredLogger
	db          *sqlx.DB
	channelName string
	cancel      context.CancelFunc
}

func NewNotifier(logger *zap.SugaredLogger, cfg *config.Config) (*Notifier, error) {
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

	return &Notifier{
		log:         logger,
		db:          db,
		channelName: cfg.DBChannelName,
	}, nil
}

func (n *Notifier) Run() error {
	n.log.Info("Run Notifier")
	defer n.log.Info("Notifier stopped")

	ctx, cancel := context.WithCancel(context.Background())
	n.cancel = cancel

	for {
		n.log.Infof("NOTIFY %s", n.channelName)
		_, err := n.db.Exec("NOTIFY " + n.channelName)
		if err != nil {
			return fmt.Errorf("NOTIFY %s: %w", n.channelName, err)
		}

		select {
		case <-time.After(time.Second * 5):
			n.log.Info("Next iteration")
		case <-ctx.Done():
			return nil
		}
	}
}

func (n *Notifier) Stop() {
	if n.cancel != nil {
		n.cancel()
	}
}
