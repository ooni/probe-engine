package kvstore_test

import (
	"errors"
	"fmt"
	"log"
	"reflect"

	"github.com/ooni/probe-engine/pkg/kvstore"
)

func ExampleMemory() {
	kvs := &kvstore.Memory{}
	if _, err := kvs.Get("akey"); !errors.Is(err, kvstore.ErrNoSuchKey) {
		log.Fatal("unexpected error", err)
	}
	val := []byte("value")
	if err := kvs.Set("akey", val); err != nil {
		log.Fatal("unexpected error", err)
	}
	out, err := kvs.Get("akey")
	if err != nil {
		log.Fatal("unexpected error", err)
	}
	fmt.Printf("%+v\n", reflect.DeepEqual(val, out))
	// Output: true
}
