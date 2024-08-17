package urllib

import (
	"fmt"
	"net/url"
	"strings"
)

type BaseUrl struct {
	Base   string
	Params map[string]string
}

func (b *BaseUrl) AddParam(key string, value interface{}) *BaseUrl {
	switch v := value.(type) {
	case bool:
		if v {
			b.Params[key] = "true"
		} else {
			b.Params[key] = "false"
		}
	case string:
		b.Params[key] = v
	case int, int8, int16, int32, int64:
		b.Params[key] = fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		b.Params[key] = fmt.Sprintf("%d", v)
	case float32, float64:
		b.Params[key] = fmt.Sprintf("%f", v)
	default:
		b.Params[key] = fmt.Sprintf("%v", v)
	}
	return b
}

func (b *BaseUrl) AddParams(params map[string]interface{}) *BaseUrl {
	for key, value := range params {
		b.AddParam(key, value)
	}
	return b
}

func (b *BaseUrl) String() string {
	var params []string
	for key, value := range b.Params {
		encodedKey := url.QueryEscape(key)
		encodedValue := url.QueryEscape(value)
		params = append(params, fmt.Sprintf("%s=%s", encodedKey, encodedValue))
	}
	if len(params) == 0 {
		return b.Base
	}
	return fmt.Sprintf("%s?%s", b.Base, strings.Join(params, "&"))
}

func Url(base string, params ...map[string]interface{}) *BaseUrl {
	var paramMap map[string]interface{}
	if len(params) > 0 {
		paramMap = params[0]
	} else {
		paramMap = make(map[string]interface{})
	}

	// Replace placeholders in the base URL
	for key, value := range paramMap {
		placeholder := fmt.Sprintf(":%s", key)
		base = strings.Replace(base, placeholder, fmt.Sprintf("%v", value), -1)
	}

	b := &BaseUrl{
		Base:   base,
		Params: make(map[string]string),
	}
	return b
}
