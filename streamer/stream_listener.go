package streamer

type streamListener interface {
	Notify(newItem interface{})
}

type defaultListener struct {
	name       string
	controller StreamController
}

func (sul *defaultListener) Notify(newItem interface{}) {
	sul.controller.Notify(sul.name, newItem)
}
