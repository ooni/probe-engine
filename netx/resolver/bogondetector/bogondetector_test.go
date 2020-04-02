package bogondetector

import "testing"

func TestIntegration(t *testing.T) {
	if Check("antani") != true {
		t.Fatal("unexpected result")
	}
	if Check("127.0.0.1") != true {
		t.Fatal("unexpected result")
	}
	if Check("1.1.1.1") != false {
		t.Fatal("unexpected result")
	}
	if Check("10.0.1.1") != true {
		t.Fatal("unexpected result")
	}
}
