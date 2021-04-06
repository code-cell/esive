package components

import (
	"context"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type RedisStore struct {
	client *redis.Client
	logger *zap.Logger
}

func NewRedisStore(client *redis.Client, logger *zap.Logger) *RedisStore {
	return &RedisStore{
		client: client,
		logger: logger.With(zap.String("service", "redis_store")),
	}
}

func (s *RedisStore) NextInt64(ctx context.Context, key string) (int64, error) {
	res := s.client.Incr(ctx, key)
	if err := res.Err(); err != nil {
		s.logger.Error("error generating new int64", zap.Error(err))
		return 0, err
	}
	s.logger.Debug("generated new int64", zap.Int64("n", res.Val()))
	return res.Val(), nil
}

func (s *RedisStore) SAdd(ctx context.Context, key, value string) (bool, error) {
	res := s.client.SAdd(ctx, key, value)
	if err := res.Err(); err != nil {
		s.logger.Error("error adding member to set", zap.Error(err), zap.String("key", key), zap.String("value", value))
		return false, err
	}
	s.logger.Debug("added member to set", zap.String("key", key), zap.String("value", value))
	return res.Val() == 1, nil
}

func (s *RedisStore) SRem(ctx context.Context, key, value string) error {
	res := s.client.SRem(ctx, key, value)
	if err := res.Err(); err != nil {
		s.logger.Error("error removing member from set", zap.Error(err), zap.String("key", key), zap.String("value", value))
		return err
	}
	s.logger.Debug("removed member from set", zap.String("key", key), zap.String("value", value))
	return nil
}

func (s *RedisStore) SMembers(ctx context.Context, key string) ([]string, error) {
	res := s.client.SMembers(ctx, key)
	if err := res.Err(); err != nil {
		s.logger.Error("error fetching members for set", zap.Error(err), zap.String("key", key))
		return nil, err
	}
	s.logger.Debug("fetched members for set", zap.String("key", key))
	return res.Val(), nil
}

func (s *RedisStore) HSaveProto(ctx context.Context, key string, values ...proto.Message) error {
	args := []interface{}{}
	for _, v := range values {
		name := string(v.ProtoReflect().Descriptor().FullName().Name())
		out, err := proto.Marshal(v)
		if err != nil {
			s.logger.Error("error marshalling proto", zap.Error(err), zap.String("key", key), zap.String("name", name))
			return err
		}
		args = append(args, name, out)
	}
	res := s.client.HMSet(ctx, key, args...)
	if err := res.Err(); err != nil {
		s.logger.Error("error setting value in hash", zap.Error(err), zap.String("key", key))
		return err
	}
	s.logger.Debug("saved proto", zap.String("key", key))
	return nil
}

func (s *RedisStore) HDelProto(ctx context.Context, key string, v proto.Message) error {
	name := string(v.ProtoReflect().Descriptor().FullName().Name())
	res := s.client.HDel(ctx, key, name)
	if err := res.Err(); err != nil {
		s.logger.Error("error deleting hash member", zap.Error(err), zap.String("key", key), zap.String("name", name))
		return err
	}
	s.logger.Debug("deleted proto", zap.String("key", key), zap.String("name", name))
	return nil
}

func (s *RedisStore) HReadProto(ctx context.Context, key string, v proto.Message) error {
	name := string(v.ProtoReflect().Descriptor().FullName().Name())
	res := s.client.HGet(ctx, key, name)
	if err := res.Err(); err != nil {
		s.logger.Error("error getting hash member", zap.Error(err), zap.String("key", key), zap.String("name", name))
		return err
	}
	b, err := res.Bytes()
	if err != nil {
		s.logger.Error("error getting bytes", zap.Error(err), zap.String("key", key), zap.String("name", name))
		return err
	}
	err = proto.Unmarshal(b, v)
	if err != nil {
		s.logger.Error("error unmarshalling protos", zap.Error(err), zap.String("key", key), zap.String("name", name))
		return err
	}
	s.logger.Debug("loaded proto", zap.String("key", key), zap.String("name", name))
	return nil
}

func (s *RedisStore) HKeys(ctx context.Context, key string) ([]string, error) {
	res := s.client.HKeys(ctx, key)
	if err := res.Err(); err != nil {
		s.logger.Error("error getting hash keys", zap.Error(err), zap.String("key", key))
		return nil, err
	}
	s.logger.Debug("fetched hash keys", zap.String("key", key))
	return res.Result()
}

func (s *RedisStore) Del(ctx context.Context, key string) error {
	res := s.client.Del(ctx, key)
	if err := res.Err(); err != nil {
		s.logger.Error("error deleting key", zap.Error(err), zap.String("key", key))
		return err
	}
	s.logger.Debug("deleted key", zap.String("key", key))
	return nil
}
