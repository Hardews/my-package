package cron

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

const (
	everytime       = 0  // 每时每刻
	pollingInterval = -1 // 每隔一段时间执行
	fixedTime       = -2 // 固定时间执行

	mouthOfSecond  = 30 * 24 * 60 * 60 * 1
	dayOfSecond    = 24 * 60 * 60 * 1
	hourOfSecond   = 60 * 60 * 1
	minuteOfSecond = 60 * 1
	second         = 1
)

type Engine struct {
	Jobs       map[string]*job
	unregister chan bool
}

type job struct {
	task   func(a ...interface{})
	times  [5]times
	ticker *time.Ticker
}

type times struct {
	base      int
	saveTimes []int
	now       int
}

func NewEngine() *Engine {
	return &Engine{
		Jobs:       make(map[string]*job),
		unregister: make(chan bool, 1),
	}
}

func (e *Engine) Start() {
	for _, j := range e.Jobs {
		t := j
		go func() {
			t.ticker = time.NewTicker(t.SetTicker())
			for {
				select {
				case <-t.ticker.C:
					t.task()
					t.NextTicker()
				case <-e.unregister:
					return
				}
			}
		}()
	}
}

func (e *Engine) Reset() {
	for i, _ := range e.Jobs {
		delete(e.Jobs, i)
	}
	e.Del()
}

func (e *Engine) Del() {
	go func() {
		for range e.Jobs {
			e.unregister <- true
			time.Sleep(2 * time.Second)
		}
		fmt.Println("all task down")
		return
	}()
}

func (e *Engine) AddFunc(name string, spec string, task func(a ...interface{})) {
	e.Jobs[name] = &job{
		task: task,
	}
	e.Jobs[name].initJob(strings.Split(spec, " "))
}

func (e *Engine) DelFunc(name string) {
	delete(e.Jobs, name)
}

func (j *job) initJob(times []string) {
	// 年 此字段在标准/默认实现中不受支持。 不处理

	// 月
	j.times[0].saveTimes = parseSpecificTime(times[4])
	j.times[0].now = 1
	j.times[0].base = mouthOfSecond

	// 日
	j.times[1].saveTimes = parseSpecificTime(times[3])
	j.times[1].now = 1
	j.times[1].base = dayOfSecond

	// 小时
	j.times[2].saveTimes = parseSpecificTime(times[2])
	j.times[2].now = 1
	j.times[2].base = hourOfSecond

	// 分钟
	j.times[3].saveTimes = parseSpecificTime(times[1])
	j.times[3].now = 1
	j.times[3].base = minuteOfSecond

	// 秒
	j.times[4].saveTimes = parseSpecificTime(times[0])
	j.times[4].now = 1
	j.times[4].base = second
}

func parseSpecificTime(aTime string) []int {
	var t []string
	var mt = make([]int, 0)
	if aTime != "*" {
		if strings.Contains(aTime, "*/") {
			t = strings.Split(aTime, "/")
			// 如果成立则指的是每过多久后执行
			mt = append(mt, pollingInterval)

			i, err := strconv.Atoi(t[1])
			if err != nil {
				log.Println(err)
			}

			mt = append(mt, i)
		} else {
			t = strings.Split(aTime, ",")
			// 否则是指定时刻该执行
			mt = append(mt, fixedTime)

			for _, s := range t {
				i, err := strconv.Atoi(s)
				if err != nil {
					log.Println(err)
				}
				mt = append(mt, i)
			}
		}

	} else {
		mt = append(mt, everytime)
	}

	return mt
}

func (j *job) NextTicker() {
	j.ticker = time.NewTicker(j.SetTicker())
}

func (j *job) SetTicker() time.Duration {
	var td = 1 * time.Second
	for i, _ := range j.times {
		switch j.times[i].saveTimes[0] {
		case everytime:
		case pollingInterval:
			td += time.Duration(j.times[i].saveTimes[1]*j.times[i].base) * time.Second
		case fixedTime:
			switch j.times[i].base {
			case mouthOfSecond:
				m := GetMouth(j.times[i].saveTimes[j.times[i].now])
				td += time.Date(2022, m, time.Now().Day(), time.Now().Hour(),
					time.Now().Minute(), time.Now().Second(), time.Now().Nanosecond(),
					time.Local).Sub(time.Now())

			case dayOfSecond:
				if j.times[i].saveTimes[j.times[i].now] < time.Now().Day() {
					j.times[i].saveTimes[j.times[i].now] += 30
				}

				td += time.Date(2022, time.Now().Month(), j.times[i].saveTimes[j.times[i].now], time.Now().Hour(),
					time.Now().Minute(), time.Now().Second(), time.Now().Nanosecond(),
					time.Local).Sub(time.Now())

			case hourOfSecond:
				if j.times[i].saveTimes[j.times[i].now] < time.Now().Hour() {
					j.times[i].saveTimes[j.times[i].now] += 24
				}

				td += time.Date(2022, time.Now().Month(), time.Now().Day(), j.times[i].saveTimes[j.times[i].now],
					time.Now().Minute(), time.Now().Second(), time.Now().Nanosecond(),
					time.Local).Sub(time.Now())

			case minuteOfSecond:

				if j.times[i].saveTimes[j.times[i].now] < time.Now().Minute() {
					j.times[i].saveTimes[j.times[i].now] += 60
				}

				td += time.Date(2022, time.Now().Month(), time.Now().Day(), time.Now().Hour(),
					j.times[i].saveTimes[j.times[i].now], time.Now().Second(), time.Now().Nanosecond(),
					time.Local).Sub(time.Now())

			case second:
				if j.times[i].saveTimes[j.times[i].now] < time.Now().Second() {
					j.times[i].saveTimes[j.times[i].now] += 60
				}

				td += time.Date(2022, time.Now().Month(), time.Now().Day(), time.Now().Hour(),
					time.Now().Minute(), j.times[i].saveTimes[j.times[i].now], time.Now().Nanosecond(),
					time.Local).Sub(time.Now())
			}
			if j.times[i].now < len(j.times[i].saveTimes)-1 {
				j.times[i].now += 1
			} else {
				j.times[i].now = 1
			}
		}
	}

	return td
}

func GetMouth(m int) time.Month {
	switch m {
	case 1:
		return time.January
	case 2:
		return time.February
	case 3:
		return time.March
	case 4:
		return time.April
	case 5:
		return time.May
	case 6:
		return time.June
	case 7:
		return time.July
	case 8:
		return time.August
	case 9:
		return time.September
	case 10:
		return time.October
	case 11:
		return time.November
	case 12:
		return time.December
	default:
		return time.July
	}
}
