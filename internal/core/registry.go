package core

type registry struct {
	tools map[string]Tool
}

func NewRegistry() ToolRegistry {
	reg := registry{tools: make(map[string]Tool)}
	return &reg
}

func (r *registry) Register(t Tool) {
	r.tools[t.Spec().Name] = t
}

func (r *registry) Get(name string) (Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

func (r *registry) Specs() []ToolSpec {
	s := make([]ToolSpec, 0, len(r.tools))
	for _, t := range r.tools {
		s = append(s, t.Spec())
	}
	return s
}
