package deamon

import (
	"fmt"
	"github.com/goantor/x"
	"math/rand"
	"testing"
	"time"
)

func init() {
	initLog()
}

func TestTask(t *testing.T) {
	RegisterTask("testing", TaskKind, func(ctx x.Context, name string, exit Exit) {
		rs := rand.New(rand.NewSource(time.Now().UnixNano()))
		n := rs.Intn(15)
		t.Logf("run %d times\n", n)
		for i := 0; i < 10; i++ {
			t.Logf("do task %d\n", i)
			time.Sleep(time.Millisecond * 500)
			if i > n {
				t.Logf("do error: %d\n", i)
				panic("take error")
			}
		}

		exit <- nil
	}, nil)

	Start()
	//Run("testing")
}

func TestQueueTask(t *testing.T) {
	RegisterTask("testing", QueueKind, func(ctx x.Context, name string, exit Exit) {
		rs := rand.New(rand.NewSource(time.Now().UnixNano()))
		n := rs.Intn(15)
		t.Logf("run %d times\n", n)
		for i := 0; i < 10; i++ {
			t.Logf("do task %d\n", i)
			time.Sleep(time.Millisecond * 500)
			if i > n {
				t.Logf("do error: %d\n", i)
				panic("take error")
			}
		}

		exit <- nil
	}, nil)

	Start()
	//Run("testing")
}

// todo 这个有问题
func TestLoopTask(t *testing.T) {
	RegisterTask("testing", LoopKind, func(ctx x.Context, name string, exit Exit) {
		rs := rand.New(rand.NewSource(time.Now().UnixNano()))
		n := rs.Intn(3)
		t.Logf("run %d times\n", n)
		for i := 0; i < 6; i++ {
			t.Logf("do task %d\n", i)
			time.Sleep(time.Millisecond * 1000)
			if i > n {
				t.Logf("do error: %d\n", i)
				panic("take error")
			}
		}

		fmt.Println("task finished")
		exit <- nil
	}, &Options{Interval: time.Second * 4})

	Start()
	//Run("testing")
}

func TestCronTask(t *testing.T) {
	RegisterTask("testing", CronKind, func(ctx x.Context, name string, exit Exit) {
		rs := rand.New(rand.NewSource(time.Now().UnixNano()))
		n := rs.Intn(15)
		t.Logf("run %d times\n", n)
		for i := 0; i < 10; i++ {
			t.Logf("do task %d\n", i)
			time.Sleep(time.Millisecond * 500)
			if i > n {
				t.Logf("do error: %d\n", i)
				panic("take error")
			}
		}

		fmt.Println("task finished")
		exit <- nil
	}, &Options{CronString: "*/15 * * * * *"})

	Start()
	//Run("testing")
}
