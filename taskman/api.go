package taskman

type (
	AgentDetails struct {
		ID            string   //`msgpack:"id"`
		Name          string   //`msgpack:"name"`
		Executors     []string //`msgpack:"executors"`
		MaxConcurrent int64    //`msgpack:"max_concurrent"`
	}
)
