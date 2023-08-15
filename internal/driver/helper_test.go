// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSlicesAreEqual(t *testing.T) {
	tests := []struct {
		name   string
		a      any
		b      any
		expect bool
	}{
		{
			name:   "both nil",
			a:      nil,
			b:      nil,
			expect: true,
		},
		{
			name:   "b nil",
			a:      []string{"a"},
			b:      nil,
			expect: false,
		},
		{
			name:   "a nil",
			a:      nil,
			b:      []string{"b"},
			expect: false,
		},
		{
			name:   "a nil",
			a:      nil,
			b:      []string{"b"},
			expect: false,
		},
		{
			name:   "different items",
			a:      []string{"a", "1"},
			b:      []string{"b", "2"},
			expect: false,
		},
		{
			name:   "same items, same order",
			a:      []string{"test", "me"},
			b:      []string{"test", "me"},
			expect: true,
		},
		{
			name:   "same items, different order",
			a:      []string{"me", "test"},
			b:      []string{"test", "me"},
			expect: false,
		},
		{
			name:   "different amount",
			a:      []string{"1", "2", "3"},
			b:      []string{"1", "2", "3", "4"},
			expect: false,
		},
		{
			name:   "wrong type",
			a:      []int{1, 2, 3},
			b:      []int{1, 2, 3},
			expect: false,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			actual := slicesAreEqual(logger.MockLogger{}, test.a, test.b)
			assert.Equal(t, test.expect, actual)
		})
	}
}
