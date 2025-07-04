package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/devproje/neko-engine/util"
	"github.com/redis/go-redis/v9"
)

type RedisHistory struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Content   string    `json:"content"`
	Answer    string    `json:"answer"`
	CreatedAt time.Time `json:"created_at"`
}

type RedisHistoryRepository interface {
	Create(history *RedisHistory) error
	Read(uid string, limit int) ([]*RedisHistory, error)
	Delete(uid string, historyID string) error
	PurgeN(uid string, n int) error
	Flush(uid string) error
}

type redisHistoryRepository struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisHistoryRepository() RedisHistoryRepository {
	return &redisHistoryRepository{
		client: util.RedisClient,
		ctx:    context.Background(),
	}
}

func (r *redisHistoryRepository) getHistoryKey(uid string) string {
	return fmt.Sprintf("history:%s", uid)
}

func (r *redisHistoryRepository) getNextID(uid string) (string, error) {
	counterKey := fmt.Sprintf("history_counter:%s", uid)
	id, err := r.client.Incr(r.ctx, counterKey).Result()
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(id, 10), nil
}

func (r *redisHistoryRepository) Create(history *RedisHistory) error {
	if history.ID == "" {
		id, err := r.getNextID(history.UserID)
		if err != nil {
			return fmt.Errorf("failed to generate ID: %w", err)
		}
		history.ID = id
	}

	if history.CreatedAt.IsZero() {
		history.CreatedAt = time.Now()
	}

	data, err := json.Marshal(history)
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	key := r.getHistoryKey(history.UserID)
	score := float64(history.CreatedAt.Unix())

	return r.client.ZAdd(r.ctx, key, redis.Z{
		Score:  score,
		Member: data,
	}).Err()
}

func (r *redisHistoryRepository) Read(uid string, limit int) ([]*RedisHistory, error) {
	key := r.getHistoryKey(uid)

	results, err := r.client.ZRevRange(r.ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to read from Redis: %w", err)
	}

	histories := make([]*RedisHistory, 0, len(results))
	for i := len(results) - 1; i >= 0; i-- {
		var history RedisHistory
		if err := json.Unmarshal([]byte(results[i]), &history); err != nil {
			continue
		}
		histories = append(histories, &history)
	}

	return histories, nil
}

func (r *redisHistoryRepository) Delete(uid string, historyID string) error {
	key := r.getHistoryKey(uid)

	allMembers, err := r.client.ZRange(r.ctx, key, 0, -1).Result()
	if err != nil {
		return fmt.Errorf("failed to get all members: %w", err)
	}

	for _, member := range allMembers {
		var history RedisHistory
		if err := json.Unmarshal([]byte(member), &history); err != nil {
			continue
		}
		if history.ID == historyID {
			return r.client.ZRem(r.ctx, key, member).Err()
		}
	}

	return fmt.Errorf("history not found")
}

func (r *redisHistoryRepository) PurgeN(uid string, n int) error {
	key := r.getHistoryKey(uid)

	return r.client.ZRemRangeByRank(r.ctx, key, -int64(n), -1).Err()
}

func (r *redisHistoryRepository) Flush(uid string) error {
	key := r.getHistoryKey(uid)
	counterKey := fmt.Sprintf("history_counter:%s", uid)

	pipe := r.client.Pipeline()
	pipe.Del(r.ctx, key)
	pipe.Del(r.ctx, counterKey)
	_, err := pipe.Exec(r.ctx)

	return err
}
