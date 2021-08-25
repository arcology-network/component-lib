package streamer

type DefaultProducer struct {
	name          string
	outputs       []string
	bufferLengths []int
}

func NewDefaultProducer(name string, outputs []string, bufferLengths []int) StreamProducer {
	return &DefaultProducer{
		name:          name,
		outputs:       outputs,
		bufferLengths: bufferLengths,
	}
}

func (p *DefaultProducer) GetName() string {
	return p.name
}

func (p *DefaultProducer) GetOutputs() []string {
	return p.outputs
}

func (p *DefaultProducer) GetBufferLengths() []int {
	return p.bufferLengths
}
