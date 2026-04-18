package rules

import (
	"strings"
	"testing"
)

func TestSignatureTypeEngine_ParamLeak(t *testing.T) {
	pkgs := loadTxBoundaryInternal(t)
	got := checkTypeInSignature(pkgs,
		"github.com/kimtaeyun/testproject-txboundary",
		"../testdata/txboundary",
		DDD(), NewConfig(),
		[]string{"database/sql.Tx"},
		[]string{"app"},
		"tx.type-in-signature",
		"tx type %q must not appear in signature outside %v",
		"keep %q confined to %v",
	)
	if len(got) < 2 {
		t.Fatalf("expected >=2 violations, got %d: %+v", len(got), got)
	}
	var sawRepo, sawSvc bool
	for _, v := range got {
		if strings.Contains(v.File, "core/repo/repository.go") {
			sawRepo = true
		}
		if strings.Contains(v.File, "core/svc/service.go") {
			sawSvc = true
		}
		if v.Rule != "tx.type-in-signature" {
			t.Errorf("unexpected rule: %s", v.Rule)
		}
	}
	if !sawRepo {
		t.Error("expected repo param violation")
	}
	if !sawSvc {
		t.Error("expected svc return violation")
	}
}

func TestSignatureTypeEngine_AllowedLayerIgnored(t *testing.T) {
	pkgs := loadTxBoundaryInternal(t)
	got := checkTypeInSignature(pkgs,
		"github.com/kimtaeyun/testproject-txboundary",
		"../testdata/txboundary",
		DDD(), NewConfig(),
		[]string{"database/sql.Tx"},
		[]string{"app", "core/repo", "core/svc"},
		"tx.type-in-signature",
		"%q %v",
		"%q %v",
	)
	if len(got) != 0 {
		t.Fatalf("expected 0 violations, got %d", len(got))
	}
}

func TestSignatureTypeEngine_StripsWrappers(t *testing.T) {
	pkgs := loadTxBoundaryInternal(t)
	got := checkTypeInSignature(pkgs,
		"github.com/kimtaeyun/testproject-txboundary",
		"../testdata/txboundary",
		DDD(), NewConfig(),
		[]string{"database/sql.Tx"},
		[]string{"app"},
		"tx.type-in-signature", "%q %v", "%q %v")
	if len(got) < 2 {
		t.Fatalf("expected >=2 violations (pointer wrapper), got %d", len(got))
	}
}

func TestSignatureTypeEngine_EmptyTypesNoop(t *testing.T) {
	pkgs := loadTxBoundaryInternal(t)
	got := checkTypeInSignature(pkgs,
		"github.com/kimtaeyun/testproject-txboundary",
		"../testdata/txboundary",
		DDD(), NewConfig(),
		nil,
		[]string{"app"},
		"tx.type-in-signature", "%q %v", "%q %v")
	if got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}
}
