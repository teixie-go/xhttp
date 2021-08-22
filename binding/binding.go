package binding

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
)

var (
	_ Binding = (*xmlBinding)(nil)
	_ Binding = (*jsonBinding)(nil)

	JSON = jsonBinding{}
	XML  = xmlBinding{}
)

type Binding interface {
	Name() string
	Bind([]byte, interface{}) error
}

//------------------------------------------------------------------------------

type xmlBinding struct{}

func (xmlBinding) Name() string {
	return "xml"
}

func (xmlBinding) Bind(body []byte, obj interface{}) error {
	decoder := xml.NewDecoder(bytes.NewReader(body))
	return decoder.Decode(obj)
}

//------------------------------------------------------------------------------

type jsonBinding struct{}

func (jsonBinding) Name() string {
	return "json"
}

func (jsonBinding) Bind(body []byte, obj interface{}) error {
	return json.Unmarshal(body, obj)
}
