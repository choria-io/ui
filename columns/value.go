// Copyright (c) 2026, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package columns

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/mattn/go-runewidth"
)

// Value is a formatted value with both a display form and a raw form. The
// display form is used by the text and Markdown renderers; the raw form is used
// by the JSON and YAML renderers so that, for example, a byte count serializes
// as a number rather than the string "1.5 KiB".
//
// Value is produced by the helper functions in this package (IBytes, Duration,
// Map, Style, Annotated and so on). It cannot be implemented outside the
// package; pass a plain value to Item, Values or Fields to have it formatted
// automatically by Format.
type Value interface {
	// String returns the value's display form: the terminal-styled text for a
	// Style value and the formatted text for the others, with a multi-line value
	// joining its lines with newlines. It lets a value such as
	// Style("{bold}OK{/bold}") be used as a plain string and satisfies
	// fmt.Stringer.
	String() string

	lines(o *options) []string
	raw() any
}

// stringOf returns the display form of a value, joining a multi-line value with
// newlines. It backs the exported String methods.
func stringOf(v Value) string {
	return strings.Join(v.lines(&options{}), "\n")
}

// toValue wraps a plain value for formatting, or returns it unchanged if it is
// already a Value.
func toValue(v any) Value {
	if vv, ok := v.(Value); ok {
		return vv
	}

	return autoValue{v: v}
}

// text resolves the plain single-line display form of a value, used by the
// composition helpers. Value display forms do not depend on margins or indents,
// so a zero options value is sufficient; the plain flag keeps any styling out of
// composed strings.
func text(v any) string {
	return strings.Join(toValue(v).lines(&options{plain: true}), " ")
}

// autoValue formats an arbitrary value using Format.
type autoValue struct {
	v any
}

func (a autoValue) String() string {
	return stringOf(a)
}

func (a autoValue) lines(o *options) []string {
	return valueLines(Format(a.v))
}

func (a autoValue) raw() any {
	if err, ok := a.v.(error); ok {
		return err.Error()
	}

	return a.v
}

// fmtValue is a value with a fixed display string and an explicit raw form. It
// backs most of the helper functions.
type fmtValue struct {
	display string
	rawv    any
}

func (f fmtValue) String() string {
	return stringOf(f)
}

func (f fmtValue) lines(o *options) []string {
	return valueLines(f.display)
}

func (f fmtValue) raw() any {
	return f.rawv
}

// emptyValue renders as empty and carries no raw value. It is returned by
// helpers that cannot produce a result, such as Percent with a zero divisor.
type emptyValue struct{}

func (emptyValue) String() string {
	return ""
}

func (emptyValue) lines(o *options) []string {
	return []string{""}
}

func (emptyValue) raw() any {
	return nil
}

// Format renders a value the way the automatic formatter does: strings as is,
// numbers with thousands separators, durations and times in human form, byte
// slices and []string joined, and everything else via fmt. It is exported so the
// same formatting can be reused outside a Document.
func Format(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case []string:
		return strings.Join(x, ", ")
	case error:
		return x.Error()
	case time.Duration:
		return HumanizeDuration(x)
	case time.Time:
		return x.Local().Format("2006-01-02 15:04:05")
	case bool:
		return strconv.FormatBool(x)
	case uint:
		return humanize.Comma(int64(x))
	case uint8:
		return humanize.Comma(int64(x))
	case uint16:
		return humanize.Comma(int64(x))
	case uint32:
		return humanize.Comma(int64(x))
	case uint64:
		if x >= math.MaxInt64 {
			return strconv.FormatUint(x, 10)
		}
		return humanize.Comma(int64(x))
	case int:
		return humanize.Comma(int64(x))
	case int8:
		return humanize.Comma(int64(x))
	case int16:
		return humanize.Comma(int64(x))
	case int32:
		return humanize.Comma(int64(x))
	case int64:
		return humanize.Comma(x)
	case float32:
		return humanize.CommafWithDigits(float64(x), 3)
	case float64:
		return humanize.CommafWithDigits(x, 3)
	case *big.Int:
		return humanize.BigComma(x)
	case fmt.Stringer:
		return x.String()
	default:
		return fmt.Sprintf("%v", x)
	}
}

// Plain formats a value without thousands separators, for identifiers such as
// ports, PIDs and years that should not be grouped. The raw form is the value.
func Plain(v any) Value {
	var s string
	switch x := v.(type) {
	case nil:
		s = ""
	case string:
		s = x
	default:
		s = fmt.Sprintf("%v", x)
	}

	return fmtValue{display: s, rawv: v}
}

// IBytes formats a byte count using IEC units (1024 based), for example
// "1.5 KiB". The raw form is the byte count.
func IBytes(bytes int64) Value {
	if bytes < 0 {
		return fmtValue{display: "-" + humanize.IBytes(uint64(-bytes)), rawv: bytes}
	}

	return fmtValue{display: humanize.IBytes(uint64(bytes)), rawv: bytes}
}

