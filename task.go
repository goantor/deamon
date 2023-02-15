package deamon

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/goantor/logs"
	"github.com/goantor/pr"
	"github.com/goantor/x"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	TaskKind Kind = iota + 1
	QueueKind
	LoopKind
	CronKind
)

type Kind int

type Exit chan error

type Options struct {
	Interval   time.Duration
	CronString string
}

type TaskHandler func(ctx x.Context, name string, exit Exit)

func RegisterTask(name string, kind Kind, handler TaskHandler, opts *Options) {
	var task ITask
	if kind == CronKind {
		task = NewCronTask(name, kind, handler, opts)
	} else {
		task = NewTask(name, kind, handler, opts)
	}

	registry.set(name, task)
}

func Start() {
	registry.doRange(func(name string, task ITask) {
		go func() {
			pr.PrintGreen("cron %s running...\n", name)
			task.Trigger()
			pr.PrintGreen("cron %s running finished\n", name)
		}()
	})
}

type ITask interface {
	Trigger()
}

func NewTask(name string, kind Kind, handler TaskHandler, opts *Options) *Task {
	return &Task{
		taskBase: &taskBase{
			name:    name,
			kind:    kind,
			handler: handler,
			opts:    opts,
			exit:    make(Exit, 0),
		},
	}
}

type taskBase struct {
	ctx     x.Context
	name    string
	kind    Kind
	handler TaskHandler
	opts    *Options
	exit    Exit
}

func BuildContext(name string) x.Context {
	gtx := &gin.Context{}
	log := logs.New("PUT", fmt.Sprintf("%s::%s", "cron", name), "")
	return x.NewContextWithGin(gtx, log)
}

func (t *taskBase) resetContent() {
	t.ctx = BuildContext(t.name)
}

func (t *taskBase) wait() {
	if err := <-t.exit; err != nil {
		t.ctx.Error(fmt.Sprintf("cron %s found error", t.name), x.H{
			"error": err,
		})
		return
	}

	t.ctx.Info(fmt.Sprintf("cron %s exit", t.name), nil)
}

func (t *taskBase) catch() {
	if err := recover(); err != nil {
		t.exit <- fmt.Errorf("cron %s catch error %v", t.name, err)
	}
}

type Task struct {
	*taskBase
}

func (t *Task) Trigger() {
	// 分配任务
	go t.distribute()

	// 监听
	t.watch()
}

func (t *Task) distribute() {
	defer t.catch()
	t.resetContent()
	t.handler(t.ctx, t.name, t.exit)
}

func (t *Task) watch() {
	switch t.kind {
	default:
		t.wait()
		break
	case LoopKind:
		t.watchLoop()
		break

	case QueueKind:
		t.watchExit()
		break
	}
}

// loop 循环顺序执行
func (t *Task) watchLoop() {
	if t.opts.Interval == 0 {
		panic(fmt.Sprintf("cron loop %s interval is `0` it is not supported", t.name))
	}

	fmt.Println("do watch loop")
	osc := make(chan os.Signal, 1)
	signal.Notify(osc, syscall.SIGTERM, syscall.SIGINT)

	go t.next()

	// 不执行了
stop:
	for { // 阻塞了
		select {
		case <-osc:
			t.ctx.Info(fmt.Sprintf("cron %s got signal do exit", t.name), nil)
			break stop
		}

	}
}

func (t *Task) next() {
	for { // 阻塞了
		time.Sleep(t.opts.Interval) // 每十秒 完成一次 如果未完成 那就得完成后再继续执行
		t.wait()

		go t.distribute()
	}
}

func (t *Task) watchExit() {
	osc := make(chan os.Signal, 1)
	signal.Notify(osc, syscall.SIGTERM, syscall.SIGINT)
	defer close(t.exit)

stop:
	for {
		select {
		case err := <-t.exit: // queue  -> listen queue
			if err != nil {
				t.ctx.Error(fmt.Sprintf("cron %s watch exit", t.name), x.H{
					"error": err,
				})
			}

			go t.distribute() // 出错了。
			break
		case <-osc:
			t.ctx.Info(fmt.Sprintf("cron %s got signal do exit", t.name), nil)
			break stop
		}
	}

}
