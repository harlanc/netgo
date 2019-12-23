package lib

import (
	"encoding/binary"
	"strconv"

	uuid "github.com/satori/go.uuid"
)

//String2Int convert from string to int
func String2Int(val string) int {

	goodsIDInt, err := strconv.Atoi(val)
	if err != nil {
		return -1
	}
	return goodsIDInt
}

//Int2String convert from int to string
func Int2String(val int) string {
	return strconv.Itoa(val)
}

//Int642String convert from int64 to string
func Int642String(val int64) string {
	return strconv.FormatInt(val, 10)
}

//Uint322String convert from uint32 to string
func Uint322String(val uint32) string {
	return strconv.FormatUint(uint64(val), 10)
}

//Float642String convert from int to string
func Float642String(val float64) string {
	return strconv.FormatFloat(val, 'E', -1, 64)
}

//Uint322ByteArray uint32 to byte array
func Uint322ByteArray(val uint32) []byte {

	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, val)
	return b
}

//NewUUID get a new UUID
func NewUUID() string {
	uuid, err := uuid.NewV4()
	if err != nil {
		return ""
	}
	return uuid.String()

}
