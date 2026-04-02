package rules_test

import (
	"path/filepath"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/rules"
)

func TestCheckTypePatterns_Valid(t *testing.T) {
	root := t.TempDir()
	module := "example.com/tp-valid"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")

	writeTestFile(t, filepath.Join(root, "internal", "worker", "worker_order.go"),
		`package worker

type OrderWorker struct{}

func (w *OrderWorker) Process() error { return nil }
`)
	writeTestFile(t, filepath.Join(root, "internal", "worker", "worker_payment.go"),
		`package worker

type PaymentWorker struct{}

func (w *PaymentWorker) Process() error { return nil }
`)

	pkgs := loadTestPackages(t, root)
	vs := rules.CheckTypePatterns(pkgs, rules.WithModel(rules.ConsumerWorker()))

	if len(vs) != 0 {
		t.Errorf("expected 0 violations, got %d: %v", len(vs), vs)
	}
}

func TestCheckTypePatterns_MissingType(t *testing.T) {
	root := t.TempDir()
	module := "example.com/tp-missing-type"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")

	writeTestFile(t, filepath.Join(root, "internal", "worker", "worker_order.go"),
		`package worker

type SomethingElse struct{}
`)

	pkgs := loadTestPackages(t, root)
	vs := rules.CheckTypePatterns(pkgs, rules.WithModel(rules.ConsumerWorker()))

	if len(vs) != 1 {
		t.Fatalf("expected 1 violation, got %d: %v", len(vs), vs)
	}
	if vs[0].Rule != "naming.worker-type-mismatch" {
		t.Errorf("rule = %q, want naming.worker-type-mismatch", vs[0].Rule)
	}
}

func TestCheckTypePatterns_MissingMethod(t *testing.T) {
	root := t.TempDir()
	module := "example.com/tp-missing-method"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")

	writeTestFile(t, filepath.Join(root, "internal", "worker", "worker_order.go"),
		`package worker

type OrderWorker struct{}
`)

	pkgs := loadTestPackages(t, root)
	vs := rules.CheckTypePatterns(pkgs, rules.WithModel(rules.ConsumerWorker()))

	if len(vs) != 1 {
		t.Fatalf("expected 1 violation, got %d: %v", len(vs), vs)
	}
	if vs[0].Rule != "naming.worker-missing-process" {
		t.Errorf("rule = %q, want naming.worker-missing-process", vs[0].Rule)
	}
}

func TestCheckTypePatterns_NonPrefixedFileIgnored(t *testing.T) {
	root := t.TempDir()
	module := "example.com/tp-nonprefix"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")

	writeTestFile(t, filepath.Join(root, "internal", "worker", "helper.go"),
		`package worker

type Helper struct{}
`)

	pkgs := loadTestPackages(t, root)
	vs := rules.CheckTypePatterns(pkgs, rules.WithModel(rules.ConsumerWorker()))

	if len(vs) != 0 {
		t.Errorf("expected 0 violations, got %d: %v", len(vs), vs)
	}
}
