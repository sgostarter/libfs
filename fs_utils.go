package sfs

import (
	"errors"
	"fmt"
	"path"
	"strconv"
	"strings"
)

var errorInvalidFileID = errors.New("invalid file id")
var errorUnknownFileID = errors.New("unknown file id")
var errorHasInitialized = errors.New("has intialized")
var errorNeedInitialized = errors.New("need intialized")

const (
	// FileIDV1 version of file id
	FileIDV1 = 1
	// FileIDV2 version of file id
	FileIDV2 = 2
	// FileIDVMAX .
	FileIDVMAX = 2

	// BuiltInFileName for v1
	BuiltInFileName = ".none"
	// BuildInDataName for v2 or above
	BuildInDataName = "_data_"
)

type fileInfo struct {
	fileIDVersion uint
	fileMD5       string
	fileSize      uint64
	fileName      string

	rDataDir string
}

func (fi *fileInfo) initFileInfos(fileMD5 string, fileSize uint64) error {
	if fi.fileMD5 != "" {
		return errorHasInitialized
	}
	fi.fileMD5 = fileMD5
	fi.fileSize = fileSize
	return nil
}

func (fi *fileInfo) getDataDir() (string, error) {
	if fi.fileMD5 == "" {
		return "", errorNeedInitialized
	}
	if fi.rDataDir != "" {
		return fi.rDataDir, nil
	}
	rDataFullPath, err := fi.getDataDirForVer(fi.fileIDVersion)
	if err != nil {
		return "", err
	}
	fi.rDataDir = rDataFullPath
	return fi.rDataDir, nil
}

func (fi *fileInfo) getDataDirForVer(verID uint) (string, error) {
	if fi.fileMD5 == "" {
		return "", errorNeedInitialized
	}

	if verID == FileIDV1 {
		return path.Join(fi.fileMD5[0:4], fi.fileMD5[4:8], fi.fileMD5[8:12], fi.fileMD5[12:16],
			fi.fileMD5[16:20], fi.fileMD5[20:24], fi.fileMD5[24:28], fi.fileMD5[28:32]), nil
	} else if verID == FileIDV2 {
		return path.Join(fmt.Sprintf("%v", fi.fileSize), fi.fileMD5[0:4], fi.fileMD5[4:8], fi.fileMD5[8:12], fi.fileMD5[12:16],
			fi.fileMD5[16:20], fi.fileMD5[20:24], fi.fileMD5[24:28], fi.fileMD5[28:32]), nil
	}

	return "", errorUnknownFileID
}

func (fi *fileInfo) getDataFile() (string, error) {
	return fi.getDataFileForVer(fi.fileIDVersion)
}

func (fi *fileInfo) getDataFileForVer(verID uint) (string, error) {
	if fi.fileMD5 == "" {
		return "", errorNeedInitialized
	}

	rDataDir, err := fi.getDataDirForVer(verID)
	if err != nil {
		return "", err
	}
	if verID == FileIDV1 {
		return path.Join(rDataDir, fmt.Sprintf("%v", fi.fileSize)), nil
	} else if verID == FileIDV2 {
		return path.Join(rDataDir, BuildInDataName), nil
	}

	return "", errorInvalidFileID
}

func (fi *fileInfo) getFileID() (string, error) {
	return fi.getFileIDByVer(fi.fileIDVersion)
}

func (fi *fileInfo) getFileIDByVer(verID uint) (string, error) {
	if fi.fileMD5 == "" {
		return "", errorNeedInitialized
	}

	if verID == FileIDV1 {
		fullDirName := fmt.Sprintf("%v-%v", fi.fileMD5, fi.fileSize)
		if fi.fileName != "" {
			dotIndex := strings.LastIndex(fi.fileName, ".")
			if dotIndex != -1 {
				fullDirName += fi.fileName[dotIndex:]
			} else {
				fullDirName += BuiltInFileName
			}
		}
		return fullDirName, nil
	} else if verID == FileIDV2 {
		return fmt.Sprintf("v2-%v-%v-%v", fi.fileSize, fi.fileMD5, fi.fileName), nil
	}

	return "", errorInvalidFileID
}

