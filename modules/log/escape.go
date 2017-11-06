/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 */

package log

import (
	"regexp"
)

// Lucene supports escaping special characters that are part of the query syntax. The current list special characters are
// + - && || ! ( ) { } [ ] ^ " ~ * ? : \
func escape(keyWord string) string {
	re, err := regexp.Compile(`\+|-|&&|\|\||!|\(|\)|\{|}|\[|]|\^|"|~|\*|\?|:|\\| `)

	if err != nil {
		panic(err)
	}

	return re.ReplaceAllStringFunc(keyWord, replace)
}

func replace(ec string) string {
	return `\\` + ec
}
