/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 */

package log

import "testing"

func TestEscape(t *testing.T) {
	input := `a b+c-d&&e||f!g(h)i{j}k[l]m^n"o~p*q?r:s\t`
	except := `a\\ b\\+c\\-d\\&&e\\||f\\!g\\(h\\)i\\{j\\}k\\[l\\]m\\^n\\"o\\~p\\*q\\?r\\:s\\\t`
	got := escape(input)
	if except != got {
		t.Errorf(`escape("%s") should be "%s", but got "%s"`, input, except, got)
	} else {
		t.Logf(`input: "%s", except: "%s".`, input, except)
	}
}
