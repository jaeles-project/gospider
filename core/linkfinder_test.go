package core

import "testing"

func Test_ParseJSSource(t *testing.T) {
	source := `
"%2faaa.com"
"https:\u002F\u002Fs.yimg.com\u002Fnq\u002Fstore-badges\u002F4\u002Fstore-badges\u002F"`
	t.Log(LinkFinder(source))
}
