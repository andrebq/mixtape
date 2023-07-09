package actor_test

import (
	"context"
	"testing"

	"github.com/andrebq/mixtape/actor"
)

func TestActorWithState(t *testing.T) {
	ac, err := actor.LoadClass(`
	fmt := import("fmt")
	json := import("json")
	state := {bias: 0}
	actor := {
		setState: func(newState) { state = newState },
		getState: func() { return state },

		double: func(message) {
			state.bias = state.bias + 1
			return message.value * 2 + state.bias
		}
	    }
	export immutable(actor)
	`)
	if err != nil {
		t.Fatal(err)
	}

	output, state, err := ac.HandleStateful(context.Background(), "double", `{"bias": 10}`, `{"value": 1.0}`)
	if err != nil {
		t.Fatal(err)
	} else if fval, ok := output.(float64); !ok {
		t.Fatalf("Should have returned a float64, but got %T", output)
	} else if fval != 13.0 {
		t.Fatalf("Should be 2.0 but got %v", fval)
	} else if state != `{"bias":11}` {
		t.Fatalf("State is not we expected: %v", state)
	}
}

func TestStateInitialization(t *testing.T) {
	ac, err := actor.LoadClass(`
	fmt := import("fmt")
	json := import("json")
	state := {}
	actor := {
		getState: func(){ return state },
		setState: func(nval){ state = nval },
		initializeState: func(args){
			state.done = true
			return true
		}
	    }
	export immutable(actor)
	`)
	if err != nil {
		t.Fatal(err)
	}

	output, state, err := ac.HandleStateful(context.Background(), "initializeState", "{}", "{}")
	if err != nil {
		t.Fatal(err)
	} else if bval, ok := output.(bool); !ok {
		t.Fatalf("Should have returned a float64, but got %T", output)
	} else if !bval {
		t.Fatal("Initialization failed")
	} else if state != `{"done":true}` {
		t.Fatalf("State is not we expected: %v", state)
	}
}

func TestBasicBehaviour(t *testing.T) {
	ac, err := actor.LoadClass(`
	fmt := import("fmt")
	json := import("json")
	actor := {
		print: func(message) { fmt.println("got message", json.encode(message)) },
		double: func(message) { return message.value * 2 }
	    }
	export immutable(actor)
	`)
	if err != nil {
		t.Fatal(err)
	}
	output, err := ac.Handle(context.Background(), "print", `[123, {"a":[123, 123.0, "abc", true]}]`)
	if err != nil {
		t.Fatal(err)
	}
	if output != nil {
		t.Fatal("the actor script does not generate any output for print calls, but got", output)
	}

	output, err = ac.Handle(context.Background(), "double", `{"value": 1.0}`)
	if err != nil {
		t.Fatal(err)
	} else if fval, ok := output.(float64); !ok {
		t.Fatalf("Should have returned a float64, but got %T", output)
	} else if fval != 2.0 {
		t.Fatalf("Should be 2.0 but got %v", fval)
	}
}
