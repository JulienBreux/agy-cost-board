package main

import "testing"

func TestMainExecution(t *testing.T) {
	orig := runCLI
	defer func() { runCLI = orig }()

	called := false
	runCLI = func() {
		called = true
	}

	main()

	if !called {
		t.Errorf("expected main to invoke runCLI")
	}
}
