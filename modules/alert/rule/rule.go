/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 * 2017-03-06  @author zhangyongkang
 */
package rule

import (
	"bytes"
	"errors"
	"fmt"

	"strings"

	"strconv"

	"github.com/golang/glog"
	"github.com/prometheus/prometheus/promql"
)

// AlertRule responsible for an alert rule record
type AlertRule struct {
	Name       string            `json:"name"`
	Condition  Condition         `json:"condition"`
	Interval   string            `json:"interval"`
	Labels     map[string]string `json:"labels"`
	Annotation map[string]string `json:"annotation"`
}

// Condition responsible for alert rule condition
type Condition struct {
	Target    string `json:"target"`
	Operation string `json:"operation"`
	Threshold string `json:"threshold"`
}

func (c *Condition) isValid() bool {
	return c.Target != "" &&
		c.Threshold != "" &&
		c.Operation != ""
}

func (c *Condition) String() string {
	return "IF (" + c.Target + ") " + c.Operation + " " + c.Threshold
}

// IsValid check if AlertRule is valid
func (a *AlertRule) IsValid() bool {
	return a.Name != "" &&
		a.Condition.isValid()
}

// String return a rule record. demo:
// ALERT gorountine_less_1 IF (go_goroutines{device_ID="local",instance="192.168.0.66:9100",job="node"}) > 1 FOR 5m LABELS{ go_label = "gorountine_less_1",alter = "goroutine",}ANNOTATIONS {summary = "Instance {{ $labels.instance }} goroutine is dangerously high", description = "This device's goroutine is has exceeded the threshold with a value of {{ $value }}.
func (a *AlertRule) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("ALERT " + a.Name + " ")
	buffer.WriteString(a.Condition.String() + " ")
	if a.Interval != "" {
		buffer.WriteString("FOR " + a.Interval + " ")
	}
	if len(a.Labels) > 0 {
		buffer.WriteString("LABELS{ ")
		for k, v := range a.Labels {
			buffer.WriteString(k + `="` + v + `", `)
		}
		buffer.WriteString("} ")
	}
	if len(a.Annotation) > 0 {
		buffer.WriteString("ANNOTATIONS {")
		for k, v := range a.Annotation {
			buffer.WriteString(k + `="` + v + `", `)
		}
		buffer.WriteString("}")
	}
	return buffer.String()
}

// Pretty print AlertRule in well format
func (a *AlertRule) Pretty() string {
	raw := a.String()
	sts, err := StringToAlertStmts(raw)
	if err != nil {
		panic(err)
	}
	if len(sts) != 1 {
		panic("single rule txt should parsed to single AlertStmt")
	}
	return sts[0].String()
}

// MatchLables check if the AlertRule match labels
func (a *AlertRule) MatchLables(labels map[string]string, matchLable bool) bool {
	if a.Labels == nil {
		return false
	}
	for k, v := range labels {
		val, found := a.Labels[k]
		if !found {
			return false
		}
		if matchLable {
			if v != val {
				return false
			}
		}
	}
	return true
}

// ASTToAlertRule convert *promql.AlertStmt to *AlertRule
func ASTToAlertRule(at *promql.AlertStmt) (*AlertRule, error) {
	ar := &AlertRule{Name: at.Name}
	if at.Duration > 0 {
		ar.Interval = strconv.FormatFloat(at.Duration.Minutes(), 'f', 0, 64) + "m"
	}
	expr, ok := at.Expr.(*promql.BinaryExpr)
	if !ok {
		return nil, fmt.Errorf("Expr convert to BinaryExpr failed")
	}

	ar.Condition.Operation = expr.Op.String()
	ar.Condition.Threshold = expr.RHS.String()
	target := expr.LHS.String()
	if strings.HasPrefix(target, "(") && strings.HasSuffix(target, ")") {
		target = target[1 : len(target)-1]
	}
	ar.Condition.Target = target
	// convert Labels and Annotation
	ar.Labels = make(map[string]string, len(at.Labels))
	for k, v := range at.Labels {
		ar.Labels[string(k)] = string(v)
	}
	ar.Annotation = make(map[string]string, len(at.Annotations))
	for k, v := range at.Annotations {
		ar.Annotation[string(k)] = string(v)
	}
	return ar, nil
}

// StringToAlertStmts parse string into []*promql.AlertStmt
func StringToAlertStmts(input string) ([]*promql.AlertStmt, error) {
	if input == "" {
		return nil, errors.New("empty input")
	}
	sts, err := promql.ParseStmts(input)
	if err != nil {
		glog.Errorf("StringToAlertStmts ParseStmts failed: %v, input: %s", err, input)
		return nil, err
	}

	asts := make([]*promql.AlertStmt, 0)
	for _, st := range sts {
		if v, ok := st.(*promql.AlertStmt); ok {
			asts = append(asts, v)
		}
	}
	if len(asts) < 1 {
		return nil, errors.New("there is no alter rule")
	}
	return asts, nil
}

// StringToAlertRules parse string to []*AlertRule
func StringToAlertRules(input string) ([]*AlertRule, error) {
	asts, err := StringToAlertStmts(input)
	if err != nil {
		return nil, err
	}
	ars := make([]*AlertRule, 0, len(asts))
	for _, ast := range asts {
		ar, err := ASTToAlertRule(ast)
		if err != nil {
			return nil, err
		}
		ars = append(ars, ar)
	}
	return ars, nil
}
