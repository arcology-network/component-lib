package aggregator

import ethCommon "github.com/HPISTechnologies/3rd-party/eth/common"

type DataPool struct {
	data map[ethCommon.Hash]interface{}
}

// NewDataPool returns a new DataPool structure.
func NewDataPool() *DataPool {
	return &DataPool{
		data: map[ethCommon.Hash]interface{}{},
	}
}

//object and raw enter pool
func (d *DataPool) add(h ethCommon.Hash, data interface{}) {
	d.data[h] = data
}

//get data and raw
func (d *DataPool) get(h ethCommon.Hash) interface{} {
	return d.data[h]
}

//remove data and raw from pool
func (d *DataPool) remove(h ethCommon.Hash) {
	delete(d.data, h)
}

//remove data and raw from pool
func (d *DataPool) count() int {
	return len(d.data)
}

//range data_pool
func (d *DataPool) Range(f func(hash ethCommon.Hash, val interface{}) bool) {
	for k, v := range d.data {
		if !f(k, v) {
			break
		}
	}
}
