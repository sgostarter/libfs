package libfs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jiuzhou-zhao/go-fundamental/pathutils"
)

const (
	Separators = string(filepath.Separator)
)

// IsSizeExistsInStorage function
func IsSizeExistsInStorage(fileSize uint64, rootPath string) (bool, error) {
	return pathutils.IsDirExists(filepath.Join(rootPath, rPathV2, fmt.Sprintf("%v", fileSize)))
}

// IsSizeMD5ExistsInStorage function
func IsSizeMD5ExistsInStorage(fileSize uint64, md5 string, rootPath string) (bool, error) {
	item, err := NewSFSItemByInfo(md5, fileSize, "", rootPath, "")
	if err != nil {
		return false, err
	}
	return item.ExistsDataInAllStorage()
}

func parseFileIDFromFilePath(filePath string, version uint) (error, string) {
	if version == FileIDV1 {
		// 5a8dd3ad0756a93ded72b823b19dd877-6.test
		filepath.Join()
		key := Separators + rPathV1 + Separators
		idx := strings.Index(filePath, key)

		if idx == -1 {
			return fmt.Errorf("unsupported %v", filePath), ""
		}
		pathIDs := strings.Split(filePath[idx+len(key):], Separators)
		if len(pathIDs) != 9 {
			return fmt.Errorf("unsupported %v", filePath), ""
		}
		fileMD5 := strings.Join(pathIDs[0:8], "")
		fileSize, err := strconv.ParseUint(pathIDs[8], 10, 64)
		if err != nil {
			return err, ""
		}
		fileInfo, err := newFileInfoFromRawInfoWithVersion(version, fileMD5, fileSize, BuildInDataName)
		if err != nil {
			return err, ""
		}
		fileID, err := fileInfo.getFileID()
		return err, fileID
	}

	if version == FileIDV2 {
		// */V2/size/md1/.../__date__
		// v2-6-5a8dd3ad0756a93ded72b823b19dd877-ab.test
		key := Separators + rPathV2 + Separators
		idx := strings.Index(filePath, key)
		if idx == -1 {
			return fmt.Errorf("unsupported %v", filePath), ""
		}
		pathIDs := strings.Split(filePath[idx+len(key):], Separators)
		if len(pathIDs) != 10 {
			return fmt.Errorf("unsupported %v", filePath), ""
		}
		fileSize, err := strconv.ParseUint(pathIDs[0], 10, 64)
		if err != nil {
			return err, ""
		}
		fileMD5 := strings.Join(pathIDs[1:9], "")
		fileInfo, err := newFileInfoFromRawInfo(fileMD5, fileSize, BuildInDataName)
		if err != nil {
			return err, ""
		}
		fileID, err := fileInfo.getFileID()
		return err, fileID
	}

	return fmt.Errorf("unknown version %v on %v", version, filePath), ""
}

func pathInfosFromFileID(fileID string) (err error, ver uint, ids []string) {
	var fileInfo *fileInfo
	fileInfo, err = newFileInfoFromFileID(fileID)
	if err != nil {
		return
	}
	var dataDir string
	dataDir, err = fileInfo.getDataDir()
	if err != nil {
		return
	}
	ver = fileInfo.fileIDVersion
	ids = strings.Split(dataDir, "/")
	return
}

func utilWalk(b, e int, forward bool, cb func(idx int) bool) {
	if forward {
		for idx := b; idx <= e; idx++ {
			r := cb(idx)
			if !r {
				break
			}
		}
	} else {
		for idx := e; idx >= b; idx-- {
			r := cb(idx)
			if !r {
				break
			}
		}
	}
}

func getFileList(parentPath string, version uint, findNext bool, ids []string, count int) (err error, files []string) {
	if count == 0 {
		return
	}
	var exists bool
	exists, err = pathutils.IsDirExists(parentPath)
	if err != nil {
		return
	}
	if !exists {
		err = fmt.Errorf("%v not exists", parentPath)
		return
	}
	var fis []os.FileInfo
	fis, err = ioutil.ReadDir(parentPath)
	if err != nil {
		return
	}

	leftCount := count
	var curID string
	var nextIDs []string
	if len(ids) > 1 {
		curID = ids[0]
		nextIDs = ids[1:]
	}

	var newFiles []string
	utilWalk(0, len(fis)-1, findNext, func(idx int) bool {
		curPath := filepath.Join(parentPath, fis[idx].Name())
		if !fis[idx].IsDir() {
			if len(ids) > 0 {
				// 哨兵作用在这里
				return false
			}
			if version == FileIDV1 {
				var fileSize int64
				fileSize, err = strconv.ParseInt(fis[idx].Name(), 10, 64)
				if err != nil || fileSize == 0 {
					return true
				}
				fileID := ""
				err, fileID = parseFileIDFromFilePath(curPath, version)
				if err != nil {
					return false
				}
				files = append(files, fileID)
				return false
			} else if version == FileIDV2 {
				if fis[idx].Name() == BuildInDataName {
					fileID := ""
					err, fileID = parseFileIDFromFilePath(curPath, version)
					if err != nil {
						return false
					}
					files = append(files, fileID)
					return false
				}
			}
			return true
		}

		if curID != "" && curID != fis[idx].Name() {
			return true
		}
		_, newFiles = getFileList(curPath, version, findNext, nextIDs, leftCount)
		if curID != "" {
			curID = ""
			nextIDs = nil
		}
		if len(newFiles) > 0 {
			files = append(files, newFiles...)
			leftCount -= len(newFiles)
		}
		if leftCount == 0 {
			return false
		}
		return true
	})

	return
}

// GetFileList function
func GetFileList(lastFileID, rootPath string, findNext bool, count int) (err error, files []string) {
	var exists bool
	exists, err = pathutils.IsDirExists(rootPath)
	if err != nil {
		return
	}
	if !exists {
		err = fmt.Errorf("%v not exists", rootPath)
		return
	}

	var ids []string
	var ver uint
	ver = FileIDV1
	if lastFileID != "" {
		err, ver, ids = pathInfosFromFileID(lastFileID)
		if err == nil {
			// 加一个哨兵
			ids = append(ids, "<guard>")
		}
	}

	var filesBatch []string
	if ver == FileIDV1 {
		err, filesBatch = getFileList(filepath.Join(rootPath, rPathV1), ver, findNext, ids, count)
		files = append(files, filesBatch...)
		count -= len(filesBatch)
		ids = make([]string, 0)
		ver = FileIDV2
	}
	if ver == FileIDV2 {
		err, filesBatch = getFileList(filepath.Join(rootPath, rPathV2), ver, findNext, ids, count)
		files = append(files, filesBatch...)
	}

	if len(files) > 0 {
		err = nil
	}
	return
}
