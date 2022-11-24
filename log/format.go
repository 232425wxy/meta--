// format.go
// 这个文件定义了打印日志的几种格式：控制台格式、JSON格式以及普通日志格式，它们的格式样例分别如下所示：

// 控制台格式：INFO [01-01|00:00:00.000] start service                            blockchain=meta--

// JSON格式：{"blockchain":"meta--","lvl":"info","msg":"start service","t":"0001-01-01T00:00:00Z"}

// 普通日志格式：t=0001-01-01T00:00:00Z lvl=trce msg="trace logger" blockchain=meta--

// 此外，还支持在调试项目时定位到代码的位置，如果有这个需求，那么只能在以控制台格式打印日志信息时有效：

// TRACE[01-01|00:00:00.000|/log/log_test.go:24] trace logger                             blockchain=meta--

package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// API 项目级可导出函数

// PrintOrigins ♏ | (o゜▽゜)o☆吴翔宇
//
// PrintOrigins 接受一个bool类型的数据作为输入参数，该方法是一个开关函数，如果传入的参数等于true，那
// 么在以后打印日志信息，会打印出输出日志信息所在的代码文件和代码行，类似于："file:line"。
func PrintOrigins(enabled bool) {
	if enabled {
		atomic.StoreUint32(&locationEnabled, 1)
	} else {
		atomic.StoreUint32(&locationEnabled, 0)
	}
}

// TerminalFormat ♏ | (o゜▽゜)o☆吴翔宇
//
// TerminalFormat 控制台打印格式：
//
//	INFO [01-01|00:00:00.000] start service                            blockchain=meta--
func TerminalFormat(useColor bool) formatter {
	return formatFunc(func(record *Record) []byte {
		var color = 0
		if useColor {
			switch record.Lvl {
			case LvlCrit:
				color = 35 // 紫色
			case LvlError:
				color = 31 // 红色
			case LvlWarn:
				color = 33 // 黄色
			case LvlInfo:
				color = 32 // 绿色
			case LvlDebug:
				color = 36 // 蓝绿色
			case LvlTrace:
				color = 34 // 蓝色
			}
		}
		buffer := new(bytes.Buffer)
		// TRACE DEBUG INFO WARN ERROR CRIT
		lvl := record.Lvl.AlignedString()
		if atomic.LoadUint32(&locationEnabled) != 0 {
			// 需要在每一条日志前加上输出日志的代码位置
			location := fmt.Sprintf("%+v", record.Call)
			for _, prefix := range locationTrims {
				location = strings.TrimPrefix(location, prefix)
			}
			align := int(atomic.LoadUint32(&locationLength))
			if align < len(location) {
				align = len(location)
				atomic.StoreUint32(&locationLength, uint32(align))
			}
			padding := strings.Repeat(" ", align-len(location))
			// 上面的代码都是为了打印输出日志信息的代码位置做准备

			if color > 0 {
				_, _ = fmt.Fprintf(buffer, "\x1b[%dm%s\x1b[0m[%s|%s]%s %s ", color, lvl, record.Time.Format(termTimeFormat), location, padding, record.Msg)
			} else {
				_, _ = fmt.Fprintf(buffer, "%s[%s|%s]%s %s ", lvl, record.Time.Format(termTimeFormat), location, padding, record.Msg)
			}
		} else {
			if color > 0 {
				_, _ = fmt.Fprintf(buffer, "\x1b[%dm%s\x1b[0m[%s] %s ", color, lvl, record.Time.Format(termTimeFormat), record.Msg)
			} else {
				_, _ = fmt.Fprintf(buffer, "%s[%s] %s ", lvl, record.Time.Format(termTimeFormat), record.Msg)
			}
		}
		length := utf8.RuneCountInString(record.Msg)
		if len(record.Ctx) > 0 && length < termMsgJust {
			// 如果此条日志记录需要打印键值对信息，且日志消息长度小于40，那么就补齐长度到40，再在后面加上键值对信息
			buffer.Write(bytes.Repeat([]byte{' '}, termMsgJust-length))
		}
		logfmt(buffer, record.Ctx, color, true)
		return buffer.Bytes()
	})
}

