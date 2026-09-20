package rwa

import "testing"

func TestRwaExecution(t *testing.T) {
	module := NewRwaModule()
	result := module.Execute()
	expected := "Phase 15: RWA Platform Launch executed successfully"

	if result != expected {
		t.Fatalf("Expected %s, got %s", expected, result)
	}
	t.Log("Passed: End-to-end cycle completed for Phase 15")
}

func TestRwaAsset_Lifecycle(t *testing.T) {
	module := NewRwaModule()

	// Error cases on invalid asset parameters
	if _, err := module.TokenizeAsset("", "Gold Vault", 1000, 50_000_00); err == nil {
		t.Fatal("expected error on empty ID")
	}
	if _, err := module.TokenizeAsset("RWA-1", "", 1000, 50_000_00); err == nil {
		t.Fatal("expected error on empty name")
	}
	if _, err := module.TokenizeAsset("RWA-1", "Gold Vault", 0, 50_000_00); err == nil {
		t.Fatal("expected error on zero units")
	}

	// Valid asset creation
	asset, err := module.TokenizeAsset("RWA-101", "Treasury Bill Yield", 10_000, 1_000_000_00)
	if err != nil {
		t.Fatalf("TokenizeAsset failed: %v", err)
	}
	if asset.TransparentUnits() != 10_000 {
		t.Fatalf("expected 10000 transparent units, got %d", asset.TransparentUnits())
	}

	// Shield units
	if err := asset.ShieldUnits(3000); err != nil {
		t.Fatalf("ShieldUnits failed: %v", err)
	}
	if asset.ShieldedUnits != 3000 {
		t.Fatalf("expected 3000 shielded units, got %d", asset.ShieldedUnits)
	}
	if asset.TransparentUnits() != 7000 {
		t.Fatalf("expected 7000 transparent units, got %d", asset.TransparentUnits())
	}

	// Shield too much error
	if err := asset.ShieldUnits(8000); err == nil {
		t.Fatal("expected error shielding more units than total")
	}

	// Unshield units
	if err := asset.UnshieldUnits(1000); err != nil {
		t.Fatalf("UnshieldUnits failed: %v", err)
	}
	if asset.ShieldedUnits != 2000 {
		t.Fatalf("expected 2000 shielded units, got %d", asset.ShieldedUnits)
	}
	if asset.TransparentUnits() != 8000 {
		t.Fatalf("expected 8000 transparent units, got %d", asset.TransparentUnits())
	}

	// Unshield too much error
	if err := asset.UnshieldUnits(5000); err == nil {
		t.Fatal("expected error unshielding more units than shielded")
	}
}
