package internal

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/world"
	kit2 "github.com/k4ties/dystopia/plugins/practice/src/kit"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"math"
	"net"
	"runtime"
	"strings"
	"time"
)

type IntersectThresholdSource struct{}

func (i IntersectThresholdSource) ReducedByArmour() bool     { return false }
func (i IntersectThresholdSource) ReducedByResistance() bool { return false }
func (i IntersectThresholdSource) Fire() bool                { return false }

const (
	DeadMessageBlankFormat = "<red>%s</red> was slain by <red>%s</red>"
	DeadMessageCountFormat = "<red>%s</red> [%d %s] was slain by <red>%s</red> [%d %s]"
)

const (
	DeadMessageBlank = iota
	DeadMessageCount
)

func FormatDeadMessage(index int, args ...any) string {
	switch index {
	case DeadMessageBlank:
		if len(args) < 2 {
			panic("not enough arguments")
		}
		return text.Colourf(DeadMessageBlankFormat, args...)
	case DeadMessageCount:
		if len(args) < 6 {
			panic("not enough arguments")
		}
		return text.Colourf(DeadMessageCountFormat, args...)
	}

	panic("should never happen")
}

func KitIncludes[T world.Item](k kit2.Kit) bool {
	for _, i := range k.Items() {
		if _, isPot := i.Item().(T); isPot {
			return true
		}
	}

	return false
}

func Count[T world.Item](inv *inventory.Inventory) int {
	var count = 0

	for _, i := range inv.Items() {
		if _, ok := i.Item().(T); ok {
			count += i.Count()
		}
	}

	return count
}

func Ip(n net.Addr) string {
	return strings.Split(n.String(), ":")[0]
}

func IpStr(n string) string {
	return strings.Split(n, ":")[0]
}

func DurationString(d time.Duration) string {
	round := func(v float64) int {
		return int(math.Floor(v))
	}
	if round(d.Seconds()) == -9223372037 { // time.Until(time.Time{})
		return "forever"
	}

	d += time.Second
	days := d.Hours() / 24
	years := days / 365

	if years >= 1 {
		return fmt.Sprintf("%d years", round(years))
	}
	if d.Minutes() < 1 && years < 1 {
		return fmt.Sprintf("%d seconds", round(d.Seconds()))
	}
	if d.Minutes() >= 1 && d.Hours() < 1 {
		return fmt.Sprintf("%d minutes", round(d.Minutes()))
	}
	if d.Hours() >= 1 && days < 1 {
		return fmt.Sprintf("%d hours", round(d.Hours()))
	}
	return fmt.Sprintf("%d days", round(days))
}

func GetCallers(n int) []string {
	var lines []string

	for i := 0; i < n; i++ {
		_, file, line, ok := runtime.Caller(i + 1)
		if !ok {
			break
		}

		lines = append(lines, fmt.Sprintf("%s, line: %d", file, line))
	}

	return lines
}
