package ulog

import "strconv"

var char = struct {
	digit   [10]byte
	newline byte
	colon   byte
	equal   byte
	hyphen  byte
	period  byte
	quote   byte
	space   byte
	T       byte
	Z       byte
}{
	digit:   [10]byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'},
	newline: '\n',
	colon:   ':',
	equal:   '=',
	hyphen:  '-',
	period:  '.',
	quote:   '"',
	space:   ' ',
	T:       'T',
	Z:       'Z',
}

var buf = struct {
	digit   [10][]byte
	digits2 [100][]byte
	newline []byte
}{
	newline: []byte{'\n'},
}

func init() {
	buf.digit[0] = []byte("0")
	buf.digit[1] = []byte("1")
	buf.digit[2] = []byte("2")
	buf.digit[3] = []byte("3")
	buf.digit[4] = []byte("4")
	buf.digit[5] = []byte("5")
	buf.digit[6] = []byte("6")
	buf.digit[7] = []byte("7")
	buf.digit[8] = []byte("8")
	buf.digit[9] = []byte("9")
	buf.digits2[0] = []byte("00")
	buf.digits2[1] = []byte("01")
	buf.digits2[2] = []byte("02")
	buf.digits2[3] = []byte("03")
	buf.digits2[4] = []byte("04")
	buf.digits2[5] = []byte("05")
	buf.digits2[6] = []byte("06")
	buf.digits2[7] = []byte("07")
	buf.digits2[8] = []byte("08")
	buf.digits2[9] = []byte("09")

	for i := 10; i <= 99; i++ {
		buf.digits2[i] = []byte(strconv.FormatInt(int64(i), 10))
	}
}
