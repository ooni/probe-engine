package libminiooni

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestFileInputFailWithOtherInputs(t *testing.T) {
	// test function panics with both command line and file inputs present
	defer func() { recover() }()

	opts := Options{Inputs: []string{"test"}, InputFilePath: "testing"}
	loadFileInputs(&opts)

	t.Errorf("expected panic when inputs are specified in multiple ways.")
}

func TestInputFile(t *testing.T) {
	// create test input file
	input1 := "my input 1"
	input2 := "my input 2"
	// add newline at end of file like vim does; while there also allow blank
	// lines in the middle of the file
	fileContents := []byte(input1 + "\n\n" + input2 + "\n")
	inputfile, err := ioutil.TempFile("", "myinput")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(inputfile.Name())

	if _, err := inputfile.Write(fileContents); err != nil {
		t.Fatal(err)
	}
	if err := inputfile.Close(); err != nil {
		t.Fatal(err)
	}

	opts := Options{InputFilePath: inputfile.Name()}
	loadFileInputs(&opts)

	if len(opts.Inputs) != 2 {
		t.Error("expected two inputs to be loaded from file")
	}

	if opts.Inputs[0] != input1 {
		t.Errorf("expected first input to be %v got %v", input1, opts.Inputs[0])
	}
	if opts.Inputs[1] != input2 {
		t.Errorf("expected second input to be %v got %v", input2, opts.Inputs[1])
	}
}
