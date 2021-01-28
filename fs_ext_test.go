package libfs

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFileList(t *testing.T) {
	TestV1l(t)
	TestV2l(t)
	lastFileID := ""
	for {
		err, files := GetFileList(lastFileID, rootPath, true, 2)
		if !assert.Nil(t, err) {
			t.Fatal()
		}
		if files == nil {
			break
		}
		if len(files) <= 0 {
			break
		}
		for _, file := range files {
			fmt.Println(file)
		}
		lastFileID = files[len(files)-1]
		if len(files) != 2 {
			break
		}
	}

	err, files := GetFileList(lastFileID, rootPath, false, 2)
	assert.Nil(t, err)
	for _, file := range files {
		t.Log(file)
	}
	err, files = GetFileList("81d95db337a18c65384d35ba7ea2efda-7.none", rootPath, false, 2)
	assert.Nil(t, err)
	assert.True(t, len(files) == 1)
	for _, file := range files {
		t.Log(file)
	}
}
