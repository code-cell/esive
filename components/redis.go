package components

import (
	"context"
	"strconv"

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

// NextInt64 uses the `key` as a sequence. It increments the sequence returning the new value.
func (s *RedisStore) NextInt64(ctx context.Context, key string) (int64, error) {
	res := s.client.Incr(ctx, key)
	if err := res.Err(); err != nil {
		s.logger.Error("error generating new int64", zap.Error(err))
		return 0, err
	}
	s.logger.Debug("generated new int64", zap.Int64("n", res.Val()))
	return res.Val(), nil
}

// SAdd Adds a member to a set
func (s *RedisStore) SAdd(ctx context.Context, key, value string) (bool, error) {
	res := s.client.SAdd(ctx, key, value)
	if err := res.Err(); err != nil {
		s.logger.Error("error adding member to set", zap.Error(err), zap.String("key", key), zap.String("value", value))
		return false, err
	}
	s.logger.Debug("added member to set", zap.String("key", key), zap.String("value", value))
	return res.Val() == 1, nil
}

// SRem Removes a member from a set
func (s *RedisStore) SRem(ctx context.Context, key, value string) error {
	res := s.client.SRem(ctx, key, value)
	if err := res.Err(); err != nil {
		s.logger.Error("error removing member from set", zap.Error(err), zap.String("key", key), zap.String("value", value))
		return err
	}
	s.logger.Debug("removed member from set", zap.String("key", key), zap.String("value", value))
	return nil
}

// HSaveProto saves a protocol buffers object into a hash, using the type of the object as a key within the hash.
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

// HDelProto deletes a protocol buffers object from a hash, using the type of the object as a key within the hash.
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

// HReadProtos reads multiple protocol buffers from a hash, using their types as keys within the hash.
func (s *RedisStore) HReadProtos(ctx context.Context, key string, values ...proto.Message) error {
	names := make([]string, len(values))
	for i, v := range values {
		names[i] = string(v.ProtoReflect().Descriptor().FullName().Name())
	}
	res := s.client.HMGet(ctx, key, names...)
	if err := res.Err(); err != nil {
		s.logger.Error("error getting hash members", zap.Error(err), zap.String("key", key), zap.Strings("names", names))
		return err
	}
	for i, item := range res.Val() {
		b := []byte(item.(string))
		err := proto.Unmarshal(b, values[i])
		if err != nil {
			s.logger.Error("error unmarshalling protos", zap.Error(err), zap.String("key", key), zap.Strings("names", names))
			return err
		}
		s.logger.Debug("loaded proto", zap.String("key", key), zap.Strings("names", names))
	}
	return nil
}

// Del deletes a key from redis.
func (s *RedisStore) Del(ctx context.Context, key string) error {
	res := s.client.Del(ctx, key)
	if err := res.Err(); err != nil {
		s.logger.Error("error deleting key", zap.Error(err), zap.String("key", key))
		return err
	}
	s.logger.Debug("deleted key", zap.String("key", key))
	return nil
}

// Sort fetches multiple protocol buffer objects (and their entities). `key` has to be an iteratable (a set most
//   probably) element with entity ids. The return will contain the Entity object and all the components.
func (s *RedisStore) Sort(ctx context.Context, key string, values ...proto.Message) ([]Entity, [][]proto.Message, error) {
	get := []string{"#"}
	for _, componentType := range values {
		componentType := componentType.ProtoReflect().Descriptor().FullName().Name()
		get = append(get, "*->"+string(componentType))
	}

	res := s.client.Sort(ctx, key, &redis.Sort{By: "nosort", Get: get})
	if err := res.Err(); err != nil {
		return nil, nil, err
	}

	entities := make([]Entity, 0)
	components := make([][]proto.Message, 0)

	resStr := res.Val()
	for i := 0; i < len(resStr)/(len(values)+1); i++ {
		idx := i * (len(values) + 1)
		parsed, err := strconv.ParseInt(resStr[idx], 10, 64)
		if err != nil {
			return nil, nil, err
		}

		entityComponents := make([]proto.Message, 0)
		allComponents := true
		for i, componentType := range values {
			err = proto.Unmarshal([]byte(resStr[idx+1+i]), componentType)
			if err != nil {
				return nil, nil, err
			}
			entityComponents = append(entityComponents, proto.Clone(componentType))
		}

		if allComponents {
			entities = append(entities, Entity(parsed))
			components = append(components, entityComponents)
		}
	}

	return entities, components, nil
}
