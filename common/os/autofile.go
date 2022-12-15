package os

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/232425wxy/meta--/common/rand"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	autoFileClosePeriod = 1000 * time.Millisecond // 默认打开的 autoFile 会在 1 秒钟以后被关闭
	autoFilePerms       = os.FileMode(0600)
)

// AutoFile 在打开 1 秒钟以后会被自动关闭，或者在收到 SIGHUB 信号时也会被关闭
// 可以用在日志文件里
type AutoFile struct {
	ID   string
	Path string // autoFile 的真正地址，是绝对路径

	closeTicker      *time.Ticker
	closeTickerStopc chan struct{} // closed when closeTicker is stopped
	hupc             chan os.Signal

	mtx  sync.Mutex
	file *os.File
}

// OpenAutoFile 创建一个 AutoFile 实例
func OpenAutoFile(path string) (*AutoFile, error) {
	var err error
	path, err = filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	af := &AutoFile{
		ID:               rand.Str(12) + ":" + path, // 随机创建 autoFile 的 ID
		Path:             path,
		closeTicker:      time.NewTicker(autoFileClosePeriod),
		closeTickerStopc: make(chan struct{}),
	}
	if err = af.openFile(); err != nil {
		_ = af.Close()
		return nil, err
	}

	// 等待 syscall.SIGHUP 信号来关闭 af
	af.hupc = make(chan os.Signal, 1)
	signal.Notify(af.hupc, syscall.SIGHUP) // 将接收到的 syscall.SIGHUP 信号放到 af.hupc channel 里
	go func() {                            // 时刻监听 af.hupc 里是否有 syscall.SIGHUP 信号，有的话，就把 af 关了
		for range af.hupc {
			_ = af.closeFile()
		}
	}()

	go af.closeFileRoutine() // 将关闭 af 的功能放到 goroutine 里

	return af, nil
}

// Close 关闭 af
func (af *AutoFile) Close() error {
	af.closeTicker.Stop()
	close(af.closeTickerStopc) // 向 af.closeFileRoutine() 发送信号，表示可以关闭 af 了
	if af.hupc != nil {        // 如果 af.hupc 初始化了，则假装向 af.hupc 里发送 syscall.SIGHUP 信号，从而关闭 af
		close(af.hupc)
	}
	return af.closeFile()
}

func (af *AutoFile) closeFileRoutine() {
	for {
		select {
		case <-af.closeTicker.C: // 默认情况下，在 af 打开 1 秒钟以后，就会调用 af.closeFile()
			_ = af.closeFile()
		case <-af.closeTickerStopc:
			return
		}
	}
}

func (af *AutoFile) closeFile() (err error) {
	af.mtx.Lock()
	defer af.mtx.Unlock()

	file := af.file // 让 file 等于 af.file，目的是为了关闭 af.file
	if file == nil {
		return nil
	}

	af.file = nil // 这一步很关键，让 af.file 等于 nil
	return file.Close()
}

// Write 向 af.file 中写入 b，返回写入内容的字节数和错误，
// 如果 af.file 等于 nil，则有必要调用 af.openFile()
func (af *AutoFile) Write(b []byte) (n int, err error) {
	af.mtx.Lock()
	defer af.mtx.Unlock()

	if af.file == nil {
		if err = af.openFile(); err != nil {
			return
		}
	}

	n, err = af.file.Write(b)
	return
}

// Sync 将 af.file 里的内容同步到硬盘上，如果 af 已经被关闭了，
// 则有必要调用 af.openFile()
func (af *AutoFile) Sync() error {
	af.mtx.Lock()
	defer af.mtx.Unlock()

	if af.file == nil {
		if err := af.openFile(); err != nil {
			return err
		}
	}
	return af.file.Sync()
}

func (af *AutoFile) openFile() error {
	file, err := os.OpenFile(af.Path, os.O_RDWR|os.O_CREATE|os.O_APPEND, autoFilePerms) // 真正的打开文件操作
	if err != nil {
		return err
	}
	af.file = file
	return nil
}