// Bytes formats a byte count using SI units (1000 based), for example "1.5 kB".
// The raw form is the byte count.
func Bytes(bytes int64) Value {
	if bytes < 0 {
		return fmtValue{display: "-" + humanize.Bytes(uint64(-bytes)), rawv: bytes}
	}

	return fmtValue{display: humanize.Bytes(uint64(bytes)), rawv: bytes}
}

// Duration formats a duration using HumanizeDuration, for example "1h30m0s" or
// "never" for the maximum duration. The raw form is the duration as a string
// parseable by time.ParseDuration.
func Duration(d time.Duration) Value {
	return fmtValue{display: HumanizeDuration(d), rawv: d.String()}
}

// Time formats a time relative to now, for example "3 minutes ago". Because the
// display form depends on the current time it is unsuitable for golden output;
// use DateTime for a stable rendering. The raw form is the time in RFC3339.
func Time(t time.Time) Value {
	return fmtValue{display: humanize.Time(t), rawv: t.UTC().Format(time.RFC3339)}
}

// DateTime formats a time using the given layout, defaulting to the local
// "2006-01-02 15:04:05". The raw form is the time in RFC3339.
func DateTime(t time.Time, layout ...string) Value {
	l := "2006-01-02 15:04:05"
	if len(layout) > 0 && layout[0] != "" {
		l = layout[0]
	}

	return fmtValue{display: t.Local().Format(l), rawv: t.UTC().Format(time.RFC3339)}
}

// Percent formats part as a percentage of whole, for example "25%". A zero whole
// yields an empty value rather than a division by zero. The raw form is the
// percentage as a number.
func Percent(part, whole float64) Value {
	if whole == 0 {
		return emptyValue{}
	}

	p := part / whole * 100

	return fmtValue{display: formatPercent(p), rawv: p}
}

// kvPair is one entry of a Map value, with the key already reduced to its string
// form and the value kept for both display and raw output.
type kvPair struct {
	key string
	val Value
}

// mapValue renders a map as a block of "key: value" lines occupying the value
// column of a single row.
type mapValue struct {
	pairs []kvPair
}

// Map renders a map as a block of aligned "key: value" lines under a single
// description, for example via Item("Labels", columns.Map(m)). Keys are sorted
// by their string form. The raw form is an ordered object.
func Map[K comparable, V any](m map[K]V) Value {
	pairs := make([]kvPair, 0, len(m))
	for k, v := range m {
		pairs = append(pairs, kvPair{key: fmt.Sprint(k), val: toValue(v)})
	}
	sortPairs(pairs)

	return mapValue{pairs: pairs}
}

func (m mapValue) String() string {
	return stringOf(m)
}

func (m mapValue) lines(o *options) []string {
	if len(m.pairs) == 0 {
		return []string{""}
	}

	maxKey := 0
	for _, p := range m.pairs {
		w := runewidth.StringWidth(p.key)
		if w > maxKey {
			maxKey = w
		}
	}

	out := make([]string, 0, len(m.pairs))
	for _, p := range m.pairs {
		val := strings.Join(p.val.lines(o), " ")
		out = append(out, padRight(p.key+":", maxKey+1)+" "+val)
	}

	return out
}

func (m mapValue) raw() any {
	om := newOmap()
	for _, p := range m.pairs {
		om.set(p.key, p.val.raw())
	}

	return om
}

func sortPairs(pairs []kvPair) {
	for i := 1; i < len(pairs); i++ {
		for j := i; j > 0 && pairs[j-1].key > pairs[j].key; j-- {
			pairs[j-1], pairs[j] = pairs[j], pairs[j-1]
		}
	}
}

func formatPercent(p float64) string {
	s := strconv.FormatFloat(p, 'f', 1, 64)
	s = strings.TrimSuffix(s, ".0")

	return s + "%"
}

// HumanizeDuration renders a duration compactly, for example "1y2d3h4m5s",
// "1h30m0s" or "1.50s", returning "never" for the maximum duration. Sub-second
// values keep go's native rounded form.
func HumanizeDuration(d time.Duration) string {
	if d == math.MaxInt64 {
		return "never"
	}

	if d < time.Millisecond {
		return d.Round(time.Microsecond).String()
	}

	if d < time.Second {
		return d.Round(time.Millisecond).String()
	}

	tsecs := d / time.Second
	tmins := tsecs / 60
	thrs := tmins / 60
	tdays := thrs / 24
	tyrs := tdays / 365

	if tyrs > 0 {
		return fmt.Sprintf("%dy%dd%dh%dm%ds", tyrs, tdays%365, thrs%24, tmins%60, tsecs%60)
	}

	if tdays > 0 {
		return fmt.Sprintf("%dd%dh%dm%ds", tdays, thrs%24, tmins%60, tsecs%60)
	}

	if thrs > 0 {
		return fmt.Sprintf("%dh%dm%ds", thrs, tmins%60, tsecs%60)
	}

	if tmins > 0 {
		return fmt.Sprintf("%dm%ds", tmins, tsecs%60)
	}

	return fmt.Sprintf("%.2fs", d.Seconds())
}
