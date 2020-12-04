package engine_test

import (
	"context"
	"errors"
	"syscall"
	"testing"

	"github.com/apex/log"
	"github.com/google/go-cmp/cmp"
	engine "github.com/ooni/probe-engine"
	"github.com/ooni/probe-engine/internal/kvstore"
	"github.com/ooni/probe-engine/model"
)

func TestInputLoaderInputNoneWithStaticInputs(t *testing.T) {
	il := engine.NewInputLoader(engine.InputLoaderConfig{
		StaticInputs: []string{"https://www.google.com/"},
		InputPolicy:  engine.InputNone,
	})
	ctx := context.Background()
	out, err := il.Load(ctx)
	if !errors.Is(err, engine.ErrNoInputExpected) {
		t.Fatalf("not the error we expected: %+v", err)
	}
	if out != nil {
		t.Fatal("not the output we expected")
	}
}

func TestInputLoaderInputNoneWithFilesInputs(t *testing.T) {
	il := engine.NewInputLoader(engine.InputLoaderConfig{
		SourceFiles: []string{
			"testdata/inputloader1.txt",
			"testdata/inputloader2.txt",
		},
		InputPolicy: engine.InputNone,
	})
	ctx := context.Background()
	out, err := il.Load(ctx)
	if !errors.Is(err, engine.ErrNoInputExpected) {
		t.Fatalf("not the error we expected: %+v", err)
	}
	if out != nil {
		t.Fatal("not the output we expected")
	}
}

func TestInputLoaderInputNoneWithBothInputs(t *testing.T) {
	il := engine.NewInputLoader(engine.InputLoaderConfig{
		StaticInputs: []string{"https://www.google.com/"},
		SourceFiles: []string{
			"testdata/inputloader1.txt",
			"testdata/inputloader2.txt",
		},
		InputPolicy: engine.InputNone,
	})
	ctx := context.Background()
	out, err := il.Load(ctx)
	if !errors.Is(err, engine.ErrNoInputExpected) {
		t.Fatalf("not the error we expected: %+v", err)
	}
	if out != nil {
		t.Fatal("not the output we expected")
	}
}

func TestInputLoaderInputNoneWithNoInput(t *testing.T) {
	il := engine.NewInputLoader(engine.InputLoaderConfig{
		InputPolicy: engine.InputNone,
	})
	ctx := context.Background()
	out, err := il.Load(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0].URL != "" {
		t.Fatal("not the output we expected")
	}
}

func TestInputLoaderInputOptionalWithNoInput(t *testing.T) {
	il := engine.NewInputLoader(engine.InputLoaderConfig{
		InputPolicy: engine.InputOptional,
	})
	ctx := context.Background()
	out, err := il.Load(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0].URL != "" {
		t.Fatal("not the output we expected")
	}
}

func TestInputLoaderInputOptionalWithInput(t *testing.T) {
	il := engine.NewInputLoader(engine.InputLoaderConfig{
		StaticInputs: []string{"https://www.google.com/"},
		SourceFiles: []string{
			"testdata/inputloader1.txt",
			"testdata/inputloader2.txt",
		},
		InputPolicy: engine.InputOptional,
	})
	ctx := context.Background()
	out, err := il.Load(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 5 {
		t.Fatal("not the output length we expected")
	}
	expect := []model.URLInfo{
		{URL: "https://www.google.com/"},
		{URL: "https://www.x.org/"},
		{URL: "https://www.slashdot.org/"},
		{URL: "https://abc.xyz/"},
		{URL: "https://run.ooni.io/"},
	}
	if diff := cmp.Diff(out, expect); diff != "" {
		t.Fatal(diff)
	}
}

func TestInputLoaderInputOptionalNonexistentFile(t *testing.T) {
	il := engine.NewInputLoader(engine.InputLoaderConfig{
		StaticInputs: []string{"https://www.google.com/"},
		SourceFiles: []string{
			"testdata/inputloader1.txt",
			"/nonexistent",
			"testdata/inputloader2.txt",
		},
		InputPolicy: engine.InputOptional,
	})
	ctx := context.Background()
	out, err := il.Load(ctx)
	if !errors.Is(err, syscall.ENOENT) {
		t.Fatalf("not the error we expected: %+v", err)
	}
	if out != nil {
		t.Fatal("not the output we expected")
	}
}

func TestInputLoaderInputRequiredlWithInput(t *testing.T) {
	il := engine.NewInputLoader(engine.InputLoaderConfig{
		StaticInputs: []string{"https://www.google.com/"},
		SourceFiles: []string{
			"testdata/inputloader1.txt",
			"testdata/inputloader2.txt",
		},
		InputPolicy: engine.InputRequired,
	})
	ctx := context.Background()
	out, err := il.Load(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 5 {
		t.Fatal("not the output length we expected")
	}
	expect := []model.URLInfo{
		{URL: "https://www.google.com/"},
		{URL: "https://www.x.org/"},
		{URL: "https://www.slashdot.org/"},
		{URL: "https://abc.xyz/"},
		{URL: "https://run.ooni.io/"},
	}
	if diff := cmp.Diff(out, expect); diff != "" {
		t.Fatal(diff)
	}
}

func TestInputLoaderInputRequiredlWithNoInputAndCancelledContext(t *testing.T) {
	sess, err := engine.NewSession(engine.SessionConfig{
		AssetsDir:       "testdata",
		KVStore:         kvstore.NewMemoryKeyValueStore(),
		Logger:          log.Log,
		SoftwareName:    "miniooni",
		SoftwareVersion: "0.1.0-dev",
		TempDir:         "testdata",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()
	il := engine.NewInputLoader(engine.InputLoaderConfig{
		InputPolicy: engine.InputRequired,
		Session:     sess,
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // fail immediately
	out, err := il.Load(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("not the error we expected: %+v", err)
	}
	if out != nil {
		t.Fatal("not the output we expected")
	}
}

func TestInputLoaderInputRequiredlWithNoInputGood(t *testing.T) {
	if testing.Short() {
		t.Skip("skip test in short mode")
	}
	sess, err := engine.NewSession(engine.SessionConfig{
		AvailableProbeServices: []model.Service{{
			Address: "https://ams-pg-test.ooni.org/",
			Type:    "https",
		}},
		AssetsDir:       "testdata",
		KVStore:         kvstore.NewMemoryKeyValueStore(),
		Logger:          log.Log,
		SoftwareName:    "miniooni",
		SoftwareVersion: "0.1.0-dev",
		TempDir:         "testdata",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()
	il := engine.NewInputLoader(engine.InputLoaderConfig{
		InputPolicy: engine.InputRequired,
		Session:     sess,
		URLLimit:    30,
	})
	ctx := context.Background()
	out, err := il.Load(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) < 10 {
		t.Fatal("not the output length we expected")
	}
}