// Size 返回 af.file 的大小，如果 af 已经被关闭了，
// 则有必要调用 af.openFile()
func (af *AutoFile) Size() (int64, error) {
	af.mtx.Lock()
	defer af.mtx.Unlock()

	if af.file == nil {
		if err := af.openFile(); err != nil {
			return -1, err
		}
	}

	stat, err := af.file.Stat()
	if err != nil {
		return -1, err
	}
	return stat.Size(), nil
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// group

const (
	defaultGroupCheckDuration = 5000 * time.Millisecond
	defaultHeadSizeLimit      = 10 * 1024 * 1024       // 10MB head 文件的大小限制
	defaultTotalSizeLimit     = 1 * 1024 * 1024 * 1024 // 1GB Group 的容量限制
	maxFilesToRemove          = 4                      // needs to be greater than 1
)

// Group 可以通过 Group 来限制 AutoFile，比如每个块的最大大小或 Group 中存储的总字节数
type Group struct {
	ID                 string
	Head               *AutoFile // The head AutoFile to write to
	headBuf            *bufio.Writer
	Dir                string // 存放 Head 文件的目录
	ticker             *time.Ticker
	mtx                sync.Mutex
	headSizeLimit      int64
	totalSizeLimit     int64
	groupCheckDuration time.Duration

	// [min, min+1, ..., max-1, max] 这些是 Group 里所有文件的索引，其中
	// [min, min+1, ..., max-1] 是 Group 里其他冗余文件的索引，而 max 是
	// head 文件索引，尽管冗余文件的名字形如 headPath.xxx，其中 xxx 就是冗余
	// 文件在 Group 里的索引，尽管 head 文件的名字中不带数字，但是可以用 max
	// 来指向 head 文件
	minIndex int
	maxIndex int

	// 该 channel 会在 processTicks routine 完成时被关闭
	// 这样我们就可以清理目录里的文件
	doneProcessTicks chan struct{}
}

// OpenGroup 根据 headPath 确定存放文件的目录在哪里，然后创建路径为 headPath 的 AutoFile
func OpenGroup(headPath string) (*Group, error) {
	dir, err := filepath.Abs(filepath.Dir(headPath)) // 获得存放 headPath 文件的目录
	if err != nil {
		return nil, err
	}
	head, err := OpenAutoFile(headPath)
	if err != nil {
		return nil, err
	}

	g := &Group{
		ID:                 "group:" + head.ID,
		Head:               head,
		headBuf:            bufio.NewWriterSize(head, 4096*10),
		Dir:                dir,
		headSizeLimit:      defaultHeadSizeLimit,
		totalSizeLimit:     defaultTotalSizeLimit,
		groupCheckDuration: defaultGroupCheckDuration,
		minIndex:           0,
		maxIndex:           0,
		doneProcessTicks:   make(chan struct{}),
	}

	gInfo := g.readGroupInfo()
	g.minIndex = gInfo.MinIndex
	g.maxIndex = gInfo.MaxIndex
	return g, nil
}

// Start 开启 processTicks goroutine 定期检查 Group 里的文件
func (g *Group) Start() {
	g.ticker = time.NewTicker(g.groupCheckDuration)
	go g.processTicks()
}

// Stop 关闭 processTicks goroutine
func (g *Group) Stop() error {
	g.ticker.Stop()
	close(g.doneProcessTicks)
	return g.FlushAndSync()
}

// WaitQuit 等待直到 processTicks goroutine 结束
func (g *Group) WaitQuit() {
	// wait for processTicks routine to finish
	<-g.doneProcessTicks
}

// Close 关闭 head 文件
func (g *Group) Close() {
	_ = g.FlushAndSync()
	g.mtx.Lock()
	_ = g.Head.closeFile()
	g.mtx.Unlock()
}

// HeadSizeLimit 返回当前 head 文件的最大大小限制
func (g *Group) HeadSizeLimit() int64 {
	g.mtx.Lock()
	defer g.mtx.Unlock()
	return g.headSizeLimit
}

// TotalSizeLimit 返回当前 Group 的最大大小限制
func (g *Group) TotalSizeLimit() int64 {
	g.mtx.Lock()
	defer g.mtx.Unlock()
	return g.totalSizeLimit
}

// MaxIndex 返回 Group 里最后一个文件的索引
func (g *Group) MaxIndex() int {
	g.mtx.Lock()
	defer g.mtx.Unlock()
	return g.maxIndex
}

// MinIndex 返回 Group 里第一个文件的索引
func (g *Group) MinIndex() int {
	g.mtx.Lock()
	defer g.mtx.Unlock()
	return g.minIndex
}

// Write 将 p  写入 Group 的 head 文件里。
// 它返回写入的字节数。如果 nn < len(p)，它还
// 返回一个 error。
// 注意:由于被写入被缓冲区，所以它们不会同步写入
func (g *Group) Write(p []byte) (nn int, err error) {
	g.mtx.Lock()
	defer g.mtx.Unlock()
	return g.headBuf.Write(p)
}

// FlushAndSync 将 head 缓冲区里的内容刷新到底层的 file 里，然后将 file 里的内容同步到硬盘上
func (g *Group) FlushAndSync() error {
	g.mtx.Lock()
	defer g.mtx.Unlock()
	err := g.headBuf.Flush()
	if err == nil {
		err = g.Head.Sync()
	}
	return err
}

// processTicks 每隔一段时间调用以下两个方法：
// .checkHeadSizeLimit() 和 .checkTotalSizeLimit()
func (g *Group) processTicks() {
	for {
		select {
		case <-g.ticker.C:
			g.checkHeadSizeLimit()
			g.checkTotalSizeLimit()
		case <-g.doneProcessTicks:
			return
		}
	}
}

// 注意: 此函数在测试中手动调用。
func (g *Group) checkHeadSizeLimit() {
	limit := g.HeadSizeLimit()
	if limit == 0 {
		return
	}
	size, err := g.Head.Size()
	if err != nil {
		return
	}
	if size >= limit {
		g.RotateFile()
	}
}

func (g *Group) checkTotalSizeLimit() {
	limit := g.TotalSizeLimit() // 获取到 Group 的最大容量限制
	if limit == 0 {
		return
	}

	gInfo := g.readGroupInfo()
	totalSize := gInfo.TotalSize            // 获取当前 Group 的总大小
	for i := 0; i < maxFilesToRemove; i++ { // 如果当前 Group 的大小大于总量限制，那么一次最多可以删除 4 个文件
		index := gInfo.MinIndex + i
		if totalSize < limit { // 如果当前 Group 的大小小于最大限制，那么就 OK，直接返回
			return
		}
		if index == gInfo.MaxIndex {
			return
		}
		pathToRemove := filePathForIndex(g.Head.Path, index, gInfo.MaxIndex)
		fInfo, err := os.Stat(pathToRemove)
		if err != nil {
			continue
		}
		err = os.Remove(pathToRemove) // 删除文件
		if err != nil {
			return
		}
		totalSize -= fInfo.Size() // 给 Group 腾点空间
	}
}

// RotateFile 关闭当前的 head 文件，然后重新创建一个 head 文件句柄，之前的 head 文件只是重新命了个名字
func (g *Group) RotateFile() {
	g.mtx.Lock()
	defer g.mtx.Unlock()

	headPath := g.Head.Path

	if err := g.headBuf.Flush(); err != nil {
		panic(err)
	}

	if err := g.Head.Sync(); err != nil {
		panic(err)
	}

	if err := g.Head.closeFile(); err != nil {
		panic(err)
	}

	indexPath := filePathForIndex(headPath, g.maxIndex, g.maxIndex+1)
	if err := os.Rename(headPath, indexPath); err != nil {
		// 给当前的 head 文件重新命个名字，这样下次打开 head 文件时，由于之前的 head 文件重命名，所以系统找不到 headPath 指向的文件，那么就重新打开一个 head 文件
		panic(err)
	}

	g.maxIndex++
}

// NewReader 返回一个新的 Group Reader
// 注意: 调用者必须关闭返回的 GroupReader.
func (g *Group) NewReader(index int) (*GroupReader, error) {
	r := newGroupReader(g)
	err := r.SetIndex(index)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// GroupInfo holds information about the group.
type GroupInfo struct {
	MinIndex  int   // Group 里的第一个文件索引，包含 head
	MaxIndex  int   // Group 里的最后一个文件索引，包含 head
	TotalSize int64 // Group 的总大小
	HeadSize  int64 // Group 里 head 文件的大小
}

// ReadGroupInfo 在扫描了 Group 里所有文件之后，返回一个 GroupInfo
func (g *Group) ReadGroupInfo() GroupInfo {
	g.mtx.Lock()
	defer g.mtx.Unlock()
	return g.readGroupInfo()
}

// Index includes the head.
// CONTRACT: caller should have called g.mtx.Lock
func (g *Group) readGroupInfo() GroupInfo {
	groupDir := filepath.Dir(g.Head.Path)  // Group 的目录
	headBase := filepath.Base(g.Head.Path) // head 文件的名字
	var minIndex, maxIndex int = -1, -1
	var totalSize, headSize int64 = 0, 0

	dir, err := os.Open(groupDir) // 打开 Group 的目录
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = dir.Close()
	}()
	fiz, err := dir.Readdir(0) // 读取 Group 里所有文件，返回 []FileInfo 和一个 error
	if err != nil {
		panic(err)
	}

	// 扫描 Group 里所有文件
	for _, fileInfo := range fiz {
		if fileInfo.Name() == headBase { // 如果是 head 文件
			fileSize := fileInfo.Size()
			totalSize += fileSize // Group 的总大小累加上 head 文件的大小
			headSize = fileSize   // 记录下 head 文件的大小
			continue
		} else if strings.HasPrefix(fileInfo.Name(), headBase) { // 如果是非 head 文件
			fileSize := fileInfo.Size()
			totalSize += fileSize
			indexedFilePattern := regexp.MustCompile(`^.+\.([0-9]{3,})$`)
			submatch := indexedFilePattern.FindSubmatch([]byte(fileInfo.Name()))
			if len(submatch) != 0 {
				// Matches
				fileIndex, err := strconv.Atoi(string(submatch[1]))
				if err != nil {
					panic(err)
				}
				if maxIndex < fileIndex {
					maxIndex = fileIndex
				}
				if minIndex == -1 || fileIndex < minIndex {
					minIndex = fileIndex
				}
			}
		}
	}

	if minIndex == -1 {
		// 如果此时 Group 里只有一个 head 文件，那么 head 文件的索引就是 0
		minIndex, maxIndex = 0, 0
	} else {
		maxIndex++
	}
	return GroupInfo{minIndex, maxIndex, totalSize, headSize}
}

func filePathForIndex(headPath string, index int, maxIndex int) string {
	if index == maxIndex {
		// 如果 index 等于 maxIndex，则直接返回 head 文件的 path
		return headPath
	}
	return fmt.Sprintf("%v.%03d", headPath, index)
}

//--------------------------------------------------------------------------------

// GroupReader 提供一个从 Group 里读取内容的接口
type GroupReader struct {
	*Group
	mtx       sync.Mutex
	curIndex  int
	curFile   *os.File
	curReader *bufio.Reader
	curLine   []byte
}

func newGroupReader(g *Group) *GroupReader {
	return &GroupReader{
		Group:     g,
		curIndex:  0,
		curFile:   nil,
		curReader: nil,
		curLine:   nil,
	}
}

// Close closes the GroupReader by closing the cursor file.
func (gr *GroupReader) Close() error {
	gr.mtx.Lock()
	defer gr.mtx.Unlock()

	if gr.curReader != nil {
		err := gr.curFile.Close()
		gr.curIndex = 0
		gr.curReader = nil
		gr.curFile = nil
		gr.curLine = nil
		return err
	}
	return nil
}

// Read implements io.Reader, reading bytes from the current Reader
// incrementing index until enough bytes are read.
func (gr *GroupReader) Read(p []byte) (n int, err error) {
	lenP := len(p)
	if lenP == 0 {
		return 0, errors.New("given empty slice")
	}

	gr.mtx.Lock()
	defer gr.mtx.Unlock()

	// 如果还没打开文件，就打开当前索引处的文件
	if gr.curReader == nil {
		if err = gr.openFile(gr.curIndex); err != nil {
			return 0, err
		}
	}

	// 迭代的读取 Group 里的文件，直到读满 p
	var nn int
	for {
		nn, err = gr.curReader.Read(p[n:])
		n += nn
		switch {
		case err == io.EOF: // 如果读到当前文件的末尾了
			if n >= lenP {
				return n, nil
			}
			// 如果没读满 p，就继续读取下一个文件
			if err1 := gr.openFile(gr.curIndex + 1); err1 != nil {
				return n, err1
			}
		case err != nil: // 如果读取文件的时候出错了
			return n, err
		case nn == 0: // 如果读取的文件是一个空文件
			return n, err
		}
	}
}

// openFile 如果给的要打开的文件索引大于 Group 的 maxIndex，则说明读取 Group 完了，那么就返回一个 io.EOF
func (gr *GroupReader) openFile(index int) error {
	// 锁定 Group，确保在此期间 head 文件不动
	gr.Group.mtx.Lock()
	defer gr.Group.mtx.Unlock()

	if index > gr.Group.maxIndex {
		return io.EOF
	}

	curFilePath := filePathForIndex(gr.Head.Path, index, gr.Group.maxIndex) // 根据给的文件索引返回指定的文件名
	curFile, err := os.OpenFile(curFilePath, os.O_RDONLY|os.O_CREATE, autoFilePerms)
	if err != nil {
		return err
	}
	curReader := bufio.NewReader(curFile)

	// 关闭当前的文件句柄
	if gr.curFile != nil {
		gr.curFile.Close()
	}
	gr.curIndex = index
	gr.curFile = curFile // 更新当前的文件句柄
	gr.curReader = curReader
	gr.curLine = nil
	return nil
}

// CurIndex 返回当前的文件索引
func (gr *GroupReader) CurIndex() int {
	gr.mtx.Lock()
	defer gr.mtx.Unlock()
	return gr.curIndex
}

// SetIndex 设置当前的文建索引，并将当前的文件句柄更新到当前文件索引处的文件
func (gr *GroupReader) SetIndex(index int) error {
	gr.mtx.Lock()
	defer gr.mtx.Unlock()
	return gr.openFile(index)
}
