package utils

import (
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// unifyPath will convert "\" to "/" in path if the os is windows
func unifyPath(path string) string {
	if IsWindows() {
		path = strings.ReplaceAll(path, "\\", "/")
	}
	return path
}

// BaseName get base name for path. ex: "github.com/p.s.m" => "p.s.m"
func BaseName(include, subFixToTrim string) string {
	include = unifyPath(include)
	subFixToTrim = unifyPath(subFixToTrim)
	last := include
	if id := strings.LastIndex(last, "/"); id >= 0 && id < len(last)-1 {
		last = last[id+1:]
	}
	if !strings.HasSuffix(last, subFixToTrim) {
		return last
	}
	return last[:len(last)-len(subFixToTrim)]
}

func BaseNameAndTrim(include string) string {
	include = unifyPath(include)
	last := include
	if id := strings.LastIndex(last, "/"); id >= 0 && id < len(last)-1 {
		last = last[id+1:]
	}

	if id := strings.LastIndex(last, "."); id != -1 {
		last = last[:id]
	}
	return last
}

func SplitPackageName(pkg, subFixToTrim string) string {
	pkg = unifyPath(pkg)
	subFixToTrim = unifyPath(subFixToTrim)
	last := SplitPackage(pkg, subFixToTrim)
	if id := strings.LastIndex(last, "/"); id >= 0 && id < len(last)-1 {
		last = last[id+1:]
	}
	return last
}

func SplitPackage(pkg, subFixToTrim string) string {
	pkg = unifyPath(pkg)
	subFixToTrim = unifyPath(subFixToTrim)
	last := strings.TrimSuffix(pkg, subFixToTrim)
	if id := strings.LastIndex(last, "/"); id >= 0 && id < len(last)-1 {
		last = last[id+1:]
	}
	return strings.ReplaceAll(last, ".", "/")
}

func PathToImport(path, subFix string) string {
	path = strings.TrimSuffix(path, subFix)
	// path = RelativePath(path)
	return strings.ReplaceAll(path, string(filepath.Separator), "/")
}

func ImportToPath(path, subFix string) string {
	// path = RelativePath(path)
	return strings.ReplaceAll(path, "/", string(filepath.Separator)) + subFix
}

func ImportToPathAndConcat(path, subFix string) string {
	path = strings.TrimSuffix(path, subFix)
	path = strings.ReplaceAll(path, "/", string(filepath.Separator))
	if i := strings.LastIndex(path, string(filepath.Separator)); i >= 0 && i < len(path)-1 && strings.Contains(path[i+1:], ".") {
		base := strings.ReplaceAll(path[i+1:], ".", "_")
		dir := path[:i]
		return dir + string(filepath.Separator) + base
	}
	return path
}

func ToVarName(paths []string) string {
	ps := strings.Join(paths, "__")
	input := []byte(url.PathEscape(ps))
	out := make([]byte, 0, len(input))
	for i := 0; i < len(input); i++ {
		c := input[i]
		if c == ':' || c == '*' {
			continue
		}
		if (c >= '0' && c <= '9' && i != 0) || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c == '_') {
			out = append(out, c)
		} else {
			out = append(out, '_')
		}
	}

	return string(out)
}

func SplitGoTags(input string) []string {
	out := make([]string, 0, 4)
	ns := len(input)

	flag := false
	prev := 0
	i := 0
	for i = 0; i < ns; i++ {
		c := input[i]
		if c == '"' {
			flag = !flag
		}
		if !flag && c == ' ' {
			if prev < i {
				out = append(out, input[prev:i])
			}
			prev = i + 1
		}
	}
	if i != 0 && prev < i {
		out = append(out, input[prev:i])
	}

	return out
}

func SubPackage(mod, dir string) string {
	if dir == "" {
		return mod
	}
	return mod + "/" + PathToImport(dir, "")
}

func SubDir(root, subPkg string) string {
	if root == "" {
		return ImportToPath(subPkg, "")
	}
	return filepath.Join(root, ImportToPath(subPkg, ""))
}

var (
	uniquePackageName        = map[string]bool{}
	uniqueMiddlewareName     = map[string]bool{}
	uniqueHandlerPackageName = map[string]bool{}
)

// GetPackageUniqueName can get a non-repeating variable name for package alias
func GetPackageUniqueName(name string) (string, error) {
	name, err := getUniqueName(name, uniquePackageName)
	if err != nil {
		return "", fmt.Errorf("can not generate unique name for package '%s', err: %v", name, err)
	}

	return name, nil
}

// GetMiddlewareUniqueName can get a non-repeating variable name for middleware name
func GetMiddlewareUniqueName(name string) (string, error) {
	name, err := getUniqueName(name, uniqueMiddlewareName)
	if err != nil {
		return "", fmt.Errorf("can not generate routing group for path '%s', err: %v", name, err)
	}

	return name, nil
}

func GetHandlerPackageUniqueName(name string) (string, error) {
	name, err := getUniqueName(name, uniqueHandlerPackageName)
	if err != nil {
		return "", fmt.Errorf("can not generate unique handler package name: '%s', err: %v", name, err)
	}

	return name, nil
}

// getUniqueName can get a non-repeating variable name
func getUniqueName(name string, uniqueNameSet map[string]bool) (string, error) {
	uniqueName := name
	if _, exist := uniqueNameSet[uniqueName]; exist {
		for i := 0; i < 10000; i++ {
			uniqueName = uniqueName + fmt.Sprintf("%d", i)
			if _, exist := uniqueNameSet[uniqueName]; !exist {
				break
			}
			uniqueName = name
			if i == 9999 {
				return "", fmt.Errorf("there is too many same package for %s", name)
			}
		}
	}
	uniqueNameSet[uniqueName] = true

	return uniqueName, nil
}

func SubPackageDir(path string) string {
	index := strings.LastIndex(path, "/")
	if index == -1 {
		return ""
	}
	return path[:index]
}

var validFuncReg = regexp.MustCompile("[_0-9a-zA-Z]")

// ToGoFuncName converts a string to a function naming style for go
func ToGoFuncName(s string) string {
	ss := []byte(s)
	for i := range ss {
		if !validFuncReg.Match([]byte{s[i]}) {
			ss[i] = '_'
		}
	}
	return string(ss)
}
