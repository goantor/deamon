package deamon

import (
	"github.com/goantor/logs"
	"sync"
)

var (
	registry = Registry{
		maps: &sync.Map{},
	}
)

func initLog() {
	logs.NewEntity(&logs.Options{
		Path:            "./logs",
		Level:           "debug",
		Stdout:          true,
		SaveDay:         1,
		TimestampFormat: "2006-01-02T15:04:06",
	}).Initialize()
}

type Registry struct {
	maps *sync.Map
}

func (r *Registry) get(name string) (task ITask, ok bool) {

	var val any
	if val, ok = r.maps.Load(name); !ok {
		return
	}

	task = val.(*Task)
	return
}

func (r *Registry) set(name string, task ITask) {
	r.maps.Store(name, task)
}

func (r *Registry) doRange(fn func(name string, task ITask)) {
	r.maps.Range(func(key, value any) bool {
		fn(key.(string), value.(ITask))
		return true
	})
}
