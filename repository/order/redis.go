package order

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zivattias/golang-redis-microservice/model"
	"github.com/zivattias/golang-redis-microservice/utils"
)

type RedisRepo struct {
	Client *redis.Client
}

func orderIDKey(id uint64) string {
	return fmt.Sprintf("order:%d", id)
}

func (r *RedisRepo) Insert(ctx context.Context, order model.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to encode order: %w", err)
	}

	key := orderIDKey(order.OrderID)

	// Redis transaction initialization
	txn := r.Client.TxPipeline()

	res := txn.SetNX(ctx, key, string(data), time.Hour*24)
	if err := res.Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to set: %w", err)
	}

	if err := txn.SAdd(ctx, "orders", key).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to add to orders set: %w", err)
	}

	if _, err := txn.Exec(ctx); err != nil {
		return fmt.Errorf("failed to exec insert transaction: %w", err)
	}

	return nil
}

func (r *RedisRepo) FindById(ctx context.Context, id uint64) (model.Order, error) {
	key := orderIDKey(id)

	value, err := r.Client.Get(ctx, key).Result()

	if errors.Is(err, redis.Nil) {
		return model.Order{}, utils.ErrNotExist
	} else if err != nil {
		return model.Order{}, fmt.Errorf("failed to get order: %w", err)
	}

	var order model.Order
	err = json.Unmarshal([]byte(value), &order)
	if err != nil {
		return model.Order{}, fmt.Errorf("failed to decode order json: %w", err)
	}

	return order, nil
}

func (r *RedisRepo) DeleteById(ctx context.Context, id uint64) error {
	key := orderIDKey(id)

	txn := r.Client.TxPipeline()

	err := txn.Del(ctx, key).Err()
	if errors.Is(err, redis.Nil) {
		txn.Discard()
		return utils.ErrNotExist
	} else if err != nil {
		txn.Discard()
		return fmt.Errorf("failed to delete order: %w", err)
	}

	if err := txn.SRem(ctx, "orders", key).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to remove from orders set: %w", err)
	}

	return nil
}

func (r *RedisRepo) UpdateById(ctx context.Context, order model.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to encode order: %w", err)
	}

	key := orderIDKey(order.OrderID)

	err = r.Client.SetXX(ctx, key, string(data), time.Hour*24).Err()
	if errors.Is(err, redis.Nil) {
		return utils.ErrNotExist
	} else if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	return nil
}

func (r *RedisRepo) FindAll(ctx context.Context, page model.FindAllPage) (model.FindAllResult, error) {
	res := r.Client.SScan(ctx, "orders", page.Offset, "*", int64(page.Size))

	keys, cursor, err := res.Result()
	if err != nil {
		return model.FindAllResult{}, fmt.Errorf("failed to find all orders: %w", err)
	}

	if len(keys) == 0 {
		return model.FindAllResult{
			Orders: []model.Order{},
			Cursor: 0,
		}, nil
	}

	xs, err := r.Client.MGet(ctx, keys...).Result()
	if err != nil {
		return model.FindAllResult{}, fmt.Errorf("failed to get orders: %w", err)
	}

	orders := make([]model.Order, len(xs))

	for i, x := range xs {
		x := x.(string)
		var order model.Order

		err := json.Unmarshal([]byte(x), &order)
		if err != nil {
			return model.FindAllResult{}, fmt.Errorf("failed to decode order json: %w", err)
		}

		orders[i] = order
	}

	return model.FindAllResult{
		Orders: orders,
		Cursor: cursor,
	}, nil
}
