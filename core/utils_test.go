package core

import (
	"testing"
)

func TestGetExtType(t *testing.T) {
	url := "https://domain.com/data/avatars/m/123/12312312.jpg?1562846649"
	t.Log(GetExtType(url))
}

func TestFixUrl(t *testing.T) {
	//
}
