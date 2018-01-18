package filter

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strconv"

	"Service/log"
	"Service/request"
	"Service/units/rule/simplejson"
)

// 条件
type condition struct {
	Key  string   `json:"key"`
	Op   string   `json:"op"`
	Expr []string `json:"expr"`
}

type conditionlist struct { // combined conditions
	Combination string      `json:"combination"`
	Conditions  []condition `json:"conditions"`
}

// 动作
type action struct {
	Key  string   `json:"key"`
	Op   string   `json:"op"`
	Expr []string `json:"expr"`
}

type optionalaction struct {
	Percent uint64   `json:"percent"`
	Actions []action `json:"actions"`
}

type combinedoactions struct {
	Actions []optionalaction `json:"actions"`
}

type actionlist struct {
	Sactions []action           `json:"sactions"` // simple actions
	Oactions []combinedoactions `json:"oactions"` // optional actions
}

// 过滤
type filterbody struct {
	Condtions conditionlist `json:"conditions"`
	Actions   actionlist    `json:"actions"`
}

type filterImpl struct {
	Filters []filterbody `json:"filters"`
}

// 支持的格式
/*
"filter":[
	filter1,
	filter2,
	filter3
]

filter:[
	条件,
	动作
]

条件：
■[条件1],
[条件2]
■["or/and/not",
	[条件1],
	[条件2]
]
动作：
■[动作1]
■[
	[动作1],
	[动作2]
]
■[
	[80%,
		[动作1-1],
		[动作1-2]
	],
	[20%,
		[动作2-1],
		[动作2-2]
	]
]
■[
	[动作0],
	[动作1],
	[
		[80%,
			[动作2-1],
			[动作2-2]
		],
		[20%,
			[动作3-1],
			[动作3-2]
		]
	]
]
*/
func (f *filterImpl) Fill(raw []byte) (err error) {
	js, err := simplejson.NewJson(raw)
	if err != nil {
		return
	}
	fs, err := js.Array()
	if err != nil {
		return
	}
	for i := 0; i < len(fs); i++ {
		if fb, err := parseSingleFilter(js.GetIndex(i)); err != nil {
			return err
		} else {
			f.Filters = append(f.Filters, fb)
		}
	}
	return nil
}

func parseSingleFilter(f *simplejson.Json) (fb filterbody, err error) {
	//fmt.Println("parseSingleFilter", SString(f))
	if f == nil {
		err = errors.New("parseSingleFilter with nil raw filter content.")
		return
	}
	fs, err := f.Array()
	if err != nil {
		return
	}

	cs := make([]*simplejson.Json, len(fs))
	for i := 0; i < len(fs); i++ {
		cs[i] = f.GetIndex(i)
	}
	fb.Condtions, err = parseConditionList(cs)
	if err != nil {
		return
	}

	return
}

func parseConditionList(cs []*simplejson.Json) (cl conditionlist, err error) {
	//fmt.Println("parseConditonList", String(cs))
	switch len(cs) {
	case 0:
		err = errors.New("parseConditionList with nil raw filter content.")
		return
	case 1:
		if cs[0] == nil {
			err = fmt.Errorf("1 parseConditionList with invalid raw filter content %+#v.", cs)
			return
		}
		s, err := cs[0].StringArray()
		if err == nil { // [condition]
			if len(s) < 3 {
				err = fmt.Errorf("2 parseConditionList with invalid raw filter content %+#v.", String(cs))
				return cl, err
			}
			cl.Combination = VarLogicAnd
			cl.Conditions = []condition{condition{Key: s[0], Op: s[1], Expr: s[2:]}}
		} else { // ["or/and/not", [condtion], [condition], ...]
			la, err := cs[0].Array()
			if err != nil {
				return cl, err
			}
			if len(la) < 2 {
				err = fmt.Errorf("3 parseConditionList with invalid raw filter content %+#v.", String(cs))
				return cl, err
			}
			ls, err := cs[0].GetIndex(0).String()
			if err != nil {
				return cl, err
			}
			switch ls {
			case VarLogicAnd:
				fallthrough
			case VarLogicNot:
				fallthrough
			case VarLogicOr:
				cl.Combination = ls
			default:
				err = fmt.Errorf("4 parseConditionList with invalid raw filter content %+#v.", String(cs))
				return cl, err
			}

			for i := 1; i < len(la); i++ {
				s, err := cs[0].GetIndex(i).StringArray()
				if err != nil {
					return cl, err
				}
				if len(s) < 3 {
					err = fmt.Errorf("5 parseConditionList with invalid raw filter content %+#v.", String(cs))
					return cl, err
				}
				cl.Conditions = append(cl.Conditions, condition{Key: s[0], Op: s[1], Expr: s[2:]})
			}
		}
	default: // c = [condition], [condition], ...
		cl.Combination = VarLogicAnd
		for _, c := range cs {
			if c == nil {
				continue
			}
			s, err := c.StringArray()
			if err != nil {
				return cl, err
			}
			if len(s) < 3 {
				err = fmt.Errorf("6 parseConditionList with invalid raw filter content %+#v.", String(cs))
				return cl, err
			}
			cl.Conditions = append(cl.Conditions, condition{Key: s[0], Op: s[1], Expr: s[2:]})
		}
		return
	}

	return cl, nil
}

