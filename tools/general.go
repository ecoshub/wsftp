package tools

import "fmt"
import "github.com/ecoshub/jint"

var LOG_SCHEME *jint.Scheme = jint.MakeScheme("event", "content")

func StdoutHandle(event, content string, err error) {
	if err == nil {
		fmt.Println(string(LOG_SCHEME.MakeJson(event, content)))
		return
	}
	fmt.Println(string(LOG_SCHEME.MakeJson(event, content + " err: " + err.Error())))
	return
}