func (fi *fileInfo) getNameFile() (string, error) {
	return fi.getNameFileForVer(fi.fileIDVersion)
}

func (fi *fileInfo) getNameFileForVer(verID uint) (string, error) {
	if fi.fileMD5 == "" {
		return "", errorNeedInitialized
	}

	rDataDir, err := fi.getDataDirForVer(verID)
	if err != nil {
		return "", err
	}
	if verID == FileIDV1 {
		var fileField string
		dotIndex := strings.LastIndex(fi.fileName, ".")
		if dotIndex != -1 {
			fileField = fi.fileName[dotIndex:]
		} else {
			fileField = BuiltInFileName
		}
		return path.Join(rDataDir, fileField), nil
	} else if verID == FileIDV2 {
		return path.Join(rDataDir, fi.fileName), nil
	}

	return "", errorInvalidFileID
}

func newFileInfoWithVer(fileName string, ver uint) (*fileInfo, error) {
	return &fileInfo{
		fileIDVersion: ver,
		fileName:      fileName,
	}, nil
}

func newFileInfoFromRawInfo(fileMd5 string, fileSize uint64, fileName string) (*fileInfo, error) {
	return newFileInfoFromRawInfoWithVersion(FileIDVMAX, fileMd5, fileSize, fileName)
}

func newFileInfoFromRawInfoWithVersion(ver uint, fileMd5 string, fileSize uint64, fileName string) (*fileInfo, error) {
	fileMd5 = strings.Trim(fileMd5, " \r\n\t")
	if len(fileMd5) != 32 {
		return nil, errorInvalidFileID
	}
	return &fileInfo{
		fileIDVersion: ver,
		fileMD5:       fileMd5,
		fileSize:      fileSize,
		fileName:      fileName,
	}, nil
}

func newFileInfoFromFileID(fileID string) (*fileInfo, error) {
	fileID = strings.Trim(fileID, " \t\r\n")
	var finfo fileInfo
	if strings.Index(fileID, "v") == 0 {
		if strings.Index(fileID, "v2-") == 0 {
			finfo.fileIDVersion = FileIDV2
		} else {
			return nil, errorInvalidFileID
		}
	} else {
		if len(fileID) <= 34 || fileID[32] != '-' {
			return nil, errorInvalidFileID
		}
		finfo.fileIDVersion = FileIDV1
	}

	if finfo.fileIDVersion == FileIDV1 {
		finfo.fileMD5 = fileID[0:32]
		if fileID[32:33] != "-" {
			return nil, errorInvalidFileID
		}
		dotIdx := strings.Index(fileID, ".")

		var fileSizeStr string
		if dotIdx == -1 {
			fileSizeStr = fileID[33:]
			finfo.fileName = BuiltInFileName
		} else {
			fileSizeStr = fileID[33:dotIdx]
			finfo.fileName = fileID[dotIdx:]
		}

		fileSize, err := strconv.ParseUint(fileSizeStr, 10, 64)
		if err != nil {
			return nil, errorInvalidFileID
		}
		finfo.fileSize = fileSize
	} else if finfo.fileIDVersion == FileIDV2 {
		parts := strings.SplitN(fileID, "-", 4)
		if len(parts) != 4 {
			return nil, errorInvalidFileID
		}
		fsize, err := strconv.ParseUint(parts[1], 10, 64)
		if err != nil {
			return nil, err
		}
		finfo.fileSize = fsize
		finfo.fileMD5 = parts[2]
		finfo.fileName = parts[3]
	} else {
		return nil, errorUnknownFileID
	}

	if finfo.fileIDVersion != FileIDV1 {
		if finfo.fileName == BuildInDataName {
			finfo.fileName = "rel-" + BuildInDataName
		}
	}

	return &finfo, nil
}
