package aggregator

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/arcology/component-lib/actor"
)

func TestAddDataThenPick(t *testing.T) {
	broker := &mockMsgBroker{}
	aggr := NewAggregator(broker, "clear", "data", "picklist", "clearlist", "output", reflect.TypeOf(""))

	// Using concreate type or interface{} both work.
	aggr.HandleMsg(&actor.Message{
		Name: "data",
		Data: []string{"hash1", "hash2"},
	})
	aggr.HandleMsg(&actor.Message{
		Name: "data",
		Data: []interface{}{"hash3"},
	})

	aggr.HandleMsg(&actor.Message{
		Name: "picklist",
		Data: []string{"hash1", "hash3"},
	})

	if broker.recvName != "output" ||
		!checkRecvData(broker.recvMsg.Data.([]string), []string{"hash1", "hash3"}) {
		t.Error("Failed")
	}
}

func TestPickThanAddData(t *testing.T) {
	broker := &mockMsgBroker{}
	aggr := NewAggregator(broker, "clear", "data", "picklist", "clearlist", "output", reflect.TypeOf(""))

	aggr.HandleMsg(&actor.Message{
		Name: "picklist",
		Data: []string{"hash1", "hash3"},
	})

	aggr.HandleMsg(&actor.Message{
		Name: "data",
		Data: []interface{}{"hash1"},
	})
	aggr.HandleMsg(&actor.Message{
		Name: "data",
		Data: []interface{}{"hash2", "hash3"},
	})

	if broker.recvName != "output" ||
		!checkRecvData(broker.recvMsg.Data.([]string), []string{"hash1", "hash3"}) {
		t.Error("Failed")
	}
}

func TestAddPartOfDataThenPickThenAddRestOfData(t *testing.T) {
	broker := &mockMsgBroker{}
	aggr := NewAggregator(broker, "clear", "data", "picklist", "clearlist", "output", reflect.TypeOf(""))

	aggr.HandleMsg(&actor.Message{
		Name: "data",
		Data: []interface{}{"hash1"},
	})

	aggr.HandleMsg(&actor.Message{
		Name: "picklist",
		Data: []string{"hash1", "hash3"},
	})

	aggr.HandleMsg(&actor.Message{
		Name: "data",
		Data: []interface{}{"hash2"},
	})
	aggr.HandleMsg(&actor.Message{
		Name: "data",
		Data: []interface{}{"hash3"},
	})

	if broker.recvName != "output" ||
		!checkRecvData(broker.recvMsg.Data.([]string), []string{"hash1", "hash3"}) {
		t.Error("Failed")
	}
}

func TestPickAll(t *testing.T) {
	broker := &mockMsgBroker{}
	aggr := NewAggregator(broker, "clear", "data", "picklist", "clearlist", "output", reflect.TypeOf(""))

	aggr.HandleMsg(&actor.Message{
		Name: "data",
		Data: []interface{}{"hash1"},
	})
	aggr.HandleMsg(&actor.Message{
		Name: "data",
		Data: []interface{}{"hash2"},
	})

	aggr.HandleMsg(&actor.Message{
		Name: "picklist",
		Data: []string{},
	})

	if broker.recvName != "output" ||
		!checkRecvData(broker.recvMsg.Data.([]string), []string{"hash1", "hash2"}) {
		t.Error("Failed to pick all with zero length pick list.")
		return
	}

	broker.recvName = ""
	broker.recvMsg = nil
	aggr.HandleMsg(&actor.Message{
		Name: "picklist",
		Data: nil,
	})
	if broker.recvName != "output" ||
		!checkRecvData(broker.recvMsg.Data.([]string), []string{"hash1", "hash2"}) {
		t.Error("Failed to pick all with nil pick list.")
		return
	}
}

func TestClearData(t *testing.T) {
	broker := &mockMsgBroker{}
	aggr := NewAggregator(broker, "clear", "data", "picklist", "clearlist", "output", reflect.TypeOf(""))

	aggr.HandleMsg(&actor.Message{
		Name: "data",
		Data: []interface{}{"hash1"},
	})
	aggr.HandleMsg(&actor.Message{
		Name: "data",
		Data: []interface{}{"hash2"},
	})
	aggr.HandleMsg(&actor.Message{
		Name: "data",
		Data: []interface{}{"hash3"},
	})

	aggr.HandleMsg(&actor.Message{
		Name: "clearlist",
		Data: []string{"hash1", "hash3"},
	})
	aggr.HandleMsg(&actor.Message{
		Name: "clear",
	})

	if !checkPool(aggr.pool, []string{"hash2"}) {
		t.Error("Failed")
	}
}

