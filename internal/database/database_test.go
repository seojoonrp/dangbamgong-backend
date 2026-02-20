package database

import (
	"context"
	"testing"
)

func TestNew(t *testing.T) {
	db := New()
	if db == nil {
		t.Fatal("New() returned nil")
	}
}

func TestPing(t *testing.T) {
	db := New()

	err := db.Client().Ping(context.Background(), nil)
	if err != nil {
		t.Fatalf("expected ping to succeed, got %v", err)
	}
}