// LogfmtFormat ♏ | (o゜▽゜)o☆吴翔宇
//
// LogfmtFormat 普通日志打印格式：
//
//	t=2022-11-22T19:51:30+08:00 lvl=info msg="Start network" app=ethereum/server consensus=POS
func LogfmtFormat() formatter {
	return formatFunc(func(record *Record) []byte {
		common := []interface{}{record.KeyNames.Time, record.Time, record.KeyNames.Lvl, record.Lvl, record.KeyNames.Msg, record.Msg}
		buf := new(bytes.Buffer)
		logfmt(buf, append(common, record.Ctx...), 0, false)
		return buf.Bytes()
	})
}

// JSONFormat ♏ | (o゜▽゜)o☆吴翔宇
//
// JSONFormat 以JSON格式打印日志：
//
//	{"app":"ethereum/server","consensus":"POS","lvl":"info","msg":"Start network","t":"2022-11-22T16:08:06.96890076+08:00"}
func JSONFormat() formatter {
	jsonMarshal := json.Marshal

	return formatFunc(func(record *Record) []byte {
		props := make(map[string]interface{})
		props[record.KeyNames.Time] = record.Time
		props[record.KeyNames.Lvl] = record.Lvl.String()
		props[record.KeyNames.Msg] = record.Msg

		for i := 0; i < len(record.Ctx); i += 2 {
			k, ok := record.Ctx[i].(string)
			if !ok {
				props[errorKey] = fmt.Sprintf("%+v is not a string key", record.Ctx[i])
			}
			props[k] = formatJSONValue(record.Ctx[i+1])
		}
		bz, err := jsonMarshal(props)
		if err != nil {
			// 一般来讲是不会出错的
			bz, _ = jsonMarshal(map[string]string{
				errorKey: err.Error(),
			})
			return bz
		}
		bz = append(bz, '\n')
		return bz
	})
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 辅助变量

// formatter ♏ | (o゜▽゜)o☆吴翔宇
//
// formatter 接口定义了日志打印格式方法。
type formatter interface {
	format(r *Record) []byte
}

type formatFunc func(*Record) []byte

func (f formatFunc) format(r *Record) []byte {
	return f(r)
}

// TerminalStringer ♏ | (o゜▽゜)o☆吴翔宇
//
// TerminalStringer 有的类型对象在日志里面输出的时候会有自己独特的输出格式，比如节点的ID号，一般来讲，
// 节点的ID号都比较长，直接输出到日志里比较难看，那这个时候，我们让节点的ID类型实现 TerminalStringer
// 接口里的方法，自定义打印格式，那么就可以按照自己喜欢的格式打印ID号了。
type TerminalStringer interface {
	TerminalString() string
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 不可导出的工具函数

// logfmt ♏ | (o゜▽゜)o☆吴翔宇
//
// logfmt 方法的目的是将日志条目里的键值对对齐输入到一个缓冲区里，即该方法接受的第一个参数，然后根据给定的颜色，对键
// 值对的键值上色，一般来讲，传入的第三个参数用来指定打印键值对时键的颜色，这个颜色一般由日志等级决定，比如如果日志等
// 级是 LvlCrit，则颜色就是紫色。
func logfmt(buf *bytes.Buffer, ctx []interface{}, color int, term bool) {
	for i := 0; i < len(ctx); i += 2 {
		if i != 0 {
			// 加一个空格
			buf.WriteByte(' ')
		}

		k, ok := ctx[i].(string) // 键最好是string类型的
		v := formatLogfmtValue(ctx[i+1], term)
		if !ok {
			k, v = errorKey, formatLogfmtValue(k, term)
		}

		fieldPaddingLock.RLock()
		padding := fieldPadding[k]
		fieldPaddingLock.RUnlock()

		// 一个汉字占用3个字节，但是一个汉字也就是一个字符，如果用len方法去计算字符串长度，返回的
		// 结果是是字节数量，但是我们想要的是字符数量，这样才容易对齐
		length := utf8.RuneCountInString(v)
		if padding < length && length <= termCtxMaxPadding {
			padding = length
			fieldPaddingLock.Lock()
			fieldPadding[k] = padding
			fieldPaddingLock.Unlock()
		}

		// 输入日志信息里的键值对
		if color > 0 {
			_, _ = fmt.Fprintf(buf, "\x1b[%dm%s\x1b[0m=", color, k)
		} else {
			buf.WriteString(k)
			buf.WriteByte('=')
		}
		buf.WriteString(v)

		// 之所以要求i小于len(ctx)-2，是因为最后一对键值对就没必要保持对齐啦
		if i < len(ctx)-2 && padding > length {
			// 保持日志里的键值对对齐
			buf.Write(bytes.Repeat([]byte{' '}, padding-length))
		}
	}
	buf.WriteByte('\n')
}

func formatLogfmtValue(value interface{}, term bool) string {
	if value == nil {
		return "nil"
	}

	switch v := value.(type) {
	case time.Time:
		return v.Format(timeFormat)

	case *big.Int:
		if v == nil {
			return "<nil>"
		}
		return formatLogfmtBigInt(v)
	}
	if term {
		if s, ok := value.(TerminalStringer); ok {
			// 用户自定义在终端输出的字符串格式，这个还是很有用的，比如用户可以自定义ID的输出长度是多少
			return escapeString(s.TerminalString())
		}
	}
	value = formatShared(value)
	switch v := value.(type) {
	case bool:
		return strconv.FormatBool(v)
	case float32:
		return strconv.FormatFloat(float64(v), floatFormat, 3, 64)
	case float64:
		return strconv.FormatFloat(v, floatFormat, 3, 64)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case uint8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case uint16:
		return strconv.FormatInt(int64(v), 10)
	// 大整数需要用千位分隔符进行分割
	case int:
		if v < 0 {
			return formatLogfmtUint64(uint64(-v), true)
		} else {
			return formatLogfmtUint64(uint64(v), false)
		}
	case int32:
		if v < 0 {
			return formatLogfmtUint64(uint64(-v), true)
		} else {
			return formatLogfmtUint64(uint64(v), false)
		}
	case int64:
		if v < 0 {
			return formatLogfmtUint64(uint64(-v), true)
		} else {
			return formatLogfmtUint64(uint64(v), false)
		}
	case uint:
		return formatLogfmtUint64(uint64(v), false)
	case uint32:
		return formatLogfmtUint64(uint64(v), false)
	case uint64:
		return formatLogfmtUint64(v, false)
	case string:
		return escapeString(v)
	default:
		return escapeString(fmt.Sprintf("%+v", value))
	}
}

// formatShared ♏ | (o゜▽゜)o☆吴翔宇
//
// formatShared 方法接受一个interface{}类型的value作为输入参数，value的底层类型属于以下三种类型，则会
// 做如下处理：
//  1. time.Time 类型：转换时间值的格式为"2006-01-02T15:04:05Z07:00"，得到输出例如为：2022-11-22T14:45:04+0800
//  2. error 类型：返回error.Error() string
//  3. 实现了 String() 方法的对象，返回其 String() 方法的返回值
//  4. 其他类型：不做处理，返回其原始值。
func formatShared(value interface{}) (result interface{}) {
	defer func() {
		if err := recover(); err != nil {
			if v := reflect.ValueOf(value); v.Kind() == reflect.Ptr && v.IsNil() {
				result = "nil"
			} else {
				panic(err)
			}
		}
	}()

	switch v := value.(type) {
	case time.Time:
		return v.Format(timeFormat)

	case error:
		return v.Error()

	case fmt.Stringer:
		return v.String()

	default:
		return v
	}
}

// formatJSONValue ♏ | (o゜▽゜)o☆吴翔宇
//
// formatJSONValue 方法接受一个interface{}类型的value作为参数，如果value的底层类型是数字类型
// 或字符串类型，就返回其原始值，如果是某个结构体类型，就会按照JSON格式将value完整地输出出来，包括
// 结构体的字段名。
func formatJSONValue(value interface{}) interface{} {
	value = formatShared(value)
	switch value.(type) {
	case int, int8, int16, int32, int64, float32, float64, uint, uint8, uint16, uint32, uint64, string:
		return value
	default:
		return fmt.Sprintf("%+v", value)
	}
}

// formatLogfmtUint64 ♏ | (o゜▽゜)o☆吴翔宇
//
// formatLogfmtUint64 方法接受两个参数，第一个参数是一个uint64类型的整数，第二个参数是一个bool值，用来
// 指示整数是否是负数，该方法的作用就是对整数进行格式化输出，例如以下两个例子：
//  1. 如果给定的两个参数为：1234, false，得到的输出是：1234
//  2. 如果给定的两个参数为：1234567890, true，得到的输出是：-1,234,567,890
func formatLogfmtUint64(n uint64, neg bool) string {
	if n < 100000 {
		if neg {
			return strconv.Itoa(-int(n))
		} else {
			return strconv.Itoa(int(n))
		}
	}

	const maxLength = 26

	out := make([]byte, maxLength)
	i := maxLength - 1
	comma := 0

	for ; n > 0; i-- {
		if comma == 3 {
			comma = 0
			out[i] = ','
		} else {
			comma++
			out[i] = '0' + byte(n%10)
			n /= 10
		}
	}
	if neg {
		out[i] = '-'
		i--
	}
	return string(out[i+1:])
}

// formatLogfmtBigInt ♏ | (o゜▽゜)o☆吴翔宇
//
// formatLogfmtBigInt 方法接受一个 *big.Int 作为输入参数，输入参数是一个大整数，该方法会修整一下
// 大整数的输出样式，具体做法就是在原先的大整数上增加千位分隔符：","。例如输入的大整数等于"12345678"，
// 则返回的值等于"12,345,678"。这个方法与 formatLogfmtBigUint64 方法的不同之处是该方法可以接受入
// 参为负整数。
func formatLogfmtBigInt(n *big.Int) string {
	if n.IsUint64() {
		return formatLogfmtUint64(n.Uint64(), false)
	}
	if n.IsInt64() {
		if n.Int64() < 0 {
			return formatLogfmtUint64(uint64(-n.Int64()), true)
		} else {
			return formatLogfmtUint64(uint64(n.Int64()), false)
		}
	}

	var (
		text  = n.String()
		buf   = make([]byte, len(text)+len(text)/3)
		comma = 0
		i     = len(buf) - 1
	)
	for j := len(text) - 1; j >= 0; j, i = j-1, i-1 {
		c := text[j]

		switch {
		case c == '-':
			buf[i] = c
		case comma == 3:
			buf[i] = ','
			i--
			comma = 0
			fallthrough
		default:
			buf[i] = c
			comma++
		}
	}
	return string(buf[i+1:])
}

// escapeString ♏ | (o゜▽゜)o☆吴翔宇
//
// escapeString 方法接受一个字符串作为输入参数，该方法如果发现给定的字符串中存在一些特殊字符，就会在
// 给定的字符串两端加上双引号，否则就什么也不做，将原字符串返回。以下字符被定义为特殊字符：
//  1. ASCII码小于0x22的字符
//  2. ASCII码大于7e的字符
//  3. '='字符
func escapeString(s string) string {
	needsQuoting := false
	for _, r := range s {
		if r <= '"' || r > '~' || r == '=' {
			needsQuoting = true
			break
		}
	}
	if !needsQuoting {
		return s
	}
	return strconv.Quote(s)
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 包级全局变量

// locationLength ♏ | (o゜▽゜)o☆吴翔宇
//
// locationLength 我们在调试程序时，希望在打印日志的同时。打印出输出日志信息的代码位置，这种位置类似于
// panic时打印的"file:line"，locationLength 就是在程序动态运行时自动的调整位置信息的字符串长度，因
// 为可能有的"file"字符串比较长，有的比较短，长短不一会导致打印出来的日志不好看，locationLength 就可以
// 帮助让日志信息里的位置信息长度对齐。
var locationLength uint32

// locationTrims ♏ | (o゜▽゜)o☆吴翔宇
//
// locationTrims 如果我们在输出日志信息时，要求定位到输出日志信息的文件位置，具体来说就是在哪个代码文件的
// 第多少行输出了[DEBUG]消息，如果显式的位置含有"github.com/232425wxy/meta--"字符串，则将其去除掉。
var locationTrims = []string{"github.com/232425wxy/meta--"}

// locationEnabled ♏ | (o゜▽゜)o☆吴翔宇
//
// locationEnabled 是一个开关，如果这个值不等于0，那么在输出日志信息时，会定位到输出日志信息的位置，具体来
// // 说就是在哪个文件的哪一行代码处输出了这个日志信息。
var locationEnabled uint32

// fieldPadding ♏ | (o゜▽゜)o☆吴翔宇
//
// fieldPadding 变量用于存储日志信息里键值对的宽度信息，为了在输出日志时保持左右对齐。
var fieldPadding = make(map[string]int)

// fieldPaddingLock ♏ | (o゜▽゜)o☆吴翔宇
//
// fieldPaddingLock 是一把锁，每次读取或改写 fieldPadding 时都要获取该锁，然后用完再释放。
var fieldPaddingLock sync.RWMutex

const (
	timeFormat        = time.RFC3339
	termTimeFormat    = "01-02|15:04:05.000"
	floatFormat       = 'f'
	termMsgJust       = 40
	termCtxMaxPadding = 40
)
