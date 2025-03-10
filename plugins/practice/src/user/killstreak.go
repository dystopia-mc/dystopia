package user

import "sync/atomic"

type KillStreak struct {
	current atomic.Int64
	max     atomic.Int64
}

func NewKillStreak(current, max int64) *KillStreak {
	k := &KillStreak{}
	k.max.Store(max)
	k.current.Store(current)

	return k
}

func (k *KillStreak) Current() int64 {
	return k.current.Load()
}

func (k *KillStreak) Max() int64 {
	return k.max.Load()
}

func (k *KillStreak) ResetCurrent() {
	k.current.Store(0)
}

func (k *KillStreak) Kill() {
	k.current.Add(1)
	after := k.current.Load()

	if k.Max() < after {
		k.max.Store(after)
	}
}
