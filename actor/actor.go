package actor

import (
	"context"
	"fmt"
	"time"

	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/stdlib"
	"github.com/rs/zerolog/log"
)

type (
	ActorClass struct {
		c *tengo.Compiled
	}
)

func LoadClass(body string) (*ActorClass, error) {
	s := tengo.NewScript([]byte(`
	// TODO: this initialization scripts needs a lot of attention
	//
	// the basic premise is:
	// every time we run the script, it should perform actor initialization (if necessary),
	// set the state of the actor,
	// dispatch the call to the method handler
	// read the state to persist it later
	//
	// right now this works, but the intention isn't very clear
	json := import("json")
	actor := import("@actor")
	contentObj := json.decode(content)

	method := actor[methodName]

	if (!method && methodName == "initializeState") {
		method = func(_dummy){ return true }
	}

	stateObj := {}
	if (state != "") {
		stateObj = json.decode(state)
	}
	setState := actor["setState"]

	if (methodName == "initializeState") {
		setState := func(_dummy){}
	}

	if (setState && (state != "")) {
		setState(stateObj)
	}

	output := method(contentObj)

	if (methodName == "initializeState") {
		output = true
	}

	getState := actor["getState"]

	if (getState) {
		stateObj = getState()
		state = json.encode(stateObj)
	}
	`))
	s.Add("content", "")
	s.Add("methodName", "")
	s.Add("state", "")
	modules := stdlib.GetModuleMap(stdlib.AllModuleNames()...)
	modules.AddSourceModule("@actor", []byte(body))
	s.SetImports(modules)
	c, err := s.Compile()
	if err != nil {
		return nil, fmt.Errorf("actor: unable to parse script: %v", err)
	}
	return &ActorClass{c}, nil
}

// Handle the given message type with the given content
func (a *ActorClass) Handle(ctx context.Context, name string, content string) (interface{}, error) {
	c := a.c.Clone()
	c.Set("methodName", name)
	c.Set("content", content)
	c.Set("state", "")

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	err := c.RunContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("actor: %v", err)
	}

	output := c.Get("output")
	log.Ctx(ctx).Info().Interface("output", output.Value()).Msg("Done")
	return output.Value(), nil
}

func (a *ActorClass) HandleStateful(ctx context.Context, name, state, content string) (interface{}, string, error) {
	c := a.c.Clone()
	c.Set("methodName", name)
	c.Set("content", content)
	c.Set("state", state)

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	err := c.RunContext(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("actor: %v", err)
	}

	output := c.Get("output")
	log.Ctx(ctx).Info().Interface("output", output.Value()).Msg("Done")
	state = c.Get("state").String()
	return output.Value(), state, nil
}
