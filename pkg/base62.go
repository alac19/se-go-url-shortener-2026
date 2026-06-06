// Package base62 provides the function of converting integers into Base62 encoded strings.
package base62

// IntToBase62 将非负整数转换为 Base62 字符串。
func IntToBase62(num uint64) string {
	var s []byte
	const CharacterSet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	for {
		b := num % 62

		num /= 62

		s = append(s, byte(CharacterSet[b]))

		if num == 0 {
			break
		}
	}

	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}

	return string(s)
}
