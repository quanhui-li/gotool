package cache

import (
	"context"
	"fmt"
	"github.com/go-playground/assert/v2"
	"github.com/golang/mock/gomock"
	"github.com/liquanhui-99/gotool/cache/mocks"
	"github.com/redis/go-redis/v9"
	"testing"
	"time"
)

func TestRedisCache_Set(t *testing.T) {
	testCases := []struct {
		name       string
		key        string
		value      string
		mock       func(ctrl *gomock.Controller) redis.Cmdable
		expiration time.Duration
		wantErr    error
	}{
		{
			name: "set value",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				status := redis.NewStatusCmd(context.Background())
				status.SetVal("OK")
				cmd := mocks.NewMockCmdable(ctrl)
				cmd.EXPECT().Set(context.Background(), "key1", "value1", time.Second).Return(status)
				return cmd
			},
			key:        "key1",
			value:      "value1",
			expiration: time.Second,
		},
		{
			name: "timeout",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				status := redis.NewStatusCmd(context.Background())
				status.SetErr(context.DeadlineExceeded)
				cmd := mocks.NewMockCmdable(ctrl)
				cmd.EXPECT().Set(context.Background(), "key1", "value1", time.Second).Return(status)
				return cmd
			},
			key:        "key1",
			value:      "value1",
			expiration: time.Second,
			wantErr:    context.DeadlineExceeded,
		},
		{
			name: "unexpected value",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				status := redis.NewStatusCmd(context.Background())
				status.SetErr(ErrFailToSetKey)
				cmd := mocks.NewMockCmdable(ctrl)
				cmd.EXPECT().Set(context.Background(), "key1", "value1", time.Second).Return(status)
				return cmd
			},
			key:        "key1",
			value:      "value1",
			expiration: time.Second,
			wantErr:    ErrFailToSetKey,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := NewRedisCache(tc.mock(ctrl))
			err := c.Set(context.Background(), tc.key, tc.value, tc.expiration)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func TestRedisCache_Get(t *testing.T) {
	testCases := []struct {
		name      string
		key       string
		mock      func(ctrl *gomock.Controller) redis.Cmdable
		wantErr   error
		wantValue string
	}{
		{
			name: "get value",
			key:  "key1",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				status := redis.NewStringCmd(context.Background())
				status.SetVal("value1")
				cmd.EXPECT().Get(context.Background(), "key1").Return(status)
				return cmd
			},
			wantValue: "value1",
		},
		{
			name: "get error",
			key:  "key2",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				status := redis.NewStringCmd(context.Background())
				status.SetErr(ErrKeyNotExisted)
				cmd.EXPECT().Get(context.Background(), "key2").Return(status)
				return cmd
			},
			wantErr: ErrKeyNotExisted,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := NewRedisCache(tc.mock(ctrl))
			val, err := c.Get(context.Background(), tc.key)
			assert.Equal(t, tc.wantErr, err)
			fmt.Println(val)
		})
	}
}

func TestRedisCache_Delete(t *testing.T) {
	testCases := []struct {
		name    string
		key     string
		mock    func(ctrl *gomock.Controller) redis.Cmdable
		wantErr error
	}{
		{
			name: "delete key",
			key:  "key1",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				status := redis.NewStatusCmd(context.Background())
				status.SetErr(nil)
				cmd.EXPECT().Del(context.Background(), "key1").Return(status)
				return cmd
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := NewRedisCache(tc.mock(ctrl))
			cnt, err := c.Delete(context.Background(), tc.key)
			assert.Equal(t, tc.wantErr, err)
			fmt.Println(cnt)
		})
	}
}
