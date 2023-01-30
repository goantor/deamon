package deamon

import (
	"github.com/goantor/x"
	"github.com/robfig/cron"
	"sync"
	"time"
)

var (
	DaemonManager *DefaultDaemonManager
	once          sync.Once
)

func init() {
	DaemonManager = NewDefaultDaemonManager()
}

type Daemon struct {
	makeHandler HandlerFunc
	daemonType  int
}

// Task 一次性任务
type Task struct {
	*Daemon
	task chan interface{}
}

// Loop 循环任务
type Loop struct {
	*Daemon
	interval time.Duration // 间隔默认为0
}

// Crontab 定时任务
type Crontab struct {
	*Daemon
	crontabOption string
	crontab       *cron.Cron
}

// Handler handlerFunc 参数
type Handler struct {
	Context x.Context
	Name    string
	Func    func() error
}

func NewHandler(ctx x.Context, name string, fn func() error) *Handler {
	return &Handler{Context: ctx, Name: name, Func: fn}
}

// TaskFunc 任务回调
type TaskFunc func(handler Handler)

// HandlerFunc 具体handler
type HandlerFunc func() TaskFunc

func NewDefaultDaemonManager() *DefaultDaemonManager {
	once.Do(func() {
		DaemonManager = &DefaultDaemonManager{
			daemons: make(map[string]interface{}),
			stops:   make(map[string]chan bool),
		}
	})
	return DaemonManager
}

type DefaultDaemonManager struct {
	daemons map[string]interface{}
	stops   map[string]chan bool
}

func (d DefaultDaemonManager) loop(stop chan bool, daemonInf interface{}) {
	daemon := daemonInf.(*Loop)
	handler := daemon.makeHandler()

	handler(Handler{})
stopHere:
	for {
		tick := time.Tick(daemon.interval)
		select {
		case <-tick:
			go handler(Handler{})

		case <-stop:
			break stopHere
		}
	}
}

func (d DefaultDaemonManager) crontab(stop chan bool, daemonInf interface{}) {
	daemon := daemonInf.(*Crontab)
	daemon.crontab.Start()
stopHere:
	for {
		select {
		case <-stop:
			daemon.crontab.Stop()
			break stopHere
		}
	}
}

func (d DefaultDaemonManager) task(stop chan bool, daemonInf interface{}) {
	daemon := daemonInf.(*Task)
	handler := daemon.makeHandler()

stopHere:
	for {
		select {
		case data := <-daemon.task:
			go handler(data.(Handler))

		case <-stop:
			close(daemon.task)
			break stopHere
		}
	}
}

func (d *DefaultDaemonManager) GetDaemons() map[string]chan bool {
	return d.stops
}

func (d *DefaultDaemonManager) RegisterTask(name string, handler HandlerFunc) TaskFunc {
	task := make(chan interface{})

	d.daemons[name] = &Task{
		Daemon: &Daemon{makeHandler: handler},
		task:   task,
	}

	return func(handler Handler) {
		task <- handler
	}
}

func (d *DefaultDaemonManager) RegisterLoop(name string, handler HandlerFunc, interval time.Duration) {
	d.daemons[name] = &Loop{
		Daemon:   &Daemon{makeHandler: handler},
		interval: interval,
	}
}

// RegisterCrontab 定时任务
// name: 任务名
// handler: 任务
// crontabOption: 定时器参数 分、时、日、月、周 "*/1 * * * *"
func (d *DefaultDaemonManager) RegisterCrontab(name string, handlerFunc HandlerFunc, crontabOption string) {
	crontab := cron.New()

	crontabHandler := handlerFunc()
	_ = crontab.AddFunc(crontabOption, func() {
		crontabHandler(Handler{})
	})

	d.daemons[name] = &Crontab{
		Daemon:        &Daemon{makeHandler: handlerFunc},
		crontabOption: crontabOption,
		crontab:       crontab,
	}
}

func (d *DefaultDaemonManager) Start() {
	for name, daemon := range d.daemons {
		d.stops[name] = make(chan bool, 1)

		switch daemon.(type) {
		case *Task:
			go d.task(d.stops[name], daemon)
			break
		case *Crontab:
			go d.crontab(d.stops[name], daemon)
			break
		case *Loop:
			go d.loop(d.stops[name], daemon)
			break
		}
	}
}

// Stop 无效、稍后开放web接口
func (d *DefaultDaemonManager) Stop() {
	for _, c := range d.stops {
		c <- true
		close(c)
	}
}
