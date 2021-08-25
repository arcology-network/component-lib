package streamer

type DefaultConsumer struct {
	name       string
	inputs     []string
	controller StreamController
}

func NewDefaultConsumer(name string, inputs []string, controller StreamController) StreamConsumer {
	return &DefaultConsumer{
		name:       name,
		inputs:     inputs,
		controller: controller,
	}
}

func (c *DefaultConsumer) GetName() string {
	return c.name
}

func (c *DefaultConsumer) GetInputs() []string {
	return c.inputs
}

func (c *DefaultConsumer) GetStreamController() StreamController {
	return c.controller
}
