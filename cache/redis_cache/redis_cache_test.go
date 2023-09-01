package redis_cache

import (
	"context"
	"github.com/go-playground/assert/v2"
	"github.com/golang/mock/gomock"
	"github.com/liquanhui-99/gotool/cache/redis_cache/mocks"
	"github.com/redis/go-redis/v9"
	"testing"
	"time"
)

func TestRedisCache_Set(t *testing.T) {
	var a = struct {
		name string
	}{
		name: "1",
	}
	testCases := []struct {
		name       string
		key        string
		val        any
		wantErr    error
		expiration time.Duration
		mock       func(controller *gomock.Controller) redis.Cmdable
	}{
		{
			name:       "set struct key",
			key:        "set struct key",
			val:        a,
			wantErr:    nil,
			expiration: 10 * time.Second,
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				statusCmd := redis.NewStatusCmd(context.Background())
				statusCmd.SetVal("OK")
				statusCmd.SetErr(nil)
				cmd.EXPECT().Set(context.Background(), "set struct key", a, 10*time.Second).
					Return(statusCmd)
				return cmd
			},
		},
		{
			name:       "set key",
			key:        "set key",
			val:        "1",
			wantErr:    nil,
			expiration: 10 * time.Second,
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				statusCmd := redis.NewStatusCmd(context.Background())
				statusCmd.SetVal("OK")
				statusCmd.SetErr(nil)
				cmd.EXPECT().Set(context.Background(), "set key", "1", 10*time.Second).
					Return(statusCmd)
				return cmd
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			cache := NewRedisCache(tc.mock(ctrl))
			err := cache.Set(context.Background(), tc.key, tc.val, tc.expiration)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func TestRedisCache_Get(t *testing.T) {
	testCases := []struct {
		name    string
		key     string
		wantErr error
		mock    func(controller *gomock.Controller) redis.Cmdable
		wantVal any
	}{
		{
			name:    "get string",
			key:     "get string",
			wantErr: nil,
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				status := redis.NewStringCmd(context.Background())
				status.SetVal("12344")
				status.SetErr(nil)
				cmd.EXPECT().Get(context.Background(), "get string").Return(status)
				return cmd
			},
			wantVal: "12344",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			cache := NewRedisCache(tc.mock(ctrl))
			val, err := cache.Get(context.Background(), tc.key)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantVal, val)
		})

	}
}

func TestRedisCache_Delete(t *testing.T) {
	testCases := []struct {
		name    string
		key     string
		wantErr error
		mock    func(controller *gomock.Controller) redis.Cmdable
	}{
		{
			name:    "delete string",
			key:     "delete string",
			wantErr: nil,
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				status := redis.NewIntCmd(context.Background())
				status.SetErr(nil)
				cmd.EXPECT().Del(context.Background(), "delete string").Return(status)
				return cmd
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			cache := NewRedisCache(tc.mock(ctrl))
			err := cache.Delete(context.Background(), tc.key)
			assert.Equal(t, tc.wantErr, err)
		})

	}
}
