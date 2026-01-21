package listener

import (
	"context"
	"encoding/json"
	"fmt"

	"config-client/config/domain/listener"

	"github.com/redis/go-redis/v9"
)

const (
	// ConfigChangeChannel Redis Pub/Sub 通道名称
	ConfigChangeChannel = "config:change"
)

// RedisConfigListener 基于Redis Pub/Sub的配置变更监听器
type RedisConfigListener struct {
	client *redis.Client
	pubsub *redis.PubSub
}

// NewRedisConfigListener 创建Redis配置监听器
func NewRedisConfigListener(client *redis.Client) *RedisConfigListener {
	return &RedisConfigListener{
		client: client,
	}
}

// Subscribe 订阅配置变更
func (l *RedisConfigListener) Subscribe(ctx context.Context) (<-chan *listener.ConfigChangeEvent, error) {
	// 订阅Redis频道
	l.pubsub = l.client.Subscribe(ctx, ConfigChangeChannel)

	// 等待订阅确认
	_, err := l.pubsub.Receive(ctx)
	if err != nil {
		return nil, fmt.Errorf("订阅配置变更频道失败: %w", err)
	}

	// 创建事件通道
	eventChan := make(chan *listener.ConfigChangeEvent, 100)

	// 启动goroutine接收消息
	go func() {
		defer close(eventChan)
		ch := l.pubsub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				// 解析消息
				var event listener.ConfigChangeEvent
				if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
					continue
				}
				// 发送事件
				select {
				case eventChan <- &event:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return eventChan, nil
}

// Publish 发布配置变更事件
func (l *RedisConfigListener) Publish(ctx context.Context, event *listener.ConfigChangeEvent) error {
	// 序列化事件
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("序列化配置变更事件失败: %w", err)
	}

	// 发布到Redis频道
	if err := l.client.Publish(ctx, ConfigChangeChannel, data).Err(); err != nil {
		return fmt.Errorf("发布配置变更事件失败: %w", err)
	}

	return nil
}

// Close 关闭监听器
func (l *RedisConfigListener) Close() error {
	if l.pubsub != nil {
		return l.pubsub.Close()
	}
	return nil
}
