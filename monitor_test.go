package monitor

import (
	"fmt"
	"testing"
)

func TestIncCounter(t *testing.T) {
	Start("Xxx", 8080)
	err := IncCounter("test", map[string]string{"key1": "value1", "key2": "value2"})
	if err != nil {
		fmt.Println(err)
	}
}
