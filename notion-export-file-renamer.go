package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// 파일 이름에서 사용한 정규식 패턴입니다.
var fileRegexp = regexp.MustCompile(` \w{32}\b`)

func processMarkdownFiles() {
	err := filepath.Walk(".",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Println("Error walking file path", err)
				return err
			}

			// .md 파일인지 확인합니다.
			if !info.IsDir() && strings.HasSuffix(info.Name(), ".md") {
				content, err := os.ReadFile(path)
				if err != nil {
					fmt.Println("Error reading file", err)
					return err
				}

				// []() 패턴을 찾아서 파일명과 동일한 조건으로 변경합니다.
				substitutedContent := regexp.MustCompile(`\[.*?\]\((.*?)\)`).ReplaceAllFunc(content,
					func(match []byte) []byte {
						matchStr := string(match)
						matches := regexp.MustCompile(`\[.*?\]\((.*?)\)`).FindStringSubmatch(matchStr)
						if len(matches) > 1 {
							fileName := filepath.Base(matches[1])
							fileNameWithoutExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))
							return []byte(fmt.Sprintf("[%s](%s)", fileNameWithoutExt, fileNameWithoutExt))
						}
						return match
					},
				)

				err = os.WriteFile(path, substitutedContent, 0644)
				if err != nil {
					fmt.Println("Error writing file", err)
					return err
				}

				// 파일 내의 링크 URL에 적용한 후에 파일명 변경을 수행합니다.
				newPath := fileRegexp.ReplaceAllString(path, "")
				if newPath != path {
					newDir := filepath.Dir(newPath)
					if _, err := os.Stat(newDir); os.IsNotExist(err) {
						err := os.MkdirAll(newDir, os.ModePerm)
						if err != nil {
							fmt.Println("Error creating directory:", err)
							return err
						}
					}

					err = os.Rename(path, newPath)
					if err != nil {
						fmt.Println("Error renaming file", err)
						return err
					}

					// 왜인지 모르겠는데 해당 처리를 해도 디렉토리가 남아 있음...(iCloud로 인한 로컬 문제일지도)
					oldDir := filepath.Dir(path)
					if _, err := os.Stat(oldDir); os.IsExist(err) {
						fmt.Println("Old dir : ", oldDir)
						err := os.RemoveAll(oldDir)
						if err != nil {
							fmt.Println("Error removing old directory:", err)
							return err
						}
					}
				}
			}

			return nil
		},
	)

	if err != nil {
		fmt.Println("Error walking file path", err)
		return
	}

	fmt.Println("Successfully processed markdown files")
}

func extractAndReplaceSrc() {
	rootDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return
	}

	rawFolder := filepath.Join(rootDir, "raw")
	files, err := os.ReadDir(rawFolder)
	if err != nil {
		fmt.Println("Error reading raw folder:", err)
		return
	}

	var zipFilePath string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".zip") {
			zipFilePath = filepath.Join(rawFolder, file.Name())
			break
		}
	}

	if zipFilePath == "" {
		fmt.Println("No zip files found in the 'raw' folder.")
		return
	}

	extractPath := filepath.Join(rootDir, "temp_extract")
	err = unzip(zipFilePath, extractPath)
	if err != nil {
		fmt.Println("Error extracting zip file:", err)
		return
	}

	srcPath := filepath.Join(rootDir, "src")
	if _, err := os.Stat(srcPath); !os.IsNotExist(err) {
		err = os.RemoveAll(srcPath)
		if err != nil {
			fmt.Println("Error removing existing src folder:", err)
			return
		}
	}

	err = os.Rename(extractPath, srcPath)
	if err != nil {
		fmt.Println("Error moving extracted folder to src:", err)
		return
	}

	err = os.RemoveAll(extractPath)
	if err != nil {
		fmt.Println("Error removing temporary extraction folder:", err)
		return
	}

	fmt.Println("Successfully processed file", srcPath)
}

func unzip(src string, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}

	defer func(r *zip.ReadCloser) {
		err = r.Close()
		if err != nil {
			fmt.Println("Error closing zip reader:", err)
		}
	}(r)

	for _, f := range r.File {
		fPath := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			err = os.MkdirAll(fPath, os.ModePerm)
			if err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fPath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			_ = outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)

		_ = outFile.Close()
		_ = rc.Close()

		if err != nil {
			return err
		}
	}

	fmt.Println("Successfully unzipped file", src)
	return nil
}

func cleanUnusedDir(dirPath string) error {
	// 디렉토리 내부에 파일이 있는지 확인
	files, err := os.ReadDir(dirPath)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return err
	}

	// 디렉토리가 비어 있으면 삭제
	if len(files) == 0 {
		err := os.Remove(dirPath) // 비어 있는 디렉토리 삭제
		if err != nil {
			fmt.Println("Error removing empty directory:", err)
			return err
		} else {
			fmt.Println("Successfully removed empty directory:", dirPath)
		}
	} else {
		fmt.Println("Directory is not empty:", dirPath)
	}

	return nil
}

func main() {
	extractAndReplaceSrc()
	processMarkdownFiles()
}
