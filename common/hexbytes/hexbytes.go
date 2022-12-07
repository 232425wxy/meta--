package hexbytes

import (
	"encoding/hex"
	"fmt"
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 实现将字节切片和16进制字符串之间相互转换

type HexBytes []byte

// Marshal ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// 定义 Marshal 方法是为了和protobuf相兼容。
func (bz HexBytes) Marshal() ([]byte, error) {
	return bz, nil
}

// Unmarshal ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Unmarshal 方法是为了和protobuf相兼容。
func (bz *HexBytes) Unmarshal(data []byte) error {
	*bz = data
	return nil
}

// MarshalJSON ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// MarshalJSON 方法将字节切片编码成16进制字符串，然后在字符串两端加上双引号，就得到了JSON格式的字节切片，
// 编码之后数据长度会增加一倍以上。
func (bz HexBytes) MarshalJSON() ([]byte, error) {
	dst := make([]byte, 2*len(bz)+2)
	dst[0] = '"'
	s := hex.EncodeToString(bz)
	copy(dst[1:], s)
	dst[len(dst)-1] = '"'
	return dst, nil
}

// UnmarshalJSON ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// UnmarshalJSON 方法与MarshalJSON方法的作用正好相反。
func (bz *HexBytes) UnmarshalJSON(data []byte) error {
	if len(data) >= 2 && data[0] == '"' && data[len(data)-1] == '"' {
		content, err := hex.DecodeString(string(data[1 : len(data)-1]))
		if err != nil {
			return err
		}
		*bz = content
		return nil
	} else {
		return fmt.Errorf("failed to unmarshal data, because the given data is not valid: %q", string(data))
	}
}

// Bytes ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Bytes 返回原始字节切片的内容。
func (bz HexBytes) Bytes() []byte {
	return bz
}

// String ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// String 返回字节切片的16进制编码字符串。
func (bz HexBytes) String() string {
	return hex.EncodeToString(bz)
}

// CompatibleWith ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// CompatibleWith 判断两个字节切片中是否存在相同的字节，哪怕只是存在一个相同的字节，该方法都会
// 返回true。
func (bz HexBytes) CompatibleWith(other HexBytes) bool {
	for _, ch1 := range bz {
		for _, ch2 := range other {
			if ch1 == ch2 {
				return true
			}
		}
	}
	return false
}
