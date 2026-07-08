package backfill

import "testing"

func TestMongoDatabaseNameUsesConfiguredDatabase(t *testing.T) {
	got, err := mongoDatabaseName("mongodb://localhost:27017/from-uri", "from-config")
	if err != nil {
		t.Fatalf("mongoDatabaseName() error = %v", err)
	}
	if got != "from-config" {
		t.Fatalf("mongoDatabaseName() = %q", got)
	}
}

func TestMongoDatabaseNameUsesURIPath(t *testing.T) {
	got, err := mongoDatabaseName("mongodb://localhost:27017/from-uri?retryWrites=true", "")
	if err != nil {
		t.Fatalf("mongoDatabaseName() error = %v", err)
	}
	if got != "from-uri" {
		t.Fatalf("mongoDatabaseName() = %q", got)
	}
}

func TestMongoDatabaseNameRequiresDatabase(t *testing.T) {
	_, err := mongoDatabaseName("mongodb://localhost:27017", "")
	if err == nil {
		t.Fatal("expected error")
	}
}
