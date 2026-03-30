package main

import (
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

func TestMakeGenerator_BohoBlockDefaultZero(t *testing.T) {
	g := makeGenerator("test")

	if g.Genesis.Config.BohoBlock == nil {
		t.Fatal("BohoBlock should be set by default")
	}
	if g.Genesis.Config.BohoBlock.Int64() != 0 {
		t.Errorf("BohoBlock = %d, want 0", g.Genesis.Config.BohoBlock.Int64())
	}
	// Boho AnzeonConfig should be nil until setBohoConfig is called
	if g.Genesis.Config.Boho != nil {
		t.Error("Boho AnzeonConfig should be nil by default")
	}
}

func TestSetBohoConfig_OverridesDefaultBlock(t *testing.T) {
	g := makeGenerator("test")

	// makeGenerator sets BohoBlock to 0
	if g.Genesis.Config.BohoBlock.Int64() != 0 {
		t.Fatalf("precondition: BohoBlock should be 0, got %d", g.Genesis.Config.BohoBlock.Int64())
	}

	// setBohoConfig should override to the user-specified value
	g.setBohoConfig(500)
	if got := g.Genesis.Config.BohoBlock.Int64(); got != 500 {
		t.Errorf("BohoBlock = %d, want 500", got)
	}
}

func TestSetBohoConfig_BlockNumber(t *testing.T) {
	tests := []struct {
		name      string
		block     int
		wantBlock int64
	}{
		{"genesis activation", 0, 0},
		{"delayed activation", 100, 100},
		{"large block number", 1000000, 1000000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := makeGenerator("test")
			g.setBohoConfig(tt.block)

			if g.Genesis.Config.BohoBlock == nil {
				t.Fatal("BohoBlock should be set")
			}
			if got := g.Genesis.Config.BohoBlock.Int64(); got != tt.wantBlock {
				t.Errorf("BohoBlock = %d, want %d", got, tt.wantBlock)
			}
		})
	}
}

func TestSetBohoConfig_GovMinterV2(t *testing.T) {
	g := makeGenerator("test")
	g.setBohoConfig(0)

	boho := g.Genesis.Config.Boho
	if boho == nil {
		t.Fatal("Boho config should be set")
	}
	if boho.SystemContracts == nil || boho.SystemContracts.GovMinter == nil {
		t.Fatal("Boho GovMinter should not be nil")
	}
	if got := boho.SystemContracts.GovMinter.Address; got != params.DefaultGovMinterAddress {
		t.Errorf("Boho GovMinter address = %v, want %v", got, params.DefaultGovMinterAddress)
	}
	if got := boho.SystemContracts.GovMinter.Version; got != "v2" {
		t.Errorf("Boho GovMinter version = %q, want %q", got, "v2")
	}
}

func TestSetBohoConfig_OnlyContainsGovMinter(t *testing.T) {
	g := makeGenerator("test")
	g.setBohoConfig(0)

	boho := g.Genesis.Config.Boho
	if boho.WBFT != nil {
		t.Error("Boho should not contain WBFT config")
	}
	if boho.Init != nil {
		t.Error("Boho should not contain Init config")
	}
	sc := boho.SystemContracts
	if sc.GovValidator != nil {
		t.Error("Boho should not override GovValidator")
	}
	if sc.NativeCoinAdapter != nil {
		t.Error("Boho should not override NativeCoinAdapter")
	}
	if sc.GovMasterMinter != nil {
		t.Error("Boho should not override GovMasterMinter")
	}
	if sc.GovCouncil != nil {
		t.Error("Boho should not override GovCouncil")
	}
}

func TestSetAnzeonConfigBase_GovMinterV1(t *testing.T) {
	g := makeGenerator("test")

	validators := []common.Address{
		common.HexToAddress("0x1111111111111111111111111111111111111111"),
	}
	blsKeys := []string{"0xaa" + strings.Repeat("bb", 47)}

	g.setAnzeonConfigBase(validators, blsKeys, 1)

	if g.Genesis.Config.Anzeon == nil {
		t.Fatal("Anzeon config should be set")
	}
	if got := g.Genesis.Config.Anzeon.SystemContracts.GovMinter.Version; got != "v1" {
		t.Errorf("Anzeon GovMinter version = %q, want %q", got, "v1")
	}
	// setAnzeonConfigBase should NOT set Boho config (that's setBohoConfig's job)
	if g.Genesis.Config.Boho != nil {
		t.Error("Boho AnzeonConfig should not be set by setAnzeonConfigBase")
	}
}
