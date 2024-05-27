package logjson

import "reflect"

type LogRule func(conf *logRuleConf)

type logRuleConf struct {
	md5  bool
	omit bool
}

func newLogRuleConfFromStr(ruleStr string) *logRuleConf {
	conf := &logRuleConf{}
	rule := conf.parseLogRule(ruleStr)
	if rule == nil {
		return nil
	}
	conf.init(rule)
	return conf
}

func newLogRuleConf(rule LogRule) *logRuleConf {
	conf := &logRuleConf{}
	conf.init(rule)
	return conf
}

func (conf *logRuleConf) parseLogRule(ruleStr string) LogRule {
	switch ruleStr {
	case "omit":
		return LogRuleOmit()
	case "md5":
		return LogRuleMd5()
	default:
		return nil
	}
}

func (conf *logRuleConf) init(rule LogRule) {
	rule(conf)
}

func (conf *logRuleConf) Omit() bool {
	return conf.omit
}

func (conf *logRuleConf) GetHandlerItem(field reflect.StructField) *handlerItem {
	if conf.omit {
		return nil
	}
	if field.Type.Kind() != reflect.String {
		return nil
	}
	item := &handlerItem{}
	if conf.md5 {
		item.marshal = md5Marshal
		return item
	}
	return nil
}

func LogRuleMd5() LogRule {
	return func(conf *logRuleConf) {
		conf.md5 = true
	}
}

func LogRuleOmit() LogRule {
	return func(conf *logRuleConf) {
		conf.omit = true
	}
}
