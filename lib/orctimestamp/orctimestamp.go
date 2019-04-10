package orctimestamp

import "time"

const (
	Layout = time.RFC3339Nano
)

func Format(t time.Time) string {
	return t.UTC().Format(Layout)
}

func Parse(s string) (time.Time, error) {
	return time.Parse(Layout, s)
}