func TestClearNonExistData(t *testing.T) {
	broker := &mockMsgBroker{}
	aggr := NewAggregator(broker, "clear", "data", "picklist", "clearlist", "output", reflect.TypeOf(""))

	aggr.HandleMsg(&actor.Message{
		Name: "data",
		Data: []interface{}{"hash1"},
	})
	aggr.HandleMsg(&actor.Message{
		Name: "data",
		Data: []interface{}{"hash2"},
	})

	aggr.HandleMsg(&actor.Message{
		Name: "clearlist",
		Data: []string{"hash1", "hash3"},
	})
	aggr.HandleMsg(&actor.Message{
		Name: "clear",
	})

	if !checkPool(aggr.pool, []string{"hash2"}) {
		t.Error("Failed")
	}
}

func TestClearAllUsingEmptyArray(t *testing.T) {
	broker := &mockMsgBroker{}
	aggr := NewAggregator(broker, "clear", "data", "picklist", "clearlist", "output", reflect.TypeOf(""))

	aggr.HandleMsg(&actor.Message{
		Name: "data",
		Data: []interface{}{"hash1"},
	})
	aggr.HandleMsg(&actor.Message{
		Name: "data",
		Data: []interface{}{"hash2"},
	})

	aggr.HandleMsg(&actor.Message{
		Name: "clearlist",
		Data: []string{},
	})
	aggr.HandleMsg(&actor.Message{
		Name: "clear",
	})

	if len(aggr.pool) != 0 {
		t.Error("Failed")
	}
}

func TestClearAllUsingNilArray(t *testing.T) {
	broker := &mockMsgBroker{}
	aggr := NewAggregator(broker, "clear", "data", "picklist", "clearlist", "output", reflect.TypeOf(""))

	aggr.HandleMsg(&actor.Message{
		Name: "data",
		Data: []interface{}{"hash1"},
	})
	aggr.HandleMsg(&actor.Message{
		Name: "data",
		Data: []interface{}{"hash2"},
	})

	aggr.HandleMsg(&actor.Message{
		Name: "clearlist",
		Data: nil,
	})
	aggr.HandleMsg(&actor.Message{
		Name: "clear",
	})

	if len(aggr.pool) != 0 {
		t.Error("Failed")
	}
}

func TestDupDataWontSendPickListTwice(t *testing.T) {
	broker := &mockMsgBroker{}
	aggr := NewAggregator(broker, "clear", "data", "picklist", "clearlist", "output", reflect.TypeOf(""))

	aggr.HandleMsg(&actor.Message{
		Name: "data",
		Data: []interface{}{"hash1"},
	})
	aggr.HandleMsg(&actor.Message{
		Name: "data",
		Data: []interface{}{"hash2"},
	})

	aggr.HandleMsg(&actor.Message{
		Name: "picklist",
		Data: []string{"hash1", "hash3"},
	})

	aggr.HandleMsg(&actor.Message{
		Name: "data",
		Data: []interface{}{"hash3"},
	})

	if broker.recvName != "output" ||
		!checkRecvData(broker.recvMsg.Data.([]string), []string{"hash1", "hash3"}) {
		t.Error("Failed")
		return
	}

	broker.recvName = ""
	broker.recvMsg = nil
	aggr.HandleMsg(&actor.Message{
		Name: "data",
		Data: []interface{}{"hash3"},
	})
	if broker.recvName != "" || broker.recvMsg != nil {
		t.Error("Failed")
	}
}

func BenchmarkAddData(b *testing.B) {
	hashes := make([]string, 1000000)
	for i := range hashes {
		hashes[i] = strconv.Itoa(i)
	}

	broker := &mockMsgBroker{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		aggr := NewAggregator(broker, "clear", "data", "picklist", "clearlist", "output", reflect.TypeOf(""))
		for i := 0; i < len(hashes)/100; i++ {
			aggr.HandleMsg(&actor.Message{
				Name: "data",
				Data: hashes[i*100 : (i+1)*100],
			})
		}
	}
}

func BenchmarkSendData(b *testing.B) {
	hashes := make([]string, 1000000)
	for i := range hashes {
		hashes[i] = strconv.Itoa(i)
	}

	broker := &mockMsgBroker{}
	aggr := NewAggregator(broker, "clear", "data", "picklist", "clearlist", "output", reflect.TypeOf(""))
	for i := 0; i < len(hashes)/100; i++ {
		aggr.HandleMsg(&actor.Message{
			Name: "data",
			Data: hashes[i*100 : (i+1)*100],
		})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		aggr.HandleMsg(&actor.Message{
			Name: "picklist",
			Data: hashes,
		})
	}
}

func checkRecvData(got []string, expected []string) bool {
	if len(got) != len(expected) {
		return false
	}

	gotSet := make(map[string]struct{})
	for _, e := range got {
		gotSet[e] = struct{}{}
	}

	for _, e := range expected {
		if _, ok := gotSet[e]; !ok {
			return false
		}
	}

	return true
}

func checkPool(pool map[string]interface{}, expected []string) bool {
	if len(pool) != len(expected) {
		return false
	}

	for _, e := range expected {
		if _, ok := pool[e]; !ok {
			return false
		}
	}

	return true
}
