package kv

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestDB(t *testing.T) (*Store, func()) {
	tmpDir, err := os.MkdirTemp("", "migraine-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	db, err := InitDB(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to init test DB: %v", err)
	}

	store := New(db)

	cleanup := func() {
		db.Close()
		os.RemoveAll(tmpDir)
	}

	return store, cleanup
}

func TestStore_SetGet(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	type testStruct struct {
		Name string
		Age  int
	}

	tests := []struct {
		name    string
		key     string
		value   testStruct
		wantErr bool
	}{
		{
			name: "valid set/get",
			key:  "test_key",
			value: testStruct{
				Name: "John",
				Age:  30,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test Set
			if err := store.Set(tt.key, tt.value); (err != nil) != tt.wantErr {
				t.Errorf("Store.Set() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Test Get
			var got testStruct
			if err := store.Get(tt.key, &got); (err != nil) != tt.wantErr {
				t.Errorf("Store.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && (got.Name != tt.value.Name || got.Age != tt.value.Age) {
				t.Errorf("Store.Get() = %v, want %v", got, tt.value)
			}
		})
	}
}
