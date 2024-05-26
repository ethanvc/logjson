package logjson

type LogRule func(conf *logRuleConf)

type logRuleConf struct {
	md5 bool
}

func newLogRuleConf(rule LogRule) *logRuleConf {
	conf := &logRuleConf{}
	conf.init(rule)
	return conf
}

func (conf *logRuleConf) init(rule LogRule) {
	rule(conf)
}

func LogRuleMd5() LogRule {
	return func(conf *logRuleConf) {
		conf.md5 = true
	}
}
