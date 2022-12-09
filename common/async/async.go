package async

type Task func(i int) (val interface{}, abort bool, err error)

type TaskResult struct {
	Value interface{}
	Error error
	OK    bool
}

type TaskResultSet struct {
	chz     []chan TaskResult
	results []TaskResult
}

func newTaskResultSet(chz []chan TaskResult) *TaskResultSet {
	set := &TaskResultSet{chz: chz, results: make([]TaskResult, len(chz))}
	for i := 0; i < len(chz); i++ {
		select {
		case result, ok := <-chz[i]:
			if ok {
				result.OK = true
				set.results[i] = result
			}
		default:
		}
	}
	return set
}

// FirstValue ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// FirstValue 返回任务集合中第一个不为空的值。
func (set *TaskResultSet) FirstValue() interface{} {
	for i := 0; i < len(set.results); i++ {
		if set.results[i].Value != nil {
			return set.results[i].Value
		}
	}
	return nil
}

// FirstError ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// FirstError 返回任务集合中第一个Error不为空的error。
func (set *TaskResultSet) FirstError() error {
	for i := 0; i < len(set.results); i++ {
		if set.results[i].Error != nil {
			return set.results[i].Error
		}
	}
	return nil
}

// Parallel ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Parallel 并发的去执行若干个任务。
func Parallel(tasks ...Task) (*TaskResultSet, bool) {
	var taskResultsChz = make([]chan TaskResult, len(tasks))
	var taskDoneChz = make(chan bool, len(tasks))
	ok := true

	for i, task := range tasks {
		var ch = make(chan TaskResult, 1)
		taskResultsChz[i] = ch
		go func(i int, task Task, ch chan TaskResult) {
			var val, abort, err = task(i)
			ch <- TaskResult{Value: val, Error: err}
			taskDoneChz <- abort
		}(i, task, ch)
	}
	for i := 0; i < len(tasks); i++ {
		abort := <-taskDoneChz
		if abort {
			ok = false
			break
		}
	}
	return newTaskResultSet(taskResultsChz), ok
}
