// Copyright (c) 2026, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestUtil(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Util")
}

var _ = Describe("TerminalWidth", func() {
	BeforeEach(func() {
		orig := stdoutWidth
		stdoutWidth = func() int { return 120 }
		DeferCleanup(func() { stdoutWidth = orig })
	})

	It("prefers a positive COLUMNS", func() {
		GinkgoT().Setenv("COLUMNS", " 80 ")
		Expect(TerminalWidth()).To(Equal(80))
	})

	It("falls back to the stdout terminal when COLUMNS is not a positive integer", func() {
		for _, v := range []string{"", "abc", "0", "-5"} {
			GinkgoT().Setenv("COLUMNS", v)
			Expect(TerminalWidth()).To(Equal(120), "COLUMNS=%q", v)
		}
	})

	It("uses the stdout terminal when COLUMNS is unset", func() {
		GinkgoT().Setenv("COLUMNS", "")
		Expect(os.Unsetenv("COLUMNS")).To(Succeed())
		Expect(TerminalWidth()).To(Equal(120))
	})

	It("returns 0 when stdout is not a terminal and COLUMNS is unset", func() {
		GinkgoT().Setenv("COLUMNS", "")
		Expect(os.Unsetenv("COLUMNS")).To(Succeed())
		stdoutWidth = func() int { return 0 }
		Expect(TerminalWidth()).To(Equal(0))
	})
})
