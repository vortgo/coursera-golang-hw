package hw1

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
)

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

func dirTree(out io.Writer, path string, printFiles bool) error {
	printTree(out, path, 0, "", printFiles)
	return nil
}

func printTree(out io.Writer, path string, depth int, prefix string, printFiles bool) {
	var (
		entity string
	)
	files, _ := ioutil.ReadDir(path)

	if !printFiles {
		files = getOnlyDirs(files)
	}

	for i, f := range files {
		isLast := checkIsLast(len(files), i+1)
		entity = getPrefixForName(isLast) + getEntityName(f) + "\n"
		out.Write([]byte(prefix + entity))
		if f.IsDir() {
			prefixForChild := getPrefixForNextDepth(isLast, prefix)
			printTree(out, path+string(os.PathSeparator)+f.Name(), depth+1, prefixForChild, printFiles)
		}
	}
}

func getPrefixForNextDepth(isLast bool, prevPrefix string) string {
	var prefixForChild string
	if isLast {
		prefixForChild = prevPrefix + "\t"
	} else {
		prefixForChild = prevPrefix + "│\t"
	}

	return prefixForChild
}

func checkIsLast(len, current int) bool {
	if len == current {
		return true
	}

	return false
}

func getPrefixForName(isLast bool) string {
	var prefix string

	if isLast {
		prefix = `└───`
	} else {
		prefix = `├───`
	}

	return prefix
}

func getEntityName(file os.FileInfo) string {
	var sizeString string

	if file.IsDir() {
		return file.Name()
	} else {
		size := file.Size()
		if size == 0 {
			sizeString = `empty`
		} else {
			sizeString = strconv.FormatInt(size, 10) + `b`
		}
		return fmt.Sprintf("%s (%s)", file.Name(), sizeString)
	}
}

func getOnlyDirs(files []os.FileInfo) []os.FileInfo {
	var dirs []os.FileInfo
	for _, val := range files {
		if val.IsDir() {
			dirs = append(dirs, val)
		}
	}
	return dirs
}
