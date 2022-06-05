// Copyright 2021 The Cockroach Authors.
//
// Use of this software is governed
// by the Apache License, Version 2.0, included in the file
// LICENSE.md

//go:build tools
// +build tools

package main

import (
	_ "github.com/cockroachdb/crlfmt"
	_ "golang.org/x/lint/golint"
	_ "golang.org/x/tools/cmd/stringer"
	_ "honnef.co/go/tools/cmd/staticcheck"
)
