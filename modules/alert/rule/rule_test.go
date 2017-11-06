/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 * 2017-03-06  @author zhangyongkang
 */
package rule

import "testing"

import "reflect"

func TestAlertRuleSummary(t *testing.T) {
	input := `ALERT gorountine_less_3 IF(go_goroutines{device_ID="local",instance="192.168.0.66:9100",job="node~~~",}) > 3 FOR 5m LABELS{ go_label = "gorountine_less_3",alter = "goroutine"} ANNOTATIONS {summary = "Instance {{ $labels.instance }} goroutine is dangerously high", description = "This device's goroutine is has exceeded the threshold with a value of {{ $value }}."}`
	asts, err := StringToAlertStmts(input)
	if err != nil {
		t.Fatalf("StringToAlertStmts failed: %v\n", err)
	}
	if len(asts) != 1 {
		t.Fatalf("StringToAlertStmts: asts should be one record get %d\n", len(asts))
	}
	ars, err := StringToAlertRules(input)
	if err != nil {
		t.Fatalf("StringToAlertRules failed: %v\n", err)
	}
	if len(ars) != 1 {
		t.Fatalf("StringToAlertRules failed, ars should be one record get %d\n", len(ars))
	}
	astStr := asts[0].String()
	arsStr := ars[0].String()
	arsPretty := ars[0].Pretty()

	strArs, err := StringToAlertRules(arsStr)
	if err != nil {
		t.Fatalf("StringToAlertRules failed: %v\n", err)
	}
	strArs2, err := StringToAlertRules(astStr)
	if err != nil {
		t.Fatalf("StringToAlertRules failed: %v\n", err)
	}
	prettyArs, err := StringToAlertRules(arsPretty)
	if err != nil {
		t.Fatalf("StringToAlertRules failed: %v\n", err)
	}
	if !reflect.DeepEqual(ars, strArs) || !reflect.DeepEqual(ars, prettyArs) || !reflect.DeepEqual(ars, strArs2) {
		t.Fatalf("string convert to struct failed, or reverse")
	}

	ar, err := ASTToAlertRule(asts[0])
	if err != nil {
		t.Fatalf("ASTToAlertRule failed: %v\n", err)
	}
	if !reflect.DeepEqual(ar, ars[0]) {
		t.Fatalf("ar(%#v) and ars[0] (%#v) should be equal but not", ar, ars[0])
	}
	// []*AlertRule to string
	ar.Name = "alertTwo"
	twoAlertRuleStr := ar.Pretty() + "\n" + ars[0].Pretty()
	twoAlertRule, err := StringToAlertRules(twoAlertRuleStr)
	if err != nil {
		t.Fatalf("convert two alert rule to string failed: %v\n", err)
	}
	if len(twoAlertRule) != 2 {
		t.Fatalf("expect len(twoAlertRule) = 2 but not (%d)\n", len(twoAlertRule))
	}
	// ars = append(ars, ar)
}
