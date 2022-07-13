package libfs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/sgostarter/libeasygo/pathutils"
	"github.com/stretchr/testify/assert"
)

const (
	testRootPath = "./teststg"
	rootPath     = "./teststg/root"
	tempPath     = "./teststg/testtemp"
)

func TestMain(m *testing.M) {
	_ = pathutils.RemoveAll(testRootPath)
	_ = pathutils.MustDirExists(rootPath)
	_ = pathutils.MustDirExists(tempPath)
	ret := m.Run()
	_ = pathutils.RemoveAll(testRootPath)
	os.Exit(ret)
}

func TestV1l(t *testing.T) {
	tempFile := filepath.Join(tempPath, "test.txt")
	err := ioutil.WriteFile(tempFile, []byte("hello!!"), 0644)
	assert.Nil(t, err)

	sfsItem, err := NewSFSItemWithVer(FileIDV1, "ab.test", rootPath, tempPath)
	assert.Nil(t, err)

	fileObj, err := os.OpenFile(tempFile, os.O_RDONLY, 0644)
	assert.Nil(t, err)
	defer func() {
		_ = fileObj.Close()
	}()

	err = sfsItem.WriteFile(fileObj)
	assert.Nil(t, err)

	err = sfsItem.WriteFileRecord()
	assert.Nil(t, err)

	fileID, err := sfsItem.GetFileID()
	assert.Nil(t, err)

	fmt.Printf("FileID: %v\n", fileID)

	sfsItem2, err := NewSFSItemFromFileID(fileID, rootPath, tempPath)
	assert.Nil(t, err)

	de, fe, err := sfsItem2.ExistsInStorage()
	assert.Nil(t, err)
	assert.True(t, de)
	assert.True(t, fe)

	df, err := sfsItem2.GetDataFile()
	assert.Nil(t, err)
	nf, err := sfsItem2.GetNameFile()
	assert.Nil(t, err)
	fmt.Println(df)
	fmt.Println(nf)
}

func TestV2l(t *testing.T) {
	tempFile := filepath.Join(tempPath, "test.txt")
	err := ioutil.WriteFile(tempFile, []byte("hello!"), 0644)
	assert.Nil(t, err)

	sfsItem, err := NewSFSItemWithVer(FileIDV2, "ab.test", rootPath, tempPath)
	assert.Nil(t, err)

	fileObj, err := os.OpenFile(tempFile, os.O_RDONLY, 0644)
	assert.Nil(t, err)
	defer fileObj.Close()

	err = sfsItem.WriteFile(fileObj)
	assert.Nil(t, err)

	err = sfsItem.WriteFileRecord()
	assert.Nil(t, err)

	fileID, err := sfsItem.GetFileID()
	assert.Nil(t, err)
	fmt.Printf("FileID: %v\n", fileID)

	sfsItem2, err := NewSFSItemFromFileID(fileID, rootPath, tempPath)
	assert.Nil(t, err)

	de, fe, err := sfsItem2.ExistsInStorage()
	assert.Nil(t, err)
	assert.True(t, fe)
	assert.True(t, de)

	df, err := sfsItem2.GetDataFile()
	assert.Nil(t, err)
	nf, err := sfsItem2.GetNameFile()
	assert.Nil(t, err)
	fmt.Println(df)
	fmt.Println(nf)

	exists, err := IsSizeExistsInStorage(5, rootPath)
	assert.Nil(t, err)
	assert.False(t, exists)

	exists, err = IsSizeExistsInStorage(6, rootPath)
	assert.Nil(t, err)
	assert.True(t, exists)
}
