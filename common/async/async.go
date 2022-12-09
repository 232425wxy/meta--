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
	return &TaskResultSet{
		chz:     chz,
		results: make([]TaskResult, len(chz)),
	}
}

// Reap ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Reap 将已经执行完毕，产生结果的任务标记一下OK为true。
func (set *TaskResultSet) Reap() *TaskResultSet {
	for i := 0; i < len(set.results); i++ {
		var ch = set.chz[i]
		select {
		case result, ok := <-ch:
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
