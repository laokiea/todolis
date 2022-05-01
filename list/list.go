package list

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fatih/color"
)

const TodolistFile = "todo"

type List interface {
	List() []*ListItem
	Add(string)
	Done(int)
	Del(int)
	Len()
	Search(string) []ListItem
}

type ListMap struct {
	m       []*ListItem
	len     int
	p       *regexp.Regexp
	lastKey string
	l       sync.RWMutex
}

type ListItem struct {
	v        string
	done     bool
	addDate  string
	doneDate string
}

func (i ListItem) String() string {
	return i.v
}

func (i ListItem) Status() string {
	if i.done {
		return color.GreenString("done")
	}
	return color.RedString("undone")
}

func (i ListItem) Display() string {
	return fmt.Sprintf("[%s][%s] %s", i.Status(), color.BlueString(i.addDate), color.New(color.FgHiBlack).Add(color.Bold).Add(color.BgWhite).Sprint(i.String()))
}

func (l *ListMap) Len() int {
	return len(l.m)
}

func (l *ListMap) IsEmpty() bool {
	return l.len == 0
}

func (l *ListMap) List() []*ListItem {
	l.l.RLock()
	defer l.l.RUnlock()
	return l.m
}

func (l *ListMap) ListSliceAll() []string {
	var s []string
	l.l.RLock()
	defer l.l.RUnlock()
	for _, i := range l.m {
		s = append(s, i.Display())
	}
	return s
}

func (l *ListMap) ListSliceUndone() []string {
	var s []string
	l.l.RLock()
	defer l.l.RUnlock()
	for _, i := range l.m {
		if !i.done {
			s = append(s, i.Display())
		}
	}
	return s
}

func (l *ListMap) Add(v string) {
	l.AddWith(v, false, time.Now().Format("2006/01/02"))
}

func (l *ListMap) AddWith(v string, done bool, date string) int {
	l.l.Lock()
	defer l.l.Unlock()
	l.len++
	l.m = append(l.m, &ListItem{v: v, done: done, addDate: date})
	return l.len - 1
}

func (l *ListMap) Done(i int) {
	l.l.Lock()
	defer l.l.Unlock()
	l.m[i].done = true
	l.m[i].doneDate = time.Now().Format("2006/01/02")
}

func (l *ListMap) Del(i int) {
	l.l.Lock()
	defer l.l.Unlock()
	l.len--
	l.m = append(l.m[:i], l.m[i+1:]...)
}

func (l *ListMap) Search(k string) []*ListItem {
	var lists []*ListItem
	if k != l.lastKey {
		l.lastKey = k
		l.p = regexp.MustCompile(`^.*` + k + `.*$`)
	}
	l.l.RLock()
	defer l.l.RUnlock()
	for _, i := range l.m {
		if l.p.MatchString(i.v) {
			lists = append(lists, i)
		}
	}
	return lists
}

// 把todolist内容写到文件里保存, 看main.go文件
func (l *ListMap) Flush() error {
	l.l.RLock()
	defer l.l.RUnlock()
	syscall.Unlink(TodolistFile)
	f, err := os.OpenFile(TodolistFile, os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		return err
	}
	var done string
	for _, i := range l.m {
		if i.done {
			done = "1"
		} else {
			done = "0"
		}
		_, err = f.Write([]byte(fmt.Sprintf("%s|%s|%s\n", i.v, done, i.addDate)))
		if err != nil {
			return err
		}
	}
	return nil
}

// 这个是在程序开始执行前
// 把文件中保存的list加载进内存中,看root.go文件(80-PreRunRoot方法)
func (l *ListMap) Load() error {
	f, err := os.OpenFile(TodolistFile, os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		item := strings.Split(line, "|")
		done := false
		if item[1] == "1" {
			done = true
		}
		l.AddWith(item[0], done, item[2])
	}
	return nil
}

var (
	GlobalLists ListMap
)
