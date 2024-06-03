package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_IsAllowUrl(t *testing.T) {
	fmt.Println(IsUrl("http://192.158.1/1"))
	assert.True(t, IsUrl("http://google.com"))
	assert.True(t, IsUrl("http://w.com/cn"))
	assert.True(t, IsUrl("http://192.158.0.1:90"))
	assert.False(t, IsUrl("http://w"))
	assert.False(t, IsUrl("fsw"))
	assert.False(t, IsUrl("http://192.158.1/1"))

}
