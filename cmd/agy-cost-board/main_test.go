package main

import "testing"

func TestMainExecution(t *testing.T) {
	orig := runCLI
	t.Cleanup(func() { runCLI = orig })

	called := false
	runCLI = func() error {
		called = true
		return nil
	}

	main()

	if !called {
		t.Errorf("expected main to invoke runCLI")
	}
}
