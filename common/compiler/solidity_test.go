// Copyright 2019 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package compiler

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testSource = `
pragma solidity >0.0.0;
contract test {
   /// @notice Will multiply ` + "`a`" + ` by 7.
   function multiply(uint a) public returns(uint d) {
       return a * 7;
   }
}
`
)

func skipWithoutSolc(t *testing.T) {
	if _, err := exec.LookPath("solc"); err != nil {
		t.Skip(err)
	}
}

func TestExtractSourceVersion(t *testing.T) {
	tests := []struct {
		source   string
		expected []string
	}{
		{"pragma solidity ^0.4.24;\n", []string{"^0.4.24"}},
		{"pragma solidity 0.4.24;\n", []string{"0.4.24"}},
		{"pragma solidity ^0.4.24;\n", []string{"^0.4.24"}},
		{"pragma solidity 0.4.24;\npragma solidity 0.4.25;\n", []string{"0.4.24", "0.4.25"}},
		{"pragma solidity 0.4.24;\n//pragma solidity 0.4.25;\n", []string{"0.4.24"}},
		{"//pragma solidity 0.4.24;\npragma solidity 0.4.25;\n", []string{"0.4.25"}},
	}
	for _, test := range tests {
		v := extractSourceVersion(test.source)
		assert.Equal(t, test.expected, v)
	}
}

func TestSolidityCompiler(t *testing.T) {
	skipWithoutSolc(t)

	contracts, err := CompileSolidityString("", testSource)
	if err != nil {
		t.Fatalf("error compiling source. result %v: %v", contracts, err)
	}
	if len(contracts) != 1 {
		t.Errorf("one contract expected, got %d", len(contracts))
	}
	c, ok := contracts["test"]
	if !ok {
		c, ok = contracts["<stdin>:test"]
		if !ok {
			t.Fatal("info for contract 'test' not present in result")
		}
	}
	if c.Code == "" {
		t.Error("empty code")
	}
	if c.Info.Source != testSource {
		t.Error("wrong source")
	}
	if c.Info.CompilerVersion == "" {
		t.Error("empty version")
	}
}

func TestSolidityCompileError(t *testing.T) {
	skipWithoutSolc(t)

	// Force syntax error by removing some characters.
	contracts, err := CompileSolidityString("", testSource[4:])
	if err == nil {
		t.Errorf("error expected compiling source. got none. result %v", contracts)
	}
	t.Logf("error: %v", err)
}

func TestSolcCanCompile(t *testing.T) {
	solcVersion := "0.8.11"
	tests := []struct {
		version  string
		expected bool
	}{
		{"^0.8.11", true},
		{"^0.4.24", false},
		{"^0.5.6", false},
		{"0.5.6", false},
	}

	for _, test := range tests {
		r, err := solcCanCompile(solcVersion, []string{test.version})
		assert.Equal(t, test.expected, r)
		assert.Nil(t, err)
	}
}

func TestSolidityCompileVersions(t *testing.T) {
	// Usually only one version of solc is installed, if any.
	// But CompileSolidityOrLoad will work in all cases.

	versions := []string{"0.4.24", "0.8.11"}
	for _, version := range versions {
		t.Logf("testing version %s", version)

		path := "../../contracts/compiler/version_" + version + ".sol"

		contracts, err := CompileSolidityOrLoad("", path)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(contracts))

		for _, contract := range contracts {
			assert.NotNil(t, contract.Code)
			assert.NotNil(t, contract.Info.AbiDefinition)
			assert.Equal(t, version, contract.Info.CompilerVersion)
		}
	}
}