func parseActionList(a *simplejson.Json) (al actionlist, err error) {
	//fmt.Println("parseActionList", SString(a))
	if a == nil {
		err = fmt.Errorf("1 parseActionList with invalid raw filter content %+#v.", a)
		return
	}

	ls, err := a.StringArray()
	if err == nil { // [action]
		if len(ls) < 3 {
			err = fmt.Errorf("2 parseActionList with invalid raw filter content %+#v.", a)
			return
		}
		al.Sactions = []action{action{Key: ls[0], Op: ls[1], Expr: ls[2:]}}
	} else {
		// [[action],[action]]
		// [["80",[action],[action]],["20",[action],[action]]]
		// [[action],[action],[["80",[action],[action]],["20",[action],[action]]],[action]]
		lf, err := a.Array()
		if err != nil {
			return al, err
		}
		gcoa := combinedoactions{}
		for i := 0; i < len(lf); i++ {
			if ss, err := a.GetIndex(i).StringArray(); err == nil { // [action]
				if len(ss) < 3 {
					err = fmt.Errorf("3 parseActionList with invalid raw filter content %+#v.", a)
					return al, err
				}
				al.Sactions = append(al.Sactions, action{Key: ss[0], Op: ss[1], Expr: ss[2:]})
			} else {
				if _, err := a.GetIndex(i).GetIndex(0).String(); err == nil { // ["80",[action],[action]]
					oa, err := parseSingleOptionalAction(a.GetIndex(i))
					if err != nil {
						return al, err
					}
					gcoa.Actions = append(gcoa.Actions, oa)
				} else { // [["80",[action],[action]],["20",[action],[action]]]
					ff, err := a.GetIndex(i).Array()
					if err != nil {
						return al, err
					}
					coa := combinedoactions{}
					for j := 0; j < len(ff); j++ {
						oa, err := parseSingleOptionalAction(a.GetIndex(i).GetIndex(j))
						if err != nil {
							return actionlist{}, err
						}
						coa.Actions = append(coa.Actions, oa)
					}
					al.Oactions = append(al.Oactions, coa)
				}
			}
		}
		if len(gcoa.Actions) > 0 && len(al.Sactions) > 0 {
			// optional action暴露出来和普通action处于同一层级，是不允许的
			// [[action],["80",[action],[action]],["20",[action],[action]]] forbidden
			err = fmt.Errorf("4 parseActionList with invalid raw filter content %+#v.", a)
			return al, err
		}
		if len(gcoa.Actions) > 0 {
			al.Oactions = append(al.Oactions, gcoa)
		}
	}

	return al, nil
}

//TODO
func parseSimpleAction(a *simplejson.Json) (sa action, err error) {
	return
}

// ["80",[action],[action]]
func parseSingleOptionalAction(a *simplejson.Json) (oa optionalaction, err error) {
	if a == nil {
		err = fmt.Errorf("1 parseSingleOptionalAction with invalid raw filter content %+#v.", a)
		return
	}
	la, err := a.Array()
	if err != nil {
		return
	}
	if len(la) < 2 {
		err = fmt.Errorf("2 parseSingleOptionalAction with invalid raw filter content %+#v.", a)
		return
	}
	percent, err := a.GetIndex(0).String()
	if err != nil {
		return
	}
	oa.Percent, err = strconv.ParseUint(percent, 10, 64)
	if err != nil {
		return
	}
	for i := 1; i < len(la); i++ {
		sa, err := a.GetIndex(i).StringArray()
		if err != nil {
			return oa, err
		}
		oa.Actions = append(oa.Actions, action{Key: sa[0], Op: sa[1], Expr: sa[2:]})
	}

	return
}

func conditionOK(req request.Request, cl conditionlist) bool {
	result := false
	for _, c := range cl.Conditions {
		b := LOF(c.Op)(KF(c.Key)(req), c.Expr, req)
		switch cl.Combination {
		case VarLogicAnd:
			if !b {
				return false
			}
			result = true
		case VarLogicNot:
			if !b {
				return true
			}
		case VarLogicOr:
			if b {
				return true
			}
		default:
			panic(fmt.Sprintf("Unsupported Logic Op in %+v", cl))
		}
	}
	return result
}

func apply(req request.Request, al actionlist) {
	for _, a := range al.Sactions {
		AOF(a.Op)(a.Key, a.Expr, req)
	}

	for _, ca := range al.Oactions {
		total := 0
		for _, oa := range ca.Actions {
			total += int(oa.Percent)
		}
		r := rand.Intn(total)
		for _, oa := range ca.Actions {
			if r >= int(oa.Percent) {
				r -= int(oa.Percent)
				continue
			}
			for _, a := range oa.Actions {
				AOF(a.Op)(a.Key, a.Expr, req)
			}
			break
		}
	}
}

func (f *filterImpl) Marshal() string {
	if f == nil {
		return ""
	}

	if fb, err := json.Marshal(f.Filters); err != nil {
		return err.Error()
	} else {
		return string(fb)
	}
}

func (f *filterImpl) Accept(req request.Request) bool {
	if f == nil || req == nil {
		return false
	}
	if len(f.Filters) == 0 {
		return true
	}
	// 其他过滤条件，所有Filter之间的关系是 or，任意一个满足即可通过
	for _, fb := range f.Filters {
		if !conditionOK(req, fb.Condtions) {
			log.Infof("[adFilter][filter] not match %+#v", fb.Condtions)
			continue
		}
		return true
	}

	// 没有任何filter条件满足
	return false
}

// for debug log
func SString(cs *simplejson.Json) (s string) {
	if b, err := cs.Encode(); err == nil {
		s += string(b)
	}
	return
}

// for debug log
func String(cs []*simplejson.Json) (s string) {
	s = "["
	defer func() {
		s += "]"
	}()

	if len(cs) == 0 {
		return
	}
	for i, _ := range cs {
		if cs[i] == nil {
			continue
		}
		if i == 0 {
			s += "["
		} else {
			s += ",["
		}
		if b, err := cs[i].Encode(); err == nil {
			s += string(b)
		}
		s += "]"
	}
	return
}
