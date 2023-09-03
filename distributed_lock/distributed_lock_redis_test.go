package distributed_lock

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-playground/assert/v2"
	"github.com/golang/mock/gomock"
	"github.com/liquanhui-99/gotool/cache/redis_cache/mocks"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestRedisDistributedLock_TryLock(t *testing.T) {
	testCases := []struct {
		name       string
		key        string
		expiration time.Duration
		wantErr    error
		wantLock   *Lock
		client     func(ctrl *gomock.Controller) redis.Cmdable
	}{
		{
			name:       "get lock",
			key:        "get lock",
			expiration: 10 * time.Second,
			wantErr:    nil,
			client: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewBoolResult(true, nil)

				cmd.EXPECT().SetNX(context.Background(), "get lock", gomock.Any(), 10*time.Second).Return(res)
				return cmd
			},
			wantLock: &Lock{
				key: "get lock",
			},
		},
		{
			name:       "lock deadline",
			key:        "lock deadline",
			expiration: 2 * time.Second,
			wantErr:    context.DeadlineExceeded,
			client: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewBoolResult(false, context.DeadlineExceeded)
				cmd.EXPECT().SetNX(context.Background(), "lock deadline", gomock.Any(), 2*time.Second).Return(res)
				return cmd
			},
		},
		{
			name:       "fail to preempt lock",
			key:        "fail to preempt lock",
			expiration: 2 * time.Second,
			wantErr:    ErrFailedToRaceLock,
			client: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewBoolResult(false, ErrFailedToRaceLock)
				cmd.EXPECT().SetNX(context.Background(), "fail to preempt lock", gomock.Any(), 2*time.Second).Return(res)
				return cmd
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			dl := NewRedisDistributedLock(tc.client(ctrl))
			lock, err := dl.TryLock(context.Background(), tc.key, tc.expiration)
			require.Equal(t, err, tc.wantErr)
			if lock != nil {
				assert.Equal(t, lock.key, tc.wantLock.key)
				if lock.val == "" {
					t.Log("锁的唯一标识不存在")
					return
				}
			}
		})
	}
}

func TestRedisDistributedLock_Unlock(t *testing.T) {
	testCases := []struct {
		name     string
		key, val string
		wantErr  error
		client   func(ctrl *gomock.Controller) redis.Cmdable
	}{
		{
			name:    "unlock DeadlineExceeded",
			key:     "unlock DeadlineExceeded",
			val:     "324324",
			wantErr: context.DeadlineExceeded,
			client: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)

				res := redis.NewCmd(context.Background())
				res.SetErr(context.DeadlineExceeded)
				cmd.EXPECT().Eval(context.Background(), unlockScript, []string{"unlock DeadlineExceeded"}, []any{"324324"}).Return(res)

				return cmd
			},
		},
		{
			name:    "unlock failed",
			key:     "unlock failed",
			val:     "32432467",
			wantErr: ErrLockNotHold,
			client: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)

				res := redis.NewCmd(context.Background())
				res.SetVal(int64(0))
				cmd.EXPECT().Eval(context.Background(), unlockScript, []string{"unlock failed"}, []any{"32432467"}).Return(res)

				return cmd
			},
		},
		{
			name:    "unlock success",
			key:     "unlock success",
			val:     "32432467",
			wantErr: nil,
			client: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)

				res := redis.NewCmd(context.Background())
				res.SetVal(int64(1))
				cmd.EXPECT().Eval(context.Background(), unlockScript, []string{"unlock success"}, []any{"32432467"}).Return(res)

				return cmd
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			lock := &Lock{
				key:    tc.key,
				val:    tc.val,
				client: tc.client(ctrl),
			}
			err := lock.Unlock(context.Background())
			assert.Equal(t, err, tc.wantErr)
		})
	}
}

func TestRedisDistributedLock_Refresh(t *testing.T) {
	testCases := []struct {
		name       string
		key, val   string
		wantErr    error
		expiration time.Duration
		client     func(ctrl *gomock.Controller) redis.Cmdable
	}{
		// 刷新超时
		{
			name:    "refresh DeadlineExceeded",
			key:     "refresh DeadlineExceeded",
			val:     "324324",
			wantErr: context.DeadlineExceeded,
			client: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)

				res := redis.NewCmd(context.Background())
				res.SetErr(context.DeadlineExceeded)
				cmd.EXPECT().Eval(context.Background(), refreshScript, []string{"refresh DeadlineExceeded"}, []any{"324324", float64(60)}).Return(res)

				return cmd
			},
			expiration: time.Minute,
		},
		{
			name:    "refresh failed",
			key:     "refresh failed",
			val:     "32432467",
			wantErr: ErrLockNotHold,
			client: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)

				res := redis.NewCmd(context.Background())
				res.SetVal(int64(0))
				cmd.EXPECT().Eval(context.Background(), refreshScript, []string{"refresh failed"},
					[]any{"32432467", float64(60)}).Return(res)

				return cmd
			},
			expiration: time.Minute,
		},
		{
			name:    "refresh success",
			key:     "refresh success",
			val:     "32432467",
			wantErr: nil,
			client: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)

				res := redis.NewCmd(context.Background())
				res.SetVal(int64(1))
				cmd.EXPECT().Eval(context.Background(), refreshScript, []string{"refresh success"}, []any{"32432467", float64(60)}).Return(res)

				return cmd
			},
			expiration: time.Minute,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			lock := &Lock{
				key:        tc.key,
				val:        tc.val,
				client:     tc.client(ctrl),
				expiration: tc.expiration,
			}
			err := lock.Refresh(context.Background())
			assert.Equal(t, err, tc.wantErr)
		})
	}
}

func ExampleLock_Refresh() {
	var l Lock
	errCh := make(chan error, 1)
	closeCh := make(chan struct{}, 1)
	timeoutCh := make(chan struct{}, 1)
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				err := l.Refresh(ctx)
				if err != nil {
					if errors.Is(err, context.DeadlineExceeded) {
						// 超时了，通知重试
						timeoutCh <- struct{}{}
						continue
					} else {
						errCh <- err
					}
				}

				cancel()
			case <-timeoutCh:
				// 续约超时重试的逻辑，这里可以加上一些超时次数等更加精细化的控制
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				err := l.Refresh(ctx)
				if err != nil {
					if errors.Is(err, context.DeadlineExceeded) {
						// 超时了，通知重试
						timeoutCh <- struct{}{}
						continue
					} else {
						errCh <- err
					}
				}
				cancel()
			case <-closeCh:
				// 业务代码通知关闭续约程序
				close(errCh)
				close(closeCh)
				break
			}
		}
	}()

	// 业务逻辑代码
	// 第一部分代码
	select {
	case <-errCh:
	// 处理错误的逻辑
	default:
		// 业务逻辑的部分
	}

	// 第二部分代码
	select {
	case <-errCh:
	// 处理错误的逻辑
	default:
		// 业务逻辑的部分
	}

	// 第三部分代码
	select {
	case <-errCh:
	// 处理错误的逻辑
	default:
		// 业务逻辑的部分
	}

	// ... 每一个部分都需要去监控续约是否有超时错误发生
	// 业务结束，通知停止自动续约
	closeCh <- struct{}{}
	fmt.Println("hello")
	// output:
	// hello
}
