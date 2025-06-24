package tarantool

import (
	"context"
	"errors"
	"fmt"
	"github.com/tarantool/go-iproto"
	"github.com/tarantool/go-tarantool/v2"
	"log"
	"log/slog"
	"time"
	"tt-demo/internal/config"
	"tt-demo/internal/model"
	"tt-demo/internal/storage"
)

const SpaceName = "key_value"

type Storage struct {
	conn *tarantool.Connection
	log  *slog.Logger
}

func New(config config.DB, log *slog.Logger) (*Storage, error) {
	const op = "storage.tarantool.New"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	dialer := tarantool.NetDialer{
		Address:  fmt.Sprintf("%s:%d", config.Host, config.Port),
		User:     config.Username,
		Password: config.Password,
	}
	opts := tarantool.Opts{
		Timeout: time.Second,
	}

	// TODO рассмотреть вариант использования ConnectionPool
	conn, err := tarantool.Connect(ctx, dialer, opts)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	createSpace(conn)

	return &Storage{conn: conn, log: log}, nil
}

func (s *Storage) GetValue(ctx context.Context, key string) (string, error) {
	const op = "storage.tarantool.GetValue"
	vals := make([]*model.KeyValue, 1)

	err := s.conn.Do(
		tarantool.NewSelectRequest(SpaceName).
			Context(ctx).
			Limit(1).
			Iterator(tarantool.IterEq).
			Key([]interface{}{key}),
	).GetTyped(&vals)

	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if len(vals) == 0 {
		return "", storage.KeyNotFound
	}
	return vals[0].Value, nil
}

func (s *Storage) SetValue(ctx context.Context, key, value string) error {
	const op = "storage.tarantool.SetValue"

	_, err := s.conn.Do(
		tarantool.NewInsertRequest(SpaceName).
			Context(ctx).
			Tuple([]interface{}{key, value}),
	).Get()

	if err != nil {
		var e tarantool.Error

		if errors.As(err, &e) && e.Code == iproto.ER_TUPLE_FOUND {
			return storage.DuplicatedKey
		} else {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}

func (s *Storage) UpdateValue(ctx context.Context, key, value string) error {
	const op = "storage.tarantool.UpdateValue"

	_, err := s.conn.Do(
		tarantool.NewUpdateRequest(SpaceName).
			Context(ctx).
			Key([]interface{}{key}).
			Operations(tarantool.NewOperations().Assign(1, value)),
	).Get()

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) Delete(ctx context.Context, key string) error {
	const op = "storage.tarantool.SetValue"

	_, err := s.conn.Do(
		tarantool.NewDeleteRequest(SpaceName).
			Context(ctx).
			Key([]interface{}{key}),
	).Get()

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) Close() error {
	return s.conn.CloseGraceful()
}

// TODO Перенести в миграции
func createSpace(conn *tarantool.Connection) {
	request := tarantool.NewEvalRequest("box.schema.space.create('key_value', { if_not_exists = true, engine = 'memtx' })")

	_, err := conn.Do(request).Get()
	if err != nil {
		log.Fatalf("Failed to create space: %s", err)
	}

	request = tarantool.NewEvalRequest("box.space.key_value:format({{ name='id', type='string' }, { name='value', type='string' }})")

	_, err = conn.Do(request).Get()
	if err != nil {
		log.Fatalf("Failed to format space: %s", err)
	}

	request = tarantool.NewEvalRequest("box.space.key_value:create_index('primary', {parts = { 'id' }, type = 'tree', if_not_exists = true })")

	_, err = conn.Do(request).Get()
	if err != nil {
		log.Fatalf("Failed to create index: %s", err)
	}
}
