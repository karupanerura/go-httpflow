package httpflow

import "fmt"

type UnexpectedContentTypeError struct {
	ContentType string
	Body        []byte
}

func (e *UnexpectedContentTypeError) Error() string {
	return fmt.Sprintf("Unexpected Content-Type: %s", e.ContentType)
}

type UnexpectedStatusCodeError struct {
	StatusCode
	Body []byte
}

func (e *UnexpectedStatusCodeError) Error() (msg string) {
	msg = fmt.Sprintf("Unexpected StatusCode %d", e.StatusCode)
	if e.Body != nil {
		msg += fmt.Sprintf(", Body = %s", truncateString(string(e.Body), 32))
	}
	return
}

func truncateString(str string, num int) string {
	bnoden := str
	if len(str) > num {
		if num > 3 {
			num -= 3
		}
		bnoden = str[0:num] + "..."
	}
	return bnoden
}
