package libfs

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/jiuzhou-zhao/go-fundamental/pathutils"
	"github.com/satori/go.uuid"
)

/*
V1
5a8dd3ad0756a93ded72b823b19dd877-6.test

V2
v2-6-5a8dd3ad0756a93ded72b823b19dd877-ab.test
*/

const (
	rPathV1 = "V1"
	rPathV2 = "V2"
)

// Item class
type Item struct {
	rootPath string
	tempPath string
	fileInfo *fileInfo
}

// NewSFSItemByInfo function
func NewSFSItemByInfo(fileMd5 string, fileSize uint64, fileName string, rootPath string, tempPath string) (*Item, error) {
	fileInfo, err := newFileInfoFromRawInfo(fileMd5, fileSize, fileName)
	if err != nil {
		return nil, err
	}
	return &Item{
		rootPath: rootPath,
		tempPath: tempPath,
		fileInfo: fileInfo,
	}, nil
}

// NewSFSItem function
func NewSFSItem(ext string, rootPath string, tempPath string) (*Item, error) {
	return NewSFSItemWithVer(FileIDV2, ext, rootPath, tempPath)
}

// NewSFSItemWithVer function
func NewSFSItemWithVer(ver uint, ext, rootPath, tempPath string) (*Item, error) {
	fileInfo, err := newFileInfoWithVer(ext, ver)
	if err != nil {
		return nil, err
	}
	return &Item{
		rootPath: rootPath,
		tempPath: tempPath,
		fileInfo: fileInfo,
	}, nil
}

// NewSFSItemFromFileID function
func NewSFSItemFromFileID(fileID string, rootPath string, tempPath string) (*Item, error) {
	fileInfo, err := newFileInfoFromFileID(fileID)
	if err != nil {
		return nil, err
	}
	return &Item{
		rootPath: rootPath,
		tempPath: tempPath,
		fileInfo: fileInfo,
	}, nil
}

func (item *Item) versionRootPath() string {
	return item.versionRootPathForVer(item.fileInfo.fileIDVersion)
}

func (item *Item) versionRootPathForVer(verID uint) string {
	switch verID {
	case FileIDV1:
		return filepath.Join(item.rootPath, rPathV1)
	case FileIDV2:
		return filepath.Join(item.rootPath, rPathV2)
	}
	log.Fatalf("invalid fileIDVersion: %+v", item)
	return ""
}

// ExistsInStorage method
func (item *Item) ExistsInStorage() (dataExists, fileExists bool, err error) {
	rDataPath, err := item.fileInfo.getDataFile()
	if err != nil {
		return false, false, err
	}
	dataExists, err = pathutils.IsFileExists(filepath.Join(item.versionRootPath(), rDataPath))
	if err != nil {
		return
	}
	rFilePath, err := item.fileInfo.getNameFile()
	if err != nil {
		return false, false, err
	}
	fileExists, err = pathutils.IsFileExists(filepath.Join(item.versionRootPath(), rFilePath))
	return
}

// ExistsDataInAllStorage method
func (item *Item) ExistsDataInAllStorage() (dataExists bool, err error) {
	for ver := FileIDV1; ver <= FileIDVMAX; ver++ {
		rDataPath, err := item.fileInfo.getDataFileForVer(uint(ver))
		if err != nil {
			continue
		}
		dataExists, err = pathutils.IsFileExists(filepath.Join(item.versionRootPathForVer(uint(ver)), rDataPath))
		if err != nil {
			continue
		}
		if dataExists {
			return true, nil
		}
	}
	return false, nil

}

// GetDataFile method
func (item *Item) GetDataFile() (string, error) {
	rDataPath, err := item.fileInfo.getDataFile()
	if err != nil {
		return "", err
	}
	return filepath.Join(item.versionRootPath(), rDataPath), nil
}

// GetNameFile method
func (item *Item) GetNameFile() (string, error) {
	rDataPath, err := item.fileInfo.getNameFile()
	if err != nil {
		return "", err
	}
	return filepath.Join(item.versionRootPath(), rDataPath), nil
}

// GetFileID method
func (item *Item) GetFileID() (string, error) {
	return item.fileInfo.getFileID()
}

// WriteFileRecord method
func (item *Item) WriteFileRecord() error {
	rFilePath, err := item.fileInfo.getNameFile()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(item.versionRootPath(), rFilePath), []byte(""), 0666)
}

// WriteFile method
func (item *Item) WriteFile(reader io.Reader) error {
	u1 := uuid.NewV4()

	md5writer := md5.New()
	r := io.TeeReader(reader, md5writer)

	fileName := filepath.Join(item.tempPath, u1.String())
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()
	fileWriter := bufio.NewWriter(file)
	defer func() {
		_ = fileWriter.Flush()
	}()

	transBytes, err := io.Copy(fileWriter, r)
	if err != nil {
		return err
	}

	bs := md5writer.Sum(nil)
	dst := make([]byte, hex.EncodedLen(len(bs)))
	hex.Encode(dst, bs)

	err = item.fileInfo.initFileInfos(string(dst), uint64(transBytes))
	if err != nil {
		return err
	}

	rDataFile, err := item.fileInfo.getDataFile()
	if err != nil {
		return err
	}
	dataFile := filepath.Join(item.versionRootPath(), rDataFile)
	err = pathutils.MakesureDirOfFileExists(dataFile)
	if err != nil {
		return err
	}

	_ = fileWriter.Flush()
	_ = file.Close()

	return os.Rename(fileName, dataFile)
}
