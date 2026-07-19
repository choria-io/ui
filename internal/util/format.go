// Copyright (c) 2026, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

// Package util holds formatting, sanitization and environment helpers shared by
// the columns and table packages so neither has to depend on the other.
package util

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
)

// Format renders a value the way the automatic formatter does: strings as is,
// numbers with thousands separators, durations and times in human form, byte
// slices and []string joined, and everything else via fmt.
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
