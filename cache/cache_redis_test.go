package cache

import (
	"context"
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
