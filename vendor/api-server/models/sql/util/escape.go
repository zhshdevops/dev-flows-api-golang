/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-11-08  @author Zhao Shuailong
 */

package util

// EscapeStringBackslash escapes special characters.
func EscapeStringBackslash(v string) string {
	buf := make([]byte, len(v)*2, len(v)*2)
	pos := 0

	for i := 0; i < len(v); i++ {
		c := v[i]
		switch c {
		case '\x00':
			buf[pos] = '\\'
			buf[pos+1] = '0'
			pos += 2
		case '\n':
			buf[pos] = '\\'
			buf[pos+1] = 'n'
			pos += 2
		case '\r':
			buf[pos] = '\\'
			buf[pos+1] = 'r'
			pos += 2
		case '\x1a':
			buf[pos] = '\\'
			buf[pos+1] = 'Z'
			pos += 2
		case '\'':
			buf[pos] = '\\'
			buf[pos+1] = '\''
			pos += 2
		case '"':
			buf[pos] = '\\'
			buf[pos+1] = '"'
			pos += 2
		case '\\':
			buf[pos] = '\\'
			buf[pos+1] = '\\'
			pos += 2
		default:
			buf[pos] = c
			pos++
		}
	}

	return string(buf[:pos])
}

func EscapeSliceForInjection(src []string) []string {
	out := make([]string, 0, len(src))
	for i := range src {
		out = append(out, EscapeStringBackslash(src[i]))
	}
	return out
}

func EscapeUnderlineInLikeStatement(v string) string {
	v = EscapeStringBackslash(v)
	buf := make([]byte, len(v)*2, len(v)*2)
	pos := 0
	for i := 0; i < len(v); i++ {
		if v[i] == '_' {
			buf[pos] = '\\'
			buf[pos+1] = '_'
			pos += 2
		} else {
			buf[pos] = v[i]
			pos++
		}
	}
	return string(buf[:pos])
}
