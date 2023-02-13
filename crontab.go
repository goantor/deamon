package deamon

import (
	"fmt"
	"github.com/robfig/cron"
	"os"
	"os/signal"
	"syscall"
)

func NewCronTask(name string, kind Kind, handler TaskHandler, opts *Options) *CronTask {
	return &CronTask{
		taskBase: &taskBase{
			name:    name,
			kind:    kind,
			handler: handler,
			opts:    opts,
			exit:    make(Exit, 0),
		},
	}
}

type CronTask struct {
	*taskBase
}

func (c CronTask) Trigger() {
	t := cron.New()

	fmt.Println(c.opts.CronString)
	err := t.AddFunc(c.opts.CronString, func() {
		go c.distribute()
		c.wait()
	})

	if err != nil {
		panic(err)
	}

	t.Start()

	osc := make(chan os.Signal, 1)
	signal.Notify(osc, syscall.SIGTERM, syscall.SIGINT)
stop:
	for {
		select {
		case <-osc:
			t.Stop()
			break stop
		}
	}

}

func (c CronTask) distribute() {
	defer c.catch()
	c.resetContent()
	c.handler(c.ctx, c.name, c.exit)
}
